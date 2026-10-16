use anyhow::Context;
use embedded_svc::mqtt::client::QoS;
use embedded_svc::wifi::{AuthMethod, ClientConfiguration, Configuration};
use esp_idf_hal::gpio::*;
use esp_idf_hal::peripherals::Peripherals;
use esp_idf_hal::reset::restart;
use esp_idf_svc::eventloop::EspSystemEventLoop;
use esp_idf_svc::mqtt::client::{EventPayload, EspMqttClient, MqttClientConfiguration};
use esp_idf_svc::nvs::EspDefaultNvsPartition;
use esp_idf_svc::wifi::EspWifi;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

const WIFI_SSID: &str = "Get schwifty";
const WIFI_PSK: &str = "P@$$w0rd1";
// const WIFI_SSID: &str = "Nothing to see here";
// const WIFI_PSK: &str = "P@$$w0rd1";
// const WIFI_SSID: &str = "freewifi";
// const WIFI_PSK: &str = "n1mr0d3l";

const MQTT_URI: &str = "mqtt://192.168.137.251:1883";

const ON: &str = "1";
const OFF: &str = "0";

const GET_TOPIC_4: &str = "home/outside/sprinklers/bank/4/state/get";
const SET_TOPIC_4: &str = "home/outside/sprinklers/bank/4/state/set";
const GET_TOPIC_3: &str = "home/outside/sprinklers/bank/3/state/get";
const SET_TOPIC_3: &str = "home/outside/sprinklers/bank/3/state/set";
const GET_TOPIC_2: &str = "home/outside/sprinklers/bank/2/state/get";
const SET_TOPIC_2: &str = "home/outside/sprinklers/bank/2/state/set";

/// Every topic we must (re-)subscribe to after every MQTT connection.
const SUBSCRIBE_TOPICS: &[&str] = &[
    SET_TOPIC_4, GET_TOPIC_4, SET_TOPIC_3, GET_TOPIC_3, SET_TOPIC_2, GET_TOPIC_2,
];

/// How often we re-ask esp_wifi to (re)connect while the link is down.
const WIFI_RECONNECT_INTERVAL: Duration = Duration::from_secs(5);

/// Last-resort escape hatch: if wifi has been down this long, reboot.
const WIFI_REBOOT_AFTER: Duration = Duration::from_secs(10 * 60);

/// How often the main loop ticks (also the max latency for a queued publish).
const MAIN_LOOP_TICK: Duration = Duration::from_millis(1000);

/// How often we print a status line to the log.
const HEARTBEAT_INTERVAL: Duration = Duration::from_secs(60);

#[derive(Debug, Clone)]
struct OutgoingMessage {
    topic: String,
    data: Vec<u8>,
}

fn apply_relay(pin: &mut PinDriver<'_, Output>, value: &str) -> bool {
    match value {
        ON => pin.set_high().is_ok(),
        OFF => pin.set_low().is_ok(),
        _ => false,
    }
}

fn main() -> anyhow::Result<()> {
    // It is necessary to call this function once. Otherwise some patches to the runtime
    // implemented by esp-idf-sys might not link properly. See https://github.com/esp-rs/esp-idf-template/issues/71
    esp_idf_svc::sys::link_patches();

    // Bind the log crate to the ESP Logging facilities
    esp_idf_svc::log::EspLogger::initialize_default();

    let peripherals = Peripherals::take().context("failed Peripherals::take()")?;
    let sys_loop = EspSystemEventLoop::take().context("failed EspSystemEventLoop::take()")?;
    let nvs = EspDefaultNvsPartition::take().context("failed EspDefaultNvsPartition::take()")?;

    let relay_4 = PinDriver::output(peripherals.pins.gpio27)?;
    let relay_3 = PinDriver::output(peripherals.pins.gpio14)?;
    let relay_2 = PinDriver::output(peripherals.pins.gpio12)?;

    let mut wifi_driver =
        EspWifi::new(peripherals.modem, sys_loop, Some(nvs)).context("failed EspWifi::new()")?;

    wifi_driver
        .set_configuration(&Configuration::Client(ClientConfiguration {
            auth_method: AuthMethod::WPA2Personal,
            ssid: WIFI_SSID.try_into().unwrap(),
            password: WIFI_PSK.try_into().unwrap(),
            ..Default::default()
        }))
        .context("failed wifi_driver.set_configuration()")?;

    let wifi_config = wifi_driver
        .get_configuration()
        .context("failed wifi_driver.get_configuration()")?;
    log::info!("wifi_config={:?}", wifi_config);

    wifi_driver
        .start()
        .context("failed wifi_driver.start()")?;

    // Kick off the first connection; the main loop keeps asking for it until it sticks.
    match wifi_driver.connect() {
        Ok(()) => log::info!("wifi connect() requested"),
        Err(e) => log::warn!("initial wifi connect() failed: {:?} (will retry)", e),
    }

    let mac = wifi_driver
        .sta_netif()
        .get_mac()
        .context("failed wifi_driver.sta_netif().get_mac()")?;
    log::info!("mac={:?}", mac.to_ascii_lowercase());

    let client_id = format!("mqtt-things-esp32-sprinklers-{:?}", mac.to_ascii_lowercase());

    let conf = MqttClientConfiguration {
        client_id: Some(&client_id),
        keep_alive_interval: Some(Duration::from_secs(30)),
        // esp-mqtt only enables auto-reconnect when this is Some(..);
        // this is the delay it waits before each reconnect attempt.
        reconnect_timeout: Some(Duration::from_secs(5)),
        ..Default::default()
    };

    let (mut client, mut connection) =
        EspMqttClient::new(MQTT_URI, &conf).context("failed EspMqttClient::new()")?;

    let (outgoing_message_sender, outgoing_message_receiver) =
        std::sync::mpsc::sync_channel::<OutgoingMessage>(1024);

    // Set by the MQTT handler thread whenever we (re)connect and must re-subscribe.
    let need_resubscribe = Arc::new(Mutex::new(true));
    // Mirrors the esp-mqtt connection state so we only subscribe when it's usable.
    let mqtt_connected = Arc::new(Mutex::new(false));

    let handler_need_resubscribe = Arc::clone(&need_resubscribe);
    let handler_mqtt_connected = Arc::clone(&mqtt_connected);

    let handler = std::thread::Builder::new()
        .stack_size(65536)
        .spawn(move || -> () {
            let mut relay_4 = relay_4;
            let mut relay_3 = relay_3;
            let mut relay_2 = relay_2;

            loop {
                match connection.next() {
                    Err(e) => log::error!("mqtt connection error: {:?}", e),
                    Ok(msg) => match msg.payload() {
                        EventPayload::BeforeConnect => log::info!("mqtt BeforeConnect"),
                        EventPayload::Connected(_) => {
                            log::info!("mqtt Connected");
                            *handler_mqtt_connected.lock().unwrap() = true;
                            // Subscriptions are NOT restored by esp-mqtt (clean session),
                            // so flag the main loop to re-subscribe everything.
                            *handler_need_resubscribe.lock().unwrap() = true;
                        }
                        EventPayload::Disconnected => {
                            log::info!("mqtt Disconnected (esp-mqtt will auto-reconnect)");
                            *handler_mqtt_connected.lock().unwrap() = false;
                        }
                        EventPayload::Subscribed(_) => log::info!("mqtt Subscribed"),
                        EventPayload::Unsubscribed(_) => log::info!("mqtt Unsubscribed"),
                        EventPayload::Published(_) => log::info!("mqtt Published"),
                        EventPayload::Deleted(_) => log::info!("mqtt Deleted"),
                        EventPayload::Error(e) => log::info!("mqtt Error: {:?}", e),
                        EventPayload::Received {
                            id,
                            topic,
                            data,
                            details,
                        } => {
                            log::info!(
                                "mqtt Received; id={:?}, topic={:?}, data={:?}, details={:?}",
                                id,
                                topic,
                                data,
                                details
                            );

                            let Some(topic) = topic else {
                                log::error!("skipping mqtt message with missing topic");
                                continue;
                            };
                            let topic = topic.to_string();

                            if data.is_empty() {
                                log::error!("skipping mqtt message with empty data");
                                continue;
                            }

                            let Ok(as_utf8) = String::from_utf8(data.to_vec()) else {
                                log::error!("failed to convert {:?} to utf8", data);
                                continue;
                            };
                            let as_utf8 = as_utf8.to_uppercase();

                            // GET topics are polled by the broker; we only act on SET.
                            if topic == GET_TOPIC_4 || topic == GET_TOPIC_3 || topic == GET_TOPIC_2 {
                                continue;
                            }

                            let (relay, get_topic) = match topic.as_str() {
                                SET_TOPIC_4 => (&mut relay_4, GET_TOPIC_4),
                                SET_TOPIC_3 => (&mut relay_3, GET_TOPIC_3),
                                SET_TOPIC_2 => (&mut relay_2, GET_TOPIC_2),
                                _ => continue,
                            };

                            if !apply_relay(relay, &as_utf8) {
                                log::error!(
                                    "failed to handle state for {:?}: {:?}",
                                    topic,
                                    as_utf8
                                );
                                continue;
                            }

                            log::info!("handled state for {:?}: {:?}", topic, as_utf8);

                            let outgoing_message = OutgoingMessage {
                                topic: get_topic.to_string(),
                                data: as_utf8.into_bytes(),
                            };

                            log::info!("requesting publish of {:?}", outgoing_message);
                            if outgoing_message_sender.send(outgoing_message).is_err() {
                                log::error!("failed to queue publish for set topic");
                            }
                        }
                    },
                }
            }
        })
        .context("failed std::thread::Builder::new()")?;

    let mut last_wifi_connect = Instant::now();
    let mut wifi_down_since: Option<Instant> = None;
    let mut last_heartbeat = Instant::now();

    loop {
        if handler.is_finished() {
            log::error!("mqtt handler thread exited; will reboot");
            restart();
        }

        //
        // wifi watchdog: esp_wifi does NOT auto-reconnect, so re-ask for it
        // every WIFI_RECONNECT_INTERVAL while the link is down.
        //
        let wifi_up = wifi_driver.is_up().unwrap_or(false);

        if !wifi_up {
            let down_since = *(wifi_down_since.get_or_insert_with(Instant::now));

            if last_wifi_connect.elapsed() >= WIFI_RECONNECT_INTERVAL {
                last_wifi_connect = Instant::now();
                match wifi_driver.connect() {
                    Ok(()) => log::info!("wifi down: connect() requested"),
                    Err(e) => log::warn!("wifi connect() failed: {:?}", e),
                }
            }

            if down_since.elapsed() > WIFI_REBOOT_AFTER {
                log::error!("wifi down for {:?}; will reboot", down_since);
                restart();
            }
        } else if wifi_down_since.take().is_some() {
            log::info!("wifi recovered");
            // esp-mqtt may be sitting on a dead socket; make sure it notices.
            last_wifi_connect = Instant::now();
        }

        //
        // mqtt: re-subscribe after every (re)connect. esp-mqtt reconnects the
        // socket itself, but subscriptions are lost (clean session), so we
        // must re-subscribe ourselves.
        //
        if *mqtt_connected.lock().unwrap() && *need_resubscribe.lock().unwrap() {
            let mut all_ok = true;

            for topic in SUBSCRIBE_TOPICS {
                match client.subscribe(topic, QoS::AtMostOnce) {
                    Ok(id) => log::info!("subscribed to {:?} (id {:?})", topic, id),
                    Err(e) => {
                        log::warn!("subscribe to {:?} failed: {:?}", topic, e);
                        all_ok = false;
                    }
                }
            }

            if all_ok {
                *need_resubscribe.lock().unwrap() = false;
            }
        }

        //
        // publish anything the handler queued; never propagate the error,
        // a publish failure while disconnected must not kill the main task.
        //
        match outgoing_message_receiver.recv_timeout(MAIN_LOOP_TICK) {
            Ok(outgoing_message) => {
                if client
                    .publish(
                        &outgoing_message.topic,
                        QoS::AtMostOnce,
                        true,
                        &outgoing_message.data,
                    )
                    .is_ok()
                {
                    log::info!("published {:?}", outgoing_message);
                } else {
                    log::warn!("publish failed (mqtt not connected?): {:?}", outgoing_message);
                }
            }
            Err(_) => {}
        }

        if last_heartbeat.elapsed() >= HEARTBEAT_INTERVAL {
            last_heartbeat = Instant::now();
            log::info!(
                "heartbeat: wifi_up={:?}, mqtt_connected={:?}, wifi_down_for={:?}",
                wifi_up,
                *mqtt_connected.lock().unwrap(),
                wifi_down_since.map(|since| since.elapsed())
            );
        }
    }
}

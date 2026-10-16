use anyhow::Context;
use embedded_svc::mqtt::client::QoS;
use embedded_svc::wifi::{AuthMethod, ClientConfiguration, Configuration};
use esp_idf_hal::gpio::*;
use esp_idf_hal::ledc::{config::TimerConfig, LedcDriver, LedcTimerDriver, Resolution};
use esp_idf_hal::peripherals::Peripherals;
use esp_idf_hal::units::FromValueType;
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

const PWM_FREQ_HZ: u32 = 100;

const ON: &str = "ON";
const OFF: &str = "OFF";
const MAX_VALUE: u32 = 65536;

const GET_SUFFIX: &str = "/get";
const SET_SUFFIX: &str = "/set";
const STATE_TOPIC_INFIX: &str = "/state";
const BRIGHTNESS_TOPIC_INFIX: &str = "/brightness";

const LED_1_PREFIX: &str = "home/outside/lights/led-string/1";
const LED_2_PREFIX: &str = "home/outside/lights/led-string/2";

const LED_1_STATE_SET_TOPIC: &str = "home/outside/lights/led-string/1/state/set";
const LED_1_STATE_GET_TOPIC: &str = "home/outside/lights/led-string/1/state/get";
const LED_1_BRIGHTNESS_SET_TOPIC: &str = "home/outside/lights/led-string/1/brightness/set";
const LED_1_BRIGHTNESS_GET_TOPIC: &str = "home/outside/lights/led-string/1/brightness/get";

const LED_2_STATE_SET_TOPIC: &str = "home/outside/lights/led-string/2/state/set";
const LED_2_STATE_GET_TOPIC: &str = "home/outside/lights/led-string/2/state/get";
const LED_2_BRIGHTNESS_SET_TOPIC: &str = "home/outside/lights/led-string/2/brightness/set";
const LED_2_BRIGHTNESS_GET_TOPIC: &str = "home/outside/lights/led-string/2/brightness/get";

/// Every topic we must (re-)subscribe to after every MQTT connection.
const SUBSCRIBE_TOPICS: &[&str] = &[
    LED_1_STATE_SET_TOPIC,
    LED_1_STATE_GET_TOPIC,
    LED_1_BRIGHTNESS_SET_TOPIC,
    LED_1_BRIGHTNESS_GET_TOPIC,
    LED_2_STATE_SET_TOPIC,
    LED_2_STATE_GET_TOPIC,
    LED_2_BRIGHTNESS_SET_TOPIC,
    LED_2_BRIGHTNESS_GET_TOPIC,
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

fn apply_state(driver: &mut LedcDriver<'_>, on: bool) -> bool {
    if on {
        driver.enable().is_ok()
    } else {
        driver.disable().is_ok()
    }
}

fn apply_brightness(driver: &mut LedcDriver<'_>, value: u32) -> Option<u32> {
    let max_duty = driver.get_max_duty(); // 65535 for 16-bit
    let duty =
        (((value as f32).clamp(0.0, MAX_VALUE as f32) / MAX_VALUE as f32) * max_duty as f32) as u32;

    if driver.set_duty(duty).is_ok() {
        Some(duty)
    } else {
        None
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

    let _led = PinDriver::output(peripherals.pins.gpio2).context("failed PinDriver::output()")?;

    let timer_config = TimerConfig {
        frequency: PWM_FREQ_HZ.Hz(),
        resolution: Resolution::Bits16,
        ..Default::default()
    };

    let timer_driver = std::sync::Arc::new(
        LedcTimerDriver::new(peripherals.ledc.timer0, &timer_config)
            .context("failed LedcTimerDriver::new()")?,
    );

    let led_driver_1 = LedcDriver::new(
        peripherals.ledc.channel0,
        std::sync::Arc::clone(&timer_driver),
        peripherals.pins.gpio4,
    )
    .context("failed LedcDriver::new()")?;

    let led_driver_2 = LedcDriver::new(
        peripherals.ledc.channel1,
        std::sync::Arc::clone(&timer_driver),
        peripherals.pins.gpio5,
    )
    .context("failed LedcDriver::new()")?;

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

    let conf = MqttClientConfiguration {
        client_id: Some("mqtt-things-esp32-dimmable-leds"),
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
            let mut led_driver_1 = led_driver_1;
            let mut led_driver_2 = led_driver_2;

            log::info!("max_duty: {:?}", led_driver_1.get_max_duty());

            // The first GET we see per channel+infix after a (re)connect is the
            // broker's retained value - apply it so the LEDs resume prior state.
            let mut led_1_reset_state_handled = false;
            let mut led_1_reset_brightness_handled = false;
            let mut led_2_reset_state_handled = false;
            let mut led_2_reset_brightness_handled = false;

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
                            // Allow the retained-state restore to run again after this outage.
                            led_1_reset_state_handled = false;
                            led_1_reset_brightness_handled = false;
                            led_2_reset_state_handled = false;
                            led_2_reset_brightness_handled = false;
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
                            let mut topic = topic.to_string();

                            if data.is_empty() {
                                log::error!("skipping mqtt message with empty data");
                                continue;
                            }

                            let Ok(as_utf8) = String::from_utf8(data.to_vec()) else {
                                log::error!("failed to convert {:?} to utf8", data);
                                continue;
                            };
                            let as_utf8 = as_utf8.to_uppercase();

                            if topic.contains(GET_SUFFIX) {
                                let mut skip = true;

                                if topic.starts_with(LED_1_PREFIX)
                                    && topic.contains(STATE_TOPIC_INFIX)
                                    && !led_1_reset_state_handled
                                {
                                    led_1_reset_state_handled = true;
                                    topic = topic.replace(GET_SUFFIX, SET_SUFFIX);
                                    log::info!("resetting led 1 from prior state");
                                    skip = false;
                                }

                                if topic.starts_with(LED_1_PREFIX)
                                    && topic.contains(BRIGHTNESS_TOPIC_INFIX)
                                    && !led_1_reset_brightness_handled
                                {
                                    led_1_reset_brightness_handled = true;
                                    topic = topic.replace(GET_SUFFIX, SET_SUFFIX);
                                    log::info!("resetting led 1 from prior brightness");
                                    skip = false;
                                }

                                if topic.starts_with(LED_2_PREFIX)
                                    && topic.contains(STATE_TOPIC_INFIX)
                                    && !led_2_reset_state_handled
                                {
                                    led_2_reset_state_handled = true;
                                    topic = topic.replace(GET_SUFFIX, SET_SUFFIX);
                                    log::info!("resetting led 2 from prior state");
                                    skip = false;
                                }

                                if topic.starts_with(LED_2_PREFIX)
                                    && topic.contains(BRIGHTNESS_TOPIC_INFIX)
                                    && !led_2_reset_brightness_handled
                                {
                                    led_2_reset_brightness_handled = true;
                                    topic = topic.replace(GET_SUFFIX, SET_SUFFIX);
                                    log::info!("resetting led 2 from prior brightness");
                                    skip = false;
                                }

                                if skip {
                                    continue;
                                }
                            }

                            log::info!(
                                "id: {:?}, topic: {:?}, data: {:?}, details: {:?}, as_utf8: {:?}",
                                id,
                                topic,
                                data,
                                details,
                                as_utf8
                            );

                            let driver = if topic.starts_with(LED_1_PREFIX) {
                                &mut led_driver_1
                            } else if topic.starts_with(LED_2_PREFIX) {
                                &mut led_driver_2
                            } else {
                                log::error!("failed to determine led driver for topic {:?}", topic);
                                continue;
                            };

                            if topic.contains(STATE_TOPIC_INFIX) {
                                if as_utf8 != ON && as_utf8 != OFF {
                                    log::error!(
                                        "failed to handle state for {:?}: {:?}",
                                        topic,
                                        as_utf8
                                    );
                                    continue;
                                }

                                if !apply_state(driver, as_utf8 == ON) {
                                    log::error!(
                                        "failed to handle state for {:?}: {:?}",
                                        topic,
                                        as_utf8
                                    );
                                    continue;
                                }

                                log::info!("handled state for {:?}: {:?}", topic, as_utf8);

                                let outgoing_message = OutgoingMessage {
                                    topic: topic.replace(SET_SUFFIX, GET_SUFFIX),
                                    data: as_utf8.clone().into_bytes(),
                                };

                                log::info!("requesting publish of {:?}", outgoing_message);
                                if outgoing_message_sender.send(outgoing_message).is_err() {
                                    log::error!("failed to queue state publish");
                                    continue;
                                }

                                if as_utf8 == OFF {
                                    let outgoing_message = OutgoingMessage {
                                        topic: topic
                                            .replace(SET_SUFFIX, GET_SUFFIX)
                                            .replace(STATE_TOPIC_INFIX, BRIGHTNESS_TOPIC_INFIX),
                                        data: "0".to_string().into_bytes(),
                                    };

                                    if outgoing_message_sender.send(outgoing_message).is_err() {
                                        log::error!("failed to queue brightness publish");
                                        continue;
                                    }
                                }
                            } else if topic.contains(BRIGHTNESS_TOPIC_INFIX) {
                                let Ok(value) = as_utf8.parse::<u32>() else {
                                    log::error!(
                                        "failed to parse brightness value for {:?}: {:?}",
                                        topic,
                                        as_utf8
                                    );
                                    continue;
                                };

                                let Some(duty) = apply_brightness(driver, value) else {
                                    log::error!(
                                        "failed to handle brightness set_duty for {:?}: {:?}",
                                        topic,
                                        as_utf8
                                    );
                                    continue;
                                };

                                log::info!("topic: {:?}, value: {:?}, duty: {:?}", topic, value, duty);

                                let outgoing_message = OutgoingMessage {
                                    topic: topic.replace(SET_SUFFIX, GET_SUFFIX),
                                    data: format!("{}", value).into_bytes(),
                                };

                                log::info!("requesting publish of {:?}", outgoing_message);
                                if outgoing_message_sender.send(outgoing_message).is_err() {
                                    log::error!("failed to queue brightness publish");
                                    continue;
                                }

                                if duty > 0 {
                                    let outgoing_message = OutgoingMessage {
                                        topic: topic
                                            .replace(SET_SUFFIX, GET_SUFFIX)
                                            .replace(BRIGHTNESS_TOPIC_INFIX, STATE_TOPIC_INFIX),
                                        data: ON.to_string().into_bytes(),
                                    };

                                    if outgoing_message_sender.send(outgoing_message).is_err() {
                                        log::error!("failed to queue state publish");
                                        continue;
                                    }
                                }
                            } else {
                                log::warn!("unrecognized topic suffix for {:?}", topic);
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

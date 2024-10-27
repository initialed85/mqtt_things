use anyhow::Context;
use embedded_svc::mqtt::client::QoS;
use embedded_svc::wifi::{AuthMethod, ClientConfiguration, Configuration};
use esp_idf_hal::gpio::*;
use esp_idf_hal::peripherals::Peripherals;
use esp_idf_hal::reset::restart;
use esp_idf_svc::mqtt::client::{EspMqttClient, MqttClientConfiguration};
use esp_idf_svc::{eventloop::EspSystemEventLoop, nvs::EspDefaultNvsPartition, wifi::EspWifi};
use std::sync::{Arc, Mutex};
use std::time::Duration;

const WIFI_SSID: &str = "Get schwifty";
const WIFI_PSK: &str = "P@$$w0rd1";
// const WIFI_SSID: &str = "Nothing to see here";
// const WIFI_PSK: &str = "P@$$w0rd1";
// const WIFI_SSID: &str = "freewifi";
// const WIFI_PSK: &str = "n1mr0d3l";

const MQTT_URI: &str = "mqtt://192.168.137.251:1883";

const ON: &str = "1";
const OFF: &str = "0";
const GET_TOPIC: &str = "home/outside/sprinklers/bank/4/state/get";
const SET_TOPIC: &str = "home/outside/sprinklers/bank/4/state/set";

#[derive(Debug, Clone)]
struct OutgoingMessage {
    topic: String,
    data: Vec<u8>,
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

    let mut relay = PinDriver::output(peripherals.pins.gpio4)?;

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

    wifi_driver.start().context("failed wifi_driver.start()")?;

    wifi_driver
        .connect()
        .context("failed wifi_driver.connect()")?;

    let expiry = std::time::SystemTime::now()
        .checked_add(std::time::Duration::from_secs(10))
        .context("failed time stuff")?;

    while !wifi_driver.is_up().context("failed wifi_driver.is_up()")? {
        let duration_since_expiry = std::time::SystemTime::now().duration_since(expiry);

        if !duration_since_expiry.is_err()
            && duration_since_expiry? > std::time::Duration::from_secs(0)
        {
            log::error!("timed out waiting for wifi_driver.is_up(); will reboot");
            restart();
        }

        log::info!("waiting for wifi_driver.is_up()...");
        std::thread::sleep(std::time::Duration::from_secs(1));
    }
    log::info!("connected.");

    let ip_info = wifi_driver
        .sta_netif()
        .get_ip_info()
        .context("failed wifi_driver.sta_netif().get_ip_info()")?;
    log::info!("ip_info={:?}", ip_info);

    let conf = MqttClientConfiguration {
        client_id: Some("mqtt-things-esp32-sprinklers"),
        keep_alive_interval: Some(Duration::from_secs(30)),
        ..Default::default()
    };

    let (mut client, mut connection) =
        EspMqttClient::new(MQTT_URI, &conf).context("failed EspMqttClient::new()")?;

    let (outgoing_message_sender, outgoing_message_receiver) =
        std::sync::mpsc::sync_channel::<OutgoingMessage>(1024);

    let ready_to_subscribe = Arc::new(Mutex::new(false));
    let thread_ready_to_subscribe = Arc::clone(&ready_to_subscribe);

    let handler = std::thread::Builder::new()
        .stack_size(65536)
        .spawn(move || -> anyhow::Result<()> {
            let mut reset_state_handled = false;

            loop {
                let reset_state_handled = &mut reset_state_handled;

                let outgoing_message_sender = &outgoing_message_sender;

                let msg = connection.next();

                match msg {
                    Err(e) => log::error!("error: {:?}", e),
                    Ok(msg) => match msg.payload() {
                        esp_idf_svc::mqtt::client::EventPayload::BeforeConnect => {
                            log::info!("mqtt BeforeConnect");
                        }
                        esp_idf_svc::mqtt::client::EventPayload::Connected(_) => {
                            log::info!("mqtt Connected");

                            let mut this_ready_to_subscribe =
                                thread_ready_to_subscribe.lock().unwrap();
                            *this_ready_to_subscribe = true;
                            drop(this_ready_to_subscribe);
                        }
                        esp_idf_svc::mqtt::client::EventPayload::Disconnected => {
                            log::info!("mqtt Disconnected");

                            log::info!("exiting handler...");
                            return Ok(());
                        }
                        esp_idf_svc::mqtt::client::EventPayload::Subscribed(_) => {
                            log::info!("mqtt Subscribed");
                        }
                        esp_idf_svc::mqtt::client::EventPayload::Unsubscribed(_) => {
                            log::info!("mqtt Unsubscribed");
                        }
                        esp_idf_svc::mqtt::client::EventPayload::Published(_) => {
                            log::info!("mqtt Published");
                        }
                        esp_idf_svc::mqtt::client::EventPayload::Deleted(_) => {
                            log::info!("mqtt Deleted");
                        }
                        esp_idf_svc::mqtt::client::EventPayload::Error(e) => {
                            log::info!("mqtt Error: {:?}", e);
                        }
                        esp_idf_svc::mqtt::client::EventPayload::Received {
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

                            if topic.is_none() {
                                log::error!("skipping mqtt message with missing topic");
                                continue;
                            }
                            let mut topic = topic.unwrap().to_string();

                            if data.is_empty() {
                                log::error!("skipping mqtt message with empty data");
                                continue;
                            }

                            let as_utf8 = String::from_utf8(data.to_vec());
                            if as_utf8.is_err() {
                                log::error!("failed to convert {:?} to utf8; {:?}", data, as_utf8);
                                continue;
                            }
                            let as_utf8 = as_utf8?.to_uppercase();

                            log::info!(
                                "id: {:?}, topic: {:?}, data: {:?}, details: {:?}, as_utf8: {:?}",
                                id,
                                topic,
                                data,
                                details,
                                as_utf8
                            );

                            if topic == GET_TOPIC {
                                let skip = true;

                                if !*reset_state_handled {
                                    *reset_state_handled = true;
                                    topic = SET_TOPIC.into();
                                }

                                if skip {
                                    continue;
                                }
                            }

                            if topic == SET_TOPIC {
                                if as_utf8 == ON {
                                    let result = relay.set_high();
                                    if result.is_err() {
                                        log::error!(
                                            "failed to handle state enable for {:?}: {:?}",
                                            topic,
                                            as_utf8
                                        );
                                        continue;
                                    }

                                    log::info!(
                                        "handled state enable for {:?}: {:?}",
                                        topic,
                                        as_utf8
                                    );
                                } else if as_utf8 == OFF {
                                    let result = relay.set_low();
                                    if result.is_err() {
                                        log::error!(
                                            "failed to handle state disable for {:?}: {:?}",
                                            topic,
                                            as_utf8
                                        );
                                        continue;
                                    }

                                    log::info!(
                                        "handled state enable for {:?}: {:?}",
                                        topic,
                                        as_utf8
                                    );
                                } else {
                                    log::error!(
                                        "failed to handle state enable for {:?}: {:?}",
                                        topic,
                                        as_utf8
                                    );
                                    continue;
                                }

                                let outgoing_message_topic = GET_TOPIC.into();
                                let outgoing_message_data = as_utf8.to_string().into_bytes();
                                let outgoing_message = OutgoingMessage {
                                    topic: outgoing_message_topic,
                                    data: outgoing_message_data,
                                };

                                log::info!("requesting publish of {:?}", outgoing_message);
                                let result = outgoing_message_sender.send(outgoing_message.clone());
                                if result.is_err() {
                                    log::error!(
                                        "failed to handle brightness publish to set topic; {:?}",
                                        outgoing_message,
                                    );
                                    continue;
                                }
                            }
                        }
                    },
                }
            }
        })
        .context("failed std::thread::Builder::new()")?;

    log::info!("waiting to be ready to subscribe...");
    loop {
        let this_ready_to_subscribe = ready_to_subscribe.try_lock().unwrap();
        if *this_ready_to_subscribe {
            drop(this_ready_to_subscribe);
            break;
        }

        std::thread::sleep(std::time::Duration::from_secs(1));
    }

    log::info!("subscribing to {:?}", SET_TOPIC);
    client
        .subscribe(SET_TOPIC, QoS::AtMostOnce)
        .context("failed client.subscribe()")?;

    log::info!("subscribing to {:?}", GET_TOPIC);
    client
        .subscribe(GET_TOPIC, QoS::AtMostOnce)
        .context("failed client.subscribe()")?;

    loop {
        let handler = &handler;
        if handler.is_finished() {
            log::error!("handler unexpectedly exited; will reboot");
            restart();
        }

        let outgoing_message =
            outgoing_message_receiver.recv_timeout(std::time::Duration::from_millis(5000));
        if outgoing_message.is_err() {
            continue;
        }
        let outgoing_message = outgoing_message?;

        client
            .publish(
                &outgoing_message.topic,
                QoS::AtMostOnce,
                true,
                &outgoing_message.data,
            )
            .context("failed client.publish")?;

        log::info!("published {:?}", outgoing_message);
    }
}

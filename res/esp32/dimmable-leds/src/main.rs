use std::time::Duration;

use anyhow::Context;

use embedded_svc::mqtt::client::QoS;
use embedded_svc::wifi::{AuthMethod, ClientConfiguration, Configuration};
use esp_idf_hal::gpio::*;
use esp_idf_hal::ledc::{config::TimerConfig, LedcDriver, LedcTimerDriver, Resolution};
use esp_idf_hal::peripherals::Peripherals;
use esp_idf_hal::prelude::*;
use esp_idf_svc::mqtt::client::{EspMqttClient, MqttClientConfiguration};
use esp_idf_svc::{eventloop::EspSystemEventLoop, nvs::EspDefaultNvsPartition, wifi::EspWifi};

const PWM_FREQ_HZ: u32 = 100;

const ON: &str = "ON";
const OFF: &str = "OFF";
const MAX_VALUE: u32 = 65536;

const GET_SUFFIX: &str = "/get";
const SET_SUFFIX: &str = "/set";

const LED_1_PREFIX: &str = "home/outside/lights/led-string/1";
const LED_2_PREFIX: &str = "home/outside/lights/led-string/2";
const STATE_TOPIC_INFIX: &str = "/state";
const BRIGHTNESS_TOPIC_INFIX: &str = "/brightness";

#[derive(Debug, Clone)]
struct OutgoingMessage {
    topic: String,
    data: Vec<u8>,
}

fn main() -> anyhow::Result<()> {
    let mut led_1_state_set_topic = LED_1_PREFIX.to_owned();
    led_1_state_set_topic.push_str(STATE_TOPIC_INFIX);
    led_1_state_set_topic.push_str(SET_SUFFIX);
    let led_1_state_set_topic = led_1_state_set_topic.as_str();

    let mut led_1_brightness_set_topic = LED_1_PREFIX.to_owned();
    led_1_brightness_set_topic.push_str(BRIGHTNESS_TOPIC_INFIX);
    led_1_brightness_set_topic.push_str(SET_SUFFIX);
    let led_1_brightness_set_topic = led_1_brightness_set_topic.as_str();

    let mut led_1_state_get_topic = LED_1_PREFIX.to_owned();
    led_1_state_get_topic.push_str(STATE_TOPIC_INFIX);
    led_1_state_get_topic.push_str(SET_SUFFIX);
    let led_1_state_get_topic = led_1_state_get_topic.as_str();

    let mut led_1_brightness_get_topic = LED_1_PREFIX.to_owned();
    led_1_brightness_get_topic.push_str(BRIGHTNESS_TOPIC_INFIX);
    led_1_brightness_get_topic.push_str(GET_SUFFIX);
    let led_1_brightness_get_topic = led_1_brightness_get_topic.as_str();

    let mut led_2_state_set_topic = LED_2_PREFIX.to_owned();
    led_2_state_set_topic.push_str(STATE_TOPIC_INFIX);
    led_2_state_set_topic.push_str(SET_SUFFIX);
    let led_2_state_set_topic = led_2_state_set_topic.as_str();

    let mut led_2_brightness_set_topic = LED_2_PREFIX.to_owned();
    led_2_brightness_set_topic.push_str(BRIGHTNESS_TOPIC_INFIX);
    led_2_brightness_set_topic.push_str(SET_SUFFIX);
    let led_2_brightness_set_topic = led_2_brightness_set_topic.as_str();

    let mut led_2_state_get_topic = LED_2_PREFIX.to_owned();
    led_2_state_get_topic.push_str(STATE_TOPIC_INFIX);
    led_2_state_get_topic.push_str(GET_SUFFIX);
    let led_2_state_get_topic = led_2_state_get_topic.as_str();

    let mut led_2_brightness_get_topic = LED_2_PREFIX.to_owned();
    led_2_brightness_get_topic.push_str(BRIGHTNESS_TOPIC_INFIX);
    led_2_brightness_get_topic.push_str(GET_SUFFIX);
    let led_2_brightness_get_topic = led_2_brightness_get_topic.as_str();

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

    let mut led_driver_1 = LedcDriver::new(
        peripherals.ledc.channel0,
        std::sync::Arc::clone(&timer_driver),
        peripherals.pins.gpio4,
    )
    .context("failed LedcDriver::new()")?;

    let mut led_driver_2 = LedcDriver::new(
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
            ssid: "Nothing to see here".try_into().unwrap(),
            password: "P@$$w0rd1".try_into().unwrap(),
            ..Default::default()
        }))
        .context("failed wifi_driver.set_configuration()")?;

    let wifi_config = wifi_driver
        .get_configuration()
        .context("failed wifi_driver.get_configuration()")?;
    println!("wifi_config={:?}", wifi_config);

    wifi_driver.start().context("failed wifi_driver.start()")?;

    wifi_driver
        .connect()
        .context("failed wifi_driver.connect()")?;

    while !wifi_driver.is_up().context("failed wifi_driver.is_up()")? {
        println!("waiting for wifi_driver.is_up()...");
        std::thread::sleep(std::time::Duration::from_millis(1000));
    }
    println!("connected.");

    let ip_info = wifi_driver
        .ap_netif()
        .get_ip_info()
        .context("failed wifi_driver.ap_netif().get_ip_info()")?;
    println!("ip_info={:?}", ip_info);

    let conf = MqttClientConfiguration {
        client_id: Some("mqtt-things-esp32-dimmable-leds"),
        keep_alive_interval: Some(Duration::from_secs(30)),
        ..Default::default()
    };

    let (mut client, mut connection) = EspMqttClient::new("mqtt://192.168.137.251:1883", &conf)
        .context("failed EspMqttClient::new()")?;

    let (outgoing_message_sender, outgoing_message_receiver) =
        std::sync::mpsc::sync_channel::<OutgoingMessage>(1024);

    std::thread::Builder::new()
        .stack_size(65536)
        .spawn(move || -> anyhow::Result<()> {
            let max_duty = led_driver_1.get_max_duty(); // 65535 for 16-bit
            log::info!("max_duty: {:?}", max_duty);

            let mut led_1_reset_state_handled = false;
            let mut led_1_reset_brightness_handled = false;

            let mut led_2_reset_state_handled = false;
            let mut led_2_reset_brightness_handled = false;

            loop {
                let led_driver_1 = &mut led_driver_1;
                let led_1_reset_state_handled = &mut led_1_reset_state_handled;
                let led_1_reset_brightness_handled = &mut led_1_reset_brightness_handled;

                let led_driver_2 = &mut led_driver_2;
                let led_2_reset_state_handled = &mut led_2_reset_state_handled;
                let led_2_reset_brightness_handled = &mut led_2_reset_brightness_handled;

                let outgoing_message_sender = &outgoing_message_sender;

                let msg = connection.next();

                match msg {
                    Err(e) => log::error!("error: {:?}", e),
                    Ok(msg) => match msg.payload() {
                        esp_idf_svc::mqtt::client::EventPayload::BeforeConnect => {}
                        esp_idf_svc::mqtt::client::EventPayload::Connected(_) => {
                            log::info!("mqtt connected");
                        }
                        esp_idf_svc::mqtt::client::EventPayload::Disconnected => {
                            log::info!("mqtt disconnected");
                        }
                        esp_idf_svc::mqtt::client::EventPayload::Subscribed(_) => {}
                        esp_idf_svc::mqtt::client::EventPayload::Unsubscribed(_) => {}
                        esp_idf_svc::mqtt::client::EventPayload::Published(_) => {}
                        esp_idf_svc::mqtt::client::EventPayload::Deleted(_) => {}
                        esp_idf_svc::mqtt::client::EventPayload::Error(e) => {
                            log::info!("mqtt error: {:?}", e);
                        }
                        esp_idf_svc::mqtt::client::EventPayload::Received {
                            id,
                            topic,
                            data,
                            details,
                        } => {
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

                            if topic.contains(GET_SUFFIX) {
                                let mut skip = true;

                                if topic.starts_with(LED_1_PREFIX)
                                    && topic.contains(STATE_TOPIC_INFIX)
                                    && !*led_1_reset_state_handled
                                {
                                    *led_1_reset_state_handled = true;
                                    topic = topic.replace(GET_SUFFIX, SET_SUFFIX);
                                    log::info!("resetting led 1 from prior state");
                                    skip = false;
                                }

                                if topic.starts_with(LED_1_PREFIX)
                                    && topic.contains(BRIGHTNESS_TOPIC_INFIX)
                                    && !*led_1_reset_brightness_handled
                                {
                                    *led_1_reset_brightness_handled = true;
                                    topic = topic.replace(GET_SUFFIX, SET_SUFFIX);
                                    log::info!("resetting led 1 from prior brightness");
                                    skip = false;
                                }

                                if topic.starts_with(LED_2_PREFIX)
                                    && topic.contains(STATE_TOPIC_INFIX)
                                    && !*led_2_reset_state_handled
                                {
                                    *led_2_reset_state_handled = true;
                                    topic = topic.replace(GET_SUFFIX, SET_SUFFIX);
                                    log::info!("resetting led 2 from prior state");
                                    skip = false;
                                }

                                if topic.starts_with(LED_2_PREFIX)
                                    && topic.contains(BRIGHTNESS_TOPIC_INFIX)
                                    && !*led_2_reset_brightness_handled
                                {
                                    *led_2_reset_brightness_handled = true;
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

                            let &driver;

                            if topic.starts_with(LED_1_PREFIX) {
                                driver = led_driver_1;
                            } else if topic.starts_with(LED_2_PREFIX) {
                                driver = led_driver_2;
                            } else {
                                log::error!("failed to determine led driver for topic {:?}", topic);
                                continue;
                            }

                            if topic.contains(STATE_TOPIC_INFIX) {
                                if as_utf8 == ON {
                                    let result = driver.enable();
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
                                    let result = driver.disable();
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

                                let outgoing_message_topic = topic.replace(SET_SUFFIX, GET_SUFFIX);
                                let outgoing_message_data = as_utf8.to_string().into_bytes();
                                let outgoing_message = OutgoingMessage {
                                    topic: outgoing_message_topic,
                                    data: outgoing_message_data,
                                };

                                let result = outgoing_message_sender.send(outgoing_message.clone());
                                if result.is_err() {
                                    log::error!(
                                        "failed to handle brightness publish to set topic; {:?}",
                                        outgoing_message,
                                    );
                                    continue;
                                }

                                log::info!("requesting publish of {:?}", outgoing_message);

                                if as_utf8 == OFF {
                                    let outgoing_message_topic = topic
                                        .replace(SET_SUFFIX, GET_SUFFIX)
                                        .replace(STATE_TOPIC_INFIX, BRIGHTNESS_TOPIC_INFIX);
                                    let outgoing_message_data = format!("{}", 0).into_bytes();
                                    let outgoing_message = OutgoingMessage {
                                        topic: outgoing_message_topic,
                                        data: outgoing_message_data,
                                    };

                                    let result =
                                        outgoing_message_sender.send(outgoing_message.clone());
                                    if result.is_err() {
                                        log::error!(
                                        "failed to handle brightness publish to set topic; {:?}",
                                        outgoing_message,
                                    );
                                        continue;
                                    }
                                }
                            } else if topic.contains(BRIGHTNESS_TOPIC_INFIX) {
                                let value = as_utf8.parse::<u32>();
                                if value.is_err() {
                                    log::error!(
                                        "failed to parse brightness value for {:?}: {:?}",
                                        topic,
                                        as_utf8
                                    );
                                    continue;
                                }
                                let value = value?;

                                let max_duty = driver.get_max_duty(); // 65535 for 16-bit
                                log::info!("max_duty: {:?}", max_duty);

                                let duty = (((value as f32).clamp(0.0, MAX_VALUE as f32)
                                    / MAX_VALUE as f32)
                                    * max_duty as f32)
                                    as u32;

                                log::info!(
                                    "topic: {:?}, value: {:?}, duty: {:?}",
                                    topic,
                                    value,
                                    duty,
                                );

                                let result = driver.set_duty(duty);
                                if result.is_err() {
                                    log::error!(
                                        "failed to handle brightness set_duty for {:?}: {:?}",
                                        topic,
                                        as_utf8
                                    );
                                    continue;
                                }

                                let outgoing_message_topic = topic.replace(SET_SUFFIX, GET_SUFFIX);
                                let outgoing_message_data = format!("{}", value).into_bytes();
                                let outgoing_message = OutgoingMessage {
                                    topic: outgoing_message_topic,
                                    data: outgoing_message_data,
                                };

                                let result = outgoing_message_sender.send(outgoing_message.clone());
                                if result.is_err() {
                                    log::error!(
                                        "failed to handle brightness publish to set topic; {:?}",
                                        outgoing_message,
                                    );
                                    continue;
                                }

                                log::info!("requesting publish of {:?}", outgoing_message);

                                if duty > 0 {
                                    let outgoing_message_topic = topic
                                        .replace(SET_SUFFIX, GET_SUFFIX)
                                        .replace(BRIGHTNESS_TOPIC_INFIX, STATE_TOPIC_INFIX);
                                    let outgoing_message_data = ON.to_string().into_bytes();
                                    let outgoing_message = OutgoingMessage {
                                        topic: outgoing_message_topic,
                                        data: outgoing_message_data,
                                    };

                                    let result =
                                        outgoing_message_sender.send(outgoing_message.clone());
                                    if result.is_err() {
                                        log::error!(
                                        "failed to handle brightness publish to set topic; {:?}",
                                        outgoing_message,
                                    );
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

    //
    // set topics
    //

    log::info!("subscribing to {:?}", led_1_state_set_topic);
    client
        .subscribe(led_1_state_set_topic, QoS::AtMostOnce)
        .context("failed client.subscribe()")?;

    log::info!("subscribing to {:?}", led_1_brightness_set_topic);
    client
        .subscribe(led_1_brightness_set_topic, QoS::AtMostOnce)
        .context("failed client.subscribe()")?;

    log::info!("subscribing to {:?}", led_2_state_set_topic);
    client
        .subscribe(led_2_state_set_topic, QoS::AtMostOnce)
        .context("failed client.subscribe()")?;

    log::info!("subscribing to {:?}", led_2_brightness_set_topic);
    client
        .subscribe(led_2_brightness_set_topic, QoS::AtMostOnce)
        .context("failed client.subscribe()")?;

    //
    // get topics
    //

    log::info!("subscribing to {:?}", led_1_state_get_topic);
    client
        .subscribe(led_1_state_get_topic, QoS::AtMostOnce)
        .context("failed client.subscribe()")?;

    log::info!("subscribing to {:?}", led_1_brightness_get_topic);
    client
        .subscribe(led_1_brightness_get_topic, QoS::AtMostOnce)
        .context("failed client.subscribe()")?;

    log::info!("subscribing to {:?}", led_2_state_get_topic);
    client
        .subscribe(led_2_state_get_topic, QoS::AtMostOnce)
        .context("failed client.subscribe()")?;

    log::info!("subscribing to {:?}", led_2_brightness_get_topic);
    client
        .subscribe(led_2_brightness_get_topic, QoS::AtMostOnce)
        .context("failed client.subscribe()")?;

    loop {
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

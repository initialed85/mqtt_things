use std::time::Duration;

use anyhow::Context;

use embedded_svc::mqtt::client::{Connection, Publish, QoS};
use embedded_svc::wifi::{AuthMethod, ClientConfiguration, Configuration};
use esp_idf_hal::gpio::*;
use esp_idf_hal::ledc::{config::TimerConfig, LedcDriver, LedcTimerDriver, Resolution};
use esp_idf_hal::peripherals::Peripherals;
use esp_idf_hal::prelude::*;
use esp_idf_svc::mqtt::client::{EspMqttClient, MqttClientConfiguration};
use esp_idf_svc::{eventloop::EspSystemEventLoop, nvs::EspDefaultNvsPartition, wifi::EspWifi};

const PWM_FREQ_HZ: u32 = 100;

fn main() -> anyhow::Result<()> {
    // It is necessary to call this function once. Otherwise some patches to the runtime
    // implemented by esp-idf-sys might not link properly. See https://github.com/esp-rs/esp-idf-template/issues/71
    esp_idf_svc::sys::link_patches();

    // Bind the log crate to the ESP Logging facilities
    esp_idf_svc::log::EspLogger::initialize_default();

    let peripherals = Peripherals::take().context("failed Peripherals::take()")?;
    let sys_loop = EspSystemEventLoop::take().context("failed EspSystemEventLoop::take()")?;
    let nvs = EspDefaultNvsPartition::take().context("failed EspDefaultNvsPartition::take()")?;

    let _led = PinDriver::output(peripherals.pins.gpio2)?;

    let timer_config = TimerConfig {
        frequency: PWM_FREQ_HZ.Hz(),
        resolution: Resolution::Bits16,
        ..Default::default()
    };

    let timer_driver = std::sync::Arc::new(LedcTimerDriver::new(
        peripherals.ledc.timer0,
        &timer_config,
    )?);

    let mut led_driver_1 = LedcDriver::new(
        peripherals.ledc.channel0,
        std::sync::Arc::clone(&timer_driver),
        peripherals.pins.gpio4,
    )?;

    let mut led_driver_2 = LedcDriver::new(
        peripherals.ledc.channel1,
        std::sync::Arc::clone(&timer_driver),
        peripherals.pins.gpio5,
    )?;

    let mut wifi_driver = EspWifi::new(peripherals.modem, sys_loop, Some(nvs))?;

    wifi_driver.set_configuration(&Configuration::Client(ClientConfiguration {
        auth_method: AuthMethod::WPA2Personal,
        ssid: "Nothing to see here".try_into().unwrap(),
        password: "P@$$w0rd1".try_into().unwrap(),
        ..Default::default()
    }))?;

    let wifi_config = wifi_driver.get_configuration()?;
    println!("wifi_config={:?}", wifi_config);

    wifi_driver.start()?;

    wifi_driver.connect()?;

    while !wifi_driver.is_up()? {
        println!("waiting for wifi_driver.is_up()...");
        std::thread::sleep(std::time::Duration::from_millis(1000));
    }
    println!("connected.");

    let ip_info = wifi_driver.ap_netif().get_ip_info()?;
    println!("ip_info={:?}", ip_info);

    let conf = MqttClientConfiguration {
        client_id: Some("mqtt-things-esp32-dimmable-leds"),
        keep_alive_interval: Some(Duration::from_secs(30)),
        ..Default::default()
    };

    let (actual_client, mut connection) = EspMqttClient::new("mqtt://192.168.137.251:1883", &conf)?;

    let outer_client = std::sync::Arc::new(std::sync::Mutex::new(actual_client));
    let inner_client = std::sync::Arc::clone(&outer_client);

    std::thread::Builder::new()
        .stack_size(32768)
        .spawn(move || -> anyhow::Result<()> {
            let max_duty = led_driver_1.get_max_duty(); // 65535 for 16-bit
            log::info!("max_duty: {:?}", max_duty);

            loop {
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
                            log::info!(
                                "id: {:?}, topic: {:?}, data: {:?}, details: {:?}",
                                id,
                                topic,
                                data,
                                details
                            );

                            let as_utf8 = String::from_utf8(data.to_vec());
                            if as_utf8.is_err() {
                                log::info!("failed to convert {:?} to utf8; {:?}", data, as_utf8);
                                continue;
                            }
                            let as_utf8 = as_utf8?;

                            let value = as_utf8.parse::<u32>();
                            if value.is_err() {
                                log::info!("failed to convert {:?} to u32; {:?}", as_utf8, value);
                                continue;
                            }
                            let value = value?;

                            let duty = value.clamp(0, max_duty);

                            log::info!("value: {:?}, duty: {:?}", value, duty);

                            let raw_value_as_str = format!("{:?}", value.clone());
                            let raw_value = raw_value_as_str.as_bytes();

                            {
                                let mut client = inner_client.lock().unwrap();

                                match topic {
                                    Some("home/outside/lights/led-string/1/state/set") => {
                                        if led_driver_1.set_duty(duty).is_ok() {
                                            // TODO: hangs? probably need a channel + keep it in main thread
                                            // let result = client.publish(
                                            //     "home/outside/lights/led-string/1/state/get",
                                            //     QoS::AtMostOnce,
                                            //     true,
                                            //     raw_value,
                                            // );
                                            // if result.is_err() {
                                            //     log::error!(
                                            //         "failed to publish to mqtt: {:?}",
                                            //         result.unwrap_err()
                                            //     );
                                            // }
                                        }
                                    }
                                    Some("home/outside/lights/led-string/2/state/set") => {
                                        if led_driver_2.set_duty(duty).is_ok() {
                                            // TODO: hangs? probably need a channel + keep it in main thread
                                            // let result = client.publish(
                                            //     "home/outside/lights/led-string/2/state/get",
                                            //     QoS::AtMostOnce,
                                            //     true,
                                            //     raw_value,
                                            // );
                                            // if result.is_err() {
                                            //     log::error!(
                                            //         "failed to publish to mqtt: {:?}",
                                            //         result.unwrap_err()
                                            //     );
                                            // }
                                        }
                                    }
                                    None => {}
                                    _ => {}
                                }

                                drop(client);
                            }
                        }
                    },
                }
            }
        })?;

    {
        let mut client = outer_client.lock().unwrap();

        client.subscribe(
            "home/outside/lights/led-string/1/state/set",
            QoS::AtMostOnce,
        )?;

        client.subscribe(
            "home/outside/lights/led-string/2/state/set",
            QoS::AtMostOnce,
        )?;

        drop(client);
    }

    loop {
        std::thread::sleep(std::time::Duration::from_millis(1000));
    }
}

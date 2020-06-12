# mqtt_things

This repo contains some Go code to expose a few different devices via MQTT and a generic "action" concept over the top of that.

## What does it do?

I'll speak to the top level commands:

- `aircon_cli`
    - Listens for 0 or 1 at `home/inside/aircon/state/set`
    - Sends packets to a [zmote](https://www.zmote.io) which sends IR blasts to Fujitsu aircon for two states
        - 1 = On at 18 deg C and low fan
        - 0 = Off
    - Writes to `home/inside/aircon/state/get`
- `heater_cli`
    - Listens for 0 or 1 at `home/inside/heater/state/set`
    - Sends serial data to an Arduino (see `res/arduino`) for the specified relay number
    - Writes to `home/inside/heater/state/get`
- `http_cli`
    - Exposes all MQTT topics ending in `/get` as JSON via HTTP endpoints with some name simplification
    - e.g.
        - `home/inside/aircon/state/get` becomes `/aircon`; returning a `last_updated` and a `payload` field as JSON 
- `lights_cli`
    - Listens for 0 or 1 at `/home/inside/lights/globe/(name)/set`
    - Interacts with a Philips Hue bridge to turn on or off the specified light
        - NOTE: there's some implicit logic that tweaks lights name (lowercased, spaces become dashes) to make clean looking URLs
    - Writes to `/home/inside/lights/globe/(name)/get`
- `mqtt_cli`
    - Uses the MQTT client used by everything else
    - Enable debugging via publish/subscribe
        - e.g.: `mqtt_cli -host 192.168.137.253 -topic \#`
- `sprinklers_cli`
    - Listens for 0 or 1 at `/home/outside/sprinklers/bank/(number)/set`
    - Sends serial data to an Arduino (see `res/arduino`) for the specified relay numbers
    - Writes to `/home/outside/sprinklers/bank/(number)/get`
- `switches_cli`
    - Interacts with a Tasmota-flashed Kogan/Powertech (Jaycar) WiFi smart plug
    - Listens for a 0 or 1 at `home/inside/switches/globe/(name)/state/set`
    - Interacts with the Tasmota device via HTTP
    - Writes to `home/inside/switches/globe/(name)/state/get` 
- `weather_cli`
    - Interacts with OpenWeatherAPI with a specific API key for a specific location
    - Writes to the following topics
        - `home/outside/weather/temperature/get`
        - `home/outside/weather/pressure/get`
        - `home/outside/weather/humidity/get`
        - `home/outside/weather/wind-speed/get`
        - `home/outside/weather/wind-direction/get`
        - `home/outside/weather/sunrise/get`
        - `home/outside/weather/sunset/get`
- `circumstances_cli`
    - given bed time, wake time and temperature brackets as arguments
    - Listens to the following topics from `weather_cli`
        - `home/outside/weather/temperature/get`
        - `home/outside/weather/sunrise/get`
        - `home/outside/weather/sunset/get`
    - Determines some circumstances based on that data and the current time
    - Writes to the following topics (assuming prefix of `home/circumstances`)
        - `home/circumstances/after_sunrise/get`
        - `home/circumstances/after_sunset/get`
        - `home/circumstances/after_bedtime/get`
        - `home/circumstances/after_waketime/get`
        - `home/circumstances/after_sunrise_15m_early/get`
        - `home/circumstances/after_sunset_15m_early/get`
        - `home/circumstances/after_bedtime_15m_early/get`
        - `home/circumstances/after_waketime_15m_early/get`
        - `home/circumstances/after_sunrise_15m_late/get`
        - `home/circumstances/after_sunset_15m_late/get`
        - `home/circumstances/after_bedtime_15m_late/get`
        - `home/circumstances/after_waketime_15m_late/get`
        - `home/circumstances/hot/get`
        - `home/circumstances/comfortable/get`
        - `home/circumstances/cold/get`
- `arp_cil`
    - Given some IP addresses as arguments
    - ARPs for those IP addresses with a 10 second timeout
    - Writes to `home/arp/(IP address)`
    - NOTE: only works on your local network (because ARP is layer 2); can be used to determine if a host is online or offline (e.g. your phone)

## How do I build it?

    ./build.sh
    
If (like me) you wanna deploy this to a Raspberry Pi, you'll need to change your command line to the following:

    GOOS=linux GOARCH=arm ./build.sh

## How do I test it?

    ./test.sh

## How do I run it?

Here's a dump of all of the command lines as I'm using them around my house (with some data sanitised):

    aircon_cli -host 192.168.137.253 -aircon 192.168.137.20
    heater_cli -host 192.168.137.253 -port /dev/ttyACM0 -relay 1
    http_cli -host 192.168.137.253 -port 8079
    lights_cli -host 192.168.137.253 -apiKey (API key) -bridgeHost 192.168.137.252
    sprinklers_cli -host 192.168.137.253 -port /dev/ttyACM0 -relay 2 -relay 3
    switches_cli -host 192.168.137.253 -switchName tuya_1 -switchHost 192.168.137.15
    weather_cli -host 192.168.137.253 -latitude (latitude) -longitude (longitude) -appId (API key)

- 192.168.137.253 = MQTT broker
- 192.168.137.20 = zmote
- 192.168.137.252 = Philips Hue bridge
- /dev/ttyACM0 = USB serial port exposed by Arduino (when plugged into a Raspberry Pi)
- 192.168.137.15 = Powertech smart plug

NOTE: You can change the MQTT client provider between [GMQ](https://github.com/yosssi/gmq) and [Paho](https://github.com/eclipse/paho.mqtt.golang) by manipulating the `MQTT_CLIENT_PROVIDER` variable; e.g.:

    MQTT_CLIENT_PROVIDER=gmq aircon_cli -host 192.168.137.253 -aircon 192.168.137.20
    MQTT_CLIENT_PROVIDER=paho aircon_cli -host 192.168.137.253 -aircon 192.168.137.20
    MQTT_CLIENT_PROVIDER=libmqtt aircon_cli -host 192.168.137.253 -aircon 192.168.137.20
    
- `paho`
    - the first library I used- doesn't seem to throw errors on lost connectivity with the server (so you can't handle it)
- `gmq` (the default)
    - seems to be the most robust; also the only one I've implemented and error-handler-driven reconnect for
- `libmqtt`
    - doesn't support wildcards, seems a little flaky

## TODO

- Consider fixing implicit behaviour in light naming ("Some Lights" = "some-lights" and vice versa)
- Consider fixing failure to DRY throughout the cmd/*_cli/main.go files

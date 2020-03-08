# mqtt_things

This repo contains some Go code to expose a few different devices via MQTT and a generic "action" concept over the top of that. 

## What does it do?

I'll speak to the top level commands:

- `aircon_cli`
    - Listens for 0 or 1 at `home/inside/aircon/state/set`
    - Sends packets to a [https://www.zmote.io/](zmote) which sends IR blasts to Fujitsu aircon for two states
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

## How do I build it?

    ./build.sh
    
## How do I test it?

    ./test.sh

## How do I run it?

Here's a dump of all of the command lines as I'm using them around my house (with some data sanitised):

    aircon_cli -host 192.168.137.253 -aircon 192.168.137.20
    heater_cli -host 192.168.137.253 -port /dev/ttyACM0 -relay 1
    http_cli -host 192.168.137.253 -port 8079
    lights_cli -host 192.168.137.253 -apiKey (API key) -bridgeHost 192.168.137.252
    sprinklers_cli -host 192.168.137.253 -port /dev/ttyACM0 -relay 2 -relay 3
    weather_cli -host 192.168.137.253 -latitude (latitude) -longitude (longitude) -appId (API key)

- 192.168.137.253 = MQTT broker
- 192.168.137.20 = zmote
- 192.168.137.252 = Philips Hue bridge
- /dev/ttyACM0 = USB serial port exposed by Arduino (when plugged into a Raspberry Pi)

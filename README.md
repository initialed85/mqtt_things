# mqtt_things

# status: working but very coupled to my own personal projects

This repo is mostly Go, with a little bit of Rust and some Arduino-flavoured C++.

It's pretty tightly coupled to my home automation setup, but there might be a few useful things in there for you; of note:

-   `cmd`
    -   `aircons_cli`
        -   Limited MQTT integration of my old Fujtsi aircon and Mitsubishi aircon via Zmote
    -   `arp_cli`
        -   Linux-only MQTT integration that reports whether or not specific MAC addresses / IP addresses are present on the local network
        -   I was using this as a "home or not" sensor
    -   `circumstances_cli`
        -   Some bespoke stuff I was using in the pre-home-assistant days to publish composed states for me to do things with
            -   e.g. it's after this time of day and OpenWeather says its sunny
    -   `heater_cli`
        -   MQTT integration w/ `res/arduino` for controlling a relay that turns on / off the gas heater in my living room
    -   `http_cli`
        -   A generalized thing to expose the state of an MQTT broker's topics as a JSON HTTP API
    -   ## `lights_cli`
        -   Limited MQTT integration of Philips Hue lights
    -   `mqtt_to_glue_bridge`
        -   Bridge between MQTT and [Glue (my own brokerless pub-sub lib)](https://github.com/initialed85/glue)
    -   `sensors_cli`
        -   Limited MQTT integration of Philips Hue presence / temperature sensors
    -   `smart_aircons_cli`
        -   Limited MQTT integration of my old Fujitsi aircon, Mitsubishi aircon and new Fujitsu aircons via Zmote and Broadlink RM4 Mini
            -   There's a reasonable Broadlink RM4 Mini library you can use here
    -   `sprinklers_cli`
        -   MQTT integration w/ `res/arduino` for controlling two relays that turn on / off my banks of sprinklers
    -   ## `switches_cli`
        -   Limited MQTT integration for switching on / off some smartplugs running Tasmota firmware
    -   `topic_cli`
        -   Handy debugging tool for subscribing to / publishing to MQTT topics
    -   `topic_exporter_cli`
    -   A generalized thing to expose the state of an MQTT broker's topics as a Prometheus exporter
    -   `open_weather_cli`
        -   MQTT integration for OpenWeather
            -   NOTE: This uses the 2.5 API which they're apparently deprecating sometime in 2024, so it's basically just garbage now
-   `res`
    -   `arduino`
    -   `esp32`

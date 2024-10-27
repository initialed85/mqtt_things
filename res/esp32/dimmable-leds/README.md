# dimmable-leds

Ref.: https://docs.esp-rs.org/book/overview/using-the-standard-library.html

```shell
CRATE_CC_NO_DEFAULTS=1 cargo build --release
espflash flash target/xtensa-esp32-espidf/release/dimmable-leds --port /dev/cu.usbserial-0001 --baud 460800
espflash monitor --port /dev/cu.usbserial-0001
```

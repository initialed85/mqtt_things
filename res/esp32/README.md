# esp32

Setup:

```shell
cargo install espup
espup install
. /Users/edwardbeech/export-esp.sh
cargo install cargo-generate

# i used "dimmable-leds" as the name
cargo generate esp-rs/esp-idf-template cargo
cd dimmable leds

cargo install cargo-espflash
cargo install espflash
```

Development:

```shell
. /Users/edwardbeech/export-esp.sh
export CRATE_CC_NO_DEFAULTS=1

CRATE_CC_NO_DEFAULTS=1 cargo build

# I think this flashes?
CRATE_CC_NO_DEFAULTS=1 cargo run

# directly flashing a built artifact
espflash flash ./target/xtensa-esp32-espidf/debug/dimmable-leds -p /dev/cu.usbserial-0001 -B 460800
```

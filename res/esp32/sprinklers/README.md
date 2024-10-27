# sprinklers

Ref.: https://docs.esp-rs.org/book/overview/using-the-standard-library.html

```shell
cargo build --release
espflash flash target/xtensa-esp32-espidf/release/sprinklers --port /dev/cu.usbserial-0001 --baud 460800
espflash monitor --port /dev/cu.usbserial-0001
```

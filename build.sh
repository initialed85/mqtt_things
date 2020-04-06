#!/usr/bin/env bash

set -e -x

mkdir -p bin/native || true

go build -v -o bin/native/aircon_cli cmd/aircon_cli/main.go
go build -v -o bin/native/heater_cli cmd/heater_cli/main.go
go build -v -o bin/native/http_cli cmd/http_cli/main.go
go build -v -o bin/native/lights_cli cmd/lights_cli/main.go
go build -v -o bin/native/mqtt_cli cmd/mqtt_cli/main.go
go build -v -o bin/native/sprinklers_cli cmd/sprinklers_cli/main.go
go build -v -o bin/native/switches_cli cmd/switches_cli/main.go
go build -v -o bin/native/weather_cli cmd/weather_cli/main.go

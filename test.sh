#!/usr/bin/env bash

set -e -x

function teardown() {
  docker rm -f eclipse-mosquitto || true
}
trap teardown EXIT

function setup() {
  docker run -d --restart=always \
    --name eclipse-mosquitto \
    -p 1883:1883 \
    -p 9001:9001 \
    eclipse-mosquitto

  sleep 1
}

setup

if [[ "${*}" == "" ]]; then
  go test -v ./pkg/aircons_client
  go test -v ./pkg/cirumstances_engine
  go test -v ./pkg/lights_client
  go test -v ./pkg/mqtt_action_router
  go test -v ./pkg/mqtt_client
  go test -v ./pkg/relays_client
  go test -v ./pkg/switches_client
  go test -v ./pkg/weather_client
else
  go test -v "${@}"
fi

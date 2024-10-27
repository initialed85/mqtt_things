#!/usr/bin/env bash

set -e -m

function cleanup() {
  true
}
trap cleanup EXIT

function build_amd64() {
  _=${1?:first argument must be CMD_NAME}
  _=${2?:second argument must be Docker image name part}

  docker build \
    --platform linux/amd64 \
    --build-arg CMD_NAME="${1}" \
    -f docker/cli/Dockerfile \
    -t "initialed85/mqtt-things-${2}:latest" \
    .

    docker image push "initialed85/mqtt-things-${2}:latest"
}

function build_arm64() {
  _=${1?:first argument must be CMD_NAME}
  _=${2?:second argument must be Docker image name part}

  docker build \
    --platform linux/arm64 \
    --build-arg CMD_NAME="${1}" \
    -f docker/cli/Dockerfile \
    -t "initialed85/mqtt-things-${2}:latest" \
    .

    docker image push "initialed85/mqtt-things-${2}:latest"
}

# for the pi under the house
build_arm64 "sprinklers_cli" "sprinklers-cli"

# for the cluster
build_amd64 "sensors_cli" "sensors-cli"
build_amd64 "smart_aircons_cli" "smart-aircons-cli"
build_amd64 "weather_cli" "weather-cli"
build_amd64 "http_cli" "http-cli"
build_amd64 "topic_exporter_cli" "topic-exporter-cli"

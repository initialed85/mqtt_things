#!/usr/bin/env bash

set -e

docker buildx create --use --name mqtt_things

function cleanup() {
  docker buildx rm mqtt_things || true
}
trap cleanup EXIT

function build_amd64() {
  _=${1?:first argument must be CMD_NAME}
  _=${2?:second argument must be Docker image name part}

  docker buildx build \
    --platform linux/amd64 \
    --build-arg CMD_NAME="${1}" \
    -f docker/cli/Dockerfile \
    -t "initialed85/mqtt-things-${2}:latest" \
    --push \
    .
}

function build_arm64() {
  _=${1?:first argument must be CMD_NAME}
  _=${2?:second argument must be Docker image name part}

  docker buildx build \
    --platform linux/arm64 \
    --build-arg CMD_NAME="${1}" \
    -f docker/cli/Dockerfile \
    -t "initialed85/mqtt-things-${2}:latest" \
    --push \
    .
}

# for the pi under the house
build_arm64 "sprinklers_cli" "sprinklers-cli"

# for the cluster
build_amd64 "sensors_cli" "sensors-cli"
build_amd64 "smart_aircons_cli" "smart-aircons-cli"
build_amd64 "weather_cli" "weather-cli"

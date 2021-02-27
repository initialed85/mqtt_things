package smart_aircons_client

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func OnToPayload(on bool) (string, error) {
	if on {
		return "ON", nil
	} else {
		return "OFF", nil
	}
}

func PayloadToOn(payload string) (bool, error) {
	payload = strings.ToUpper(strings.TrimSpace(payload))
	if payload == "ON" {
		return true, nil
	} else if payload == "OFF" {
		return false, nil
	}

	return false, fmt.Errorf("%#+v not one of %#+v or %#+v", payload, "ON", "OFF")
}

func ModeToPayload(mode string) (string, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))

	if mode != "cool" && mode != "heat" && mode != "fan_only" {
		return "", fmt.Errorf("%#+v not one of %#+v, %#+v or %#+v", mode, "cool", "heat", "fan_only")
	}

	return mode, nil
}

func PayloadToMode(payload string) (string, error) {
	return ModeToPayload(payload)
}

func TemperatureToPayload(temperature int64) (string, error) {
	if temperature < 18 || temperature > 30 {
		return "", fmt.Errorf("%#+v out of range 18 - 30 inclusive", temperature)
	}

	return strconv.FormatFloat(float64(temperature), 'f', 1, 64), nil
}

func PayloadToTemperature(payload string) (int64, error) {
	temperatureAsFloat, err := strconv.ParseFloat(payload, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float from %#+v for rounding", payload)
	}

	temperature := int64(math.Round(temperatureAsFloat))

	// TODO: strictly speaking because of the float round, this <= 17.49r and >= 30.5
	if temperature < 18 || temperature > 30 {
		return 0, fmt.Errorf("%#+v out of range 18 - 30 inclusive", temperature)
	}

	return temperature, nil
}

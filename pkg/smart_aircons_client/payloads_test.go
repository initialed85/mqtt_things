package smart_aircons_client

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func TestPayloads_On(t *testing.T) {
	payload := ""
	on := false
	var err error

	payload, err = OnToPayload(true)
	handleError(err)
	assert.Equal(t, "ON", payload)

	payload, err = OnToPayload(false)
	handleError(err)
	assert.Equal(t, "OFF", payload)

	on, err = PayloadToOn("ON")
	handleError(err)
	assert.Equal(t, true, on)

	on, err = PayloadToOn("OFF")
	handleError(err)
	assert.Equal(t, false, on)

	_, err = PayloadToOn("HAM")
	assert.NotNil(t, err)
}

func TestPayloads_Mode(t *testing.T) {
	payload := ""
	mode := ""
	var err error

	payload, err = ModeToPayload("cool")
	handleError(err)
	assert.Equal(t, "cool", payload)

	payload, err = ModeToPayload("heat")
	handleError(err)
	assert.Equal(t, "heat", payload)

	payload, err = ModeToPayload("fan_only")
	handleError(err)
	assert.Equal(t, "fan_only", payload)

	_, err = ModeToPayload("HAM")
	assert.NotNil(t, err)

	mode, err = PayloadToMode("cool")
	handleError(err)
	assert.Equal(t, "cool", mode)

	mode, err = PayloadToMode("heat")
	handleError(err)
	assert.Equal(t, "heat", mode)

	mode, err = PayloadToMode("fan_only")
	handleError(err)
	assert.Equal(t, "fan_only", mode)

	_, err = PayloadToMode("HAM")
	assert.NotNil(t, err)
}

func TestPayloads_Temperature(t *testing.T) {
	payload := ""
	var temperature int64
	var err error

	payload, err = TemperatureToPayload(18)
	handleError(err)
	assert.Equal(t, "18.0", payload)

	payload, err = TemperatureToPayload(30)
	handleError(err)
	assert.Equal(t, "30.0", payload)

	_, err = TemperatureToPayload(17)
	assert.NotNil(t, err)

	_, err = TemperatureToPayload(31)
	assert.NotNil(t, err)

	temperature, err = PayloadToTemperature("17.5")
	handleError(err)
	assert.Equal(t, int64(18), temperature)

	temperature, err = PayloadToTemperature("18.0")
	handleError(err)
	assert.Equal(t, int64(18), temperature)

	temperature, err = PayloadToTemperature("18.1")
	handleError(err)
	assert.Equal(t, int64(18), temperature)

	temperature, err = PayloadToTemperature("30.0")
	handleError(err)
	assert.Equal(t, int64(30), temperature)

	temperature, err = PayloadToTemperature("29.9")
	handleError(err)
	assert.Equal(t, int64(30), temperature)

	temperature, err = PayloadToTemperature("30.49")
	handleError(err)
	assert.Equal(t, int64(30), temperature)

	_, err = PayloadToTemperature("17.49")
	assert.NotNil(t, err)

	_, err = PayloadToTemperature("30.51")
	assert.NotNil(t, err)

	_, err = PayloadToTemperature("HAM")
	assert.NotNil(t, err)
}

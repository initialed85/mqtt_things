package smart_aircons_client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAircon_On(t *testing.T) {
	lastOn := false
	onChanged := false
	var err error

	a := NewModel(
		func(on bool, mode string, temperature int64) error {
			lastOn = on
			onChanged = true

			return nil
		},
	)

	err = a.SetOn(true)
	handleError(err)
	assert.True(t, lastOn)
	assert.True(t, onChanged)

	err = a.SetOn(false)
	handleError(err)
	assert.False(t, lastOn)
	assert.True(t, onChanged)

	onChanged = false

	err = a.SetOn(true)
	handleError(err)
	assert.True(t, lastOn)
	assert.True(t, onChanged)
}

func TestAircon_Mode(t *testing.T) {
	lastMode := ""
	onChanged := false
	var err error

	a := NewModel(
		func(on bool, mode string, temperature int64) error {
			lastMode = mode
			onChanged = true

			return nil
		},
	)

	err = a.SetMode("cool")
	handleError(err)
	assert.Equal(t, "cool", lastMode)
	assert.True(t, onChanged)

	err = a.SetMode("heat")
	handleError(err)
	assert.Equal(t, "heat", lastMode)
	assert.True(t, onChanged)

	onChanged = false

	err = a.SetMode("fan_only")
	handleError(err)
	assert.Equal(t, "fan_only", lastMode)
	assert.True(t, onChanged)

	err = a.SetMode("HAM")
	assert.NotNil(t, err)
}

func TestAircon_Temperature(t *testing.T) {
	lastTemperature := int64(0)
	onChanged := false
	var err error

	a := NewModel(
		func(on bool, mode string, temperature int64) error {
			lastTemperature = temperature
			onChanged = true

			return nil
		},
	)

	err = a.SetTemperature(25)
	handleError(err)
	assert.Equal(t, int64(25), lastTemperature)
	assert.True(t, onChanged)

	err = a.SetTemperature(23)
	handleError(err)
	assert.Equal(t, int64(23), lastTemperature)
	assert.True(t, onChanged)

	onChanged = false

	err = a.SetTemperature(24)
	handleError(err)
	assert.Equal(t, int64(24), lastTemperature)
	assert.True(t, onChanged)

	err = a.SetTemperature(17)
	assert.NotNil(t, err)
}

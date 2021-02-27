package smart_aircons_client

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
)

func getTopic(topic string, isGet bool) string {
	getOrSet := "set"
	if isGet {
		getOrSet = "get"
	}

	return fmt.Sprintf(
		"home/inside/smart-aircons/living-room/%v/%v",
		topic,
		getOrSet,
	)
}

func getMessage(topic string, payload string) mqtt.Message {
	return mqtt.Message{
		Topic:   topic,
		Payload: payload,
	}
}

func TestRouter_On(t *testing.T) {
	lastOn := false
	onChanged := false

	router := NewRouter(
		"home/inside/smart-aircons/living-room",
		func(on bool) error {
			lastOn = on
			onChanged = true

			return nil
		},
		func(mode string) error {
			return nil
		},
		func(temperature int64) error {
			return nil
		},
	)

	// restoring state from the broker
	router.EnableGetCallbacks()
	router.Handle(getMessage(getTopic(topicOnInfix, true), "ON"))
	assert.True(t, lastOn)
	assert.True(t, onChanged)

	// new state
	onChanged = false
	router.DisableGetCallbacks()
	router.Handle(getMessage(getTopic(topicOnInfix, false), "OFF"))
	assert.False(t, lastOn)
	assert.True(t, onChanged)

	// new state
	onChanged = false
	router.Handle(getMessage(getTopic(topicOnInfix, false), "ON"))
	assert.True(t, lastOn)
	assert.True(t, onChanged)

	// insane state
	onChanged = false
	router.Handle(getMessage(getTopic(topicOnInfix, false), "HAM"))
	assert.False(t, onChanged)
}

func TestRouter_Mode(t *testing.T) {
	lastMode := ""
	onChanged := false

	router := NewRouter(
		"home/inside/smart-aircons/living-room",
		func(on bool) error {
			return nil
		},
		func(mode string) error {
			lastMode = mode
			onChanged = true

			return nil
		},
		func(temperature int64) error {
			return nil
		},
	)

	// restoring state from the broker
	router.EnableGetCallbacks()
	router.Handle(getMessage(getTopic(topicModeInfix, true), "cool"))
	assert.Equal(t, "cool", lastMode)
	assert.True(t, onChanged)

	// new state
	onChanged = false
	router.DisableGetCallbacks()
	router.Handle(getMessage(getTopic(topicModeInfix, false), "heat"))
	assert.Equal(t, "heat", lastMode)
	assert.True(t, onChanged)

	// new state
	onChanged = false
	router.Handle(getMessage(getTopic(topicModeInfix, false), "fan_only"))
	assert.Equal(t, "fan_only", lastMode)
	assert.True(t, onChanged)

	// insane state
	onChanged = false
	router.Handle(getMessage(getTopic(topicModeInfix, false), "HAM"))
	assert.False(t, onChanged)
}

func TestRouter_Temperature(t *testing.T) {
	lastTemperature := int64(0)
	onChanged := false

	router := NewRouter(
		"home/inside/smart-aircons/living-room",
		func(on bool) error {
			return nil
		},
		func(mode string) error {
			return nil
		},
		func(temperature int64) error {
			lastTemperature = temperature
			onChanged = true

			return nil
		},
	)

	// restoring state from the broker
	router.EnableGetCallbacks()
	router.Handle(getMessage(getTopic(topicTemperatureInfix, true), "24.0"))
	assert.Equal(t, int64(24), lastTemperature)
	assert.True(t, onChanged)

	// new state
	onChanged = false
	router.DisableGetCallbacks()
	router.Handle(getMessage(getTopic(topicTemperatureInfix, false), "23.0"))
	assert.Equal(t, int64(23), lastTemperature)
	assert.True(t, onChanged)

	// new state
	onChanged = false
	router.Handle(getMessage(getTopic(topicTemperatureInfix, false), "24.0"))
	assert.Equal(t, int64(24), lastTemperature)
	assert.True(t, onChanged)

	// insane state
	onChanged = false
	router.Handle(getMessage(getTopic(topicTemperatureInfix, false), "HAM"))
	assert.False(t, onChanged)
}

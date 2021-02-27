package smart_aircons_client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_On(t *testing.T) {
	lastHost := ""
	lastCode := ""
	lastTopic := ""
	lastPayload := ""

	client := NewClient(
		"home/inside/smart-aircons/living-room",
		"some_host",
		"fujitsu",
		func(host string, code string) error {
			lastHost = host
			lastCode = code

			return nil
		},
		func(topic string, qos byte, retained bool, payload interface{}) error {
			lastTopic = topic
			lastPayload = payload.(string)

			return nil
		},
	)

	//
	// restore state from the broker
	//

	client.EnableRestoreMode()

	client.Handle(getMessage(getTopic("power", true), "ON"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["fan_only"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/power/get", lastTopic)
	assert.Equal(t, "ON", lastPayload)

	client.Handle(getMessage(getTopic("mode", true), "cool"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["cool_24"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/mode/get", lastTopic)
	assert.Equal(t, "cool", lastPayload)

	client.Handle(getMessage(getTopic("temperature", true), "23.0"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["cool_23"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/temperature/get", lastTopic)
	assert.Equal(t, "23.0", lastPayload)

	client.DisableRestoreMode()

	//
	// set to off
	//

	client.Handle(getMessage(getTopic("power", false), "OFF"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["off"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/power/get", lastTopic)
	assert.Equal(t, "OFF", lastPayload)

	client.Handle(getMessage(getTopic("mode", false), "cool"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["off"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/mode/get", lastTopic)
	assert.Equal(t, "cool", lastPayload)

	client.Handle(getMessage(getTopic("temperature", false), "23.0"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["off"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/temperature/get", lastTopic)
	assert.Equal(t, "23.0", lastPayload)

	//
	// set to fan mode
	//

	client.Handle(getMessage(getTopic("power", false), "ON"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["cool_23"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/power/get", lastTopic)
	assert.Equal(t, "ON", lastPayload)

	client.Handle(getMessage(getTopic("mode", false), "fan_only"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["fan_only"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/mode/get", lastTopic)
	assert.Equal(t, "fan_only", lastPayload)

	client.Handle(getMessage(getTopic("temperature", false), "24.0"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["fan_only"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/temperature/get", lastTopic)
	assert.Equal(t, "24.0", lastPayload)

	//
	// set to heating
	//

	client.Handle(getMessage(getTopic("power", false), "ON"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["fan_only"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/power/get", lastTopic)
	assert.Equal(t, "ON", lastPayload)

	client.Handle(getMessage(getTopic("mode", false), "heat"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["heat_24"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/mode/get", lastTopic)
	assert.Equal(t, "heat", lastPayload)

	client.Handle(getMessage(getTopic("temperature", false), "23.0"))
	assert.Equal(t, "some_host", lastHost)
	assert.Equal(t, allCodes["fujitsu"]["heat_23"], lastCode)
	assert.Equal(t, "home/inside/smart-aircons/living-room/temperature/get", lastTopic)
	assert.Equal(t, "23.0", lastPayload)
}

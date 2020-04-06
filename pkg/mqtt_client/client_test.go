package mqtt_client

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	topic = "some_topic"
)

func TestClient_All(t *testing.T) {
	// NOTE: requires a local eclipse-mosquitto Docker container to be running
	enableTestMode("127.0.0.1")

	client := New("192.168.137.253", "", "")

	err := client.Connect()
	assert.Nil(t, err)

	var capturedMessage *Message
	callback := func(message Message) {
		capturedMessage = &message
	}

	err = client.Subscribe(topic, ExactlyOnce, callback)
	assert.Nil(t, err)

	err = client.Publish(topic, ExactlyOnce, false, "Some data")
	assert.Nil(t, err)

	started := time.Now()
	timeout := time.Second * 10
	timeoutTime := started.Add(timeout)

	for timeoutTime.Sub(time.Now()) > 0 && capturedMessage == nil {
		time.Sleep(time.Millisecond * 100)
	}

	assert.NotNil(t, capturedMessage, "failed to receive a message in %v", timeout)
	if capturedMessage != nil {
		assert.Equal(t, false, capturedMessage.Duplicate)
		assert.Equal(t, ExactlyOnce, capturedMessage.Qos)
		assert.Equal(t, topic, capturedMessage.Topic)
		assert.Equal(t, "Some data", capturedMessage.Payload)

		// not really sure what Paho's Message.Ack() does; skipping it for now
		// capturedMessage.Ack()
	}

	err = client.Disconnect()
	assert.Nil(t, err)
}

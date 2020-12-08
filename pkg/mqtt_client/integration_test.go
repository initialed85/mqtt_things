package mqtt_client

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testTopic = "some_topic"
	testHost  = "127.0.0.1"
)

func executeTestCase(t *testing.T, client Client) {
	err := client.Connect()
	assert.Nil(t, err)

	time.Sleep(time.Millisecond * 100)

	var capturedMessage *Message
	callback := func(message Message) {
		capturedMessage = &message
	}

	err = client.Subscribe(testTopic, ExactlyOnce, callback)
	assert.Nil(t, err)

	time.Sleep(time.Millisecond * 100)

	err = client.Publish(testTopic, ExactlyOnce, false, "Some data")
	assert.Nil(t, err)

	time.Sleep(time.Millisecond * 100)

	started := time.Now()
	timeout := time.Second * 5
	timeoutTime := started.Add(timeout)

	for timeoutTime.Sub(time.Now()) > 0 && capturedMessage == nil {
		time.Sleep(time.Millisecond * 100)
	}

	assert.NotNil(t, capturedMessage, "failed to receive a message in %v", timeout)
	if capturedMessage != nil {
		assert.Equal(t, testTopic, capturedMessage.Topic)
		assert.Equal(t, "Some data", capturedMessage.Payload)

		// not really sure what Paho's Message.Ack() does; skipping it for now
		// capturedMessage.Ack()
	}

	err = client.Disconnect()
	assert.Nil(t, err)
}

func TestPaho(t *testing.T) {
	EnablePahoTestMode(testHost)

	client := GetPahoClient(testHost, "", "", func(client Client, err error) {
		log.Fatal(err)
	})

	executeTestCase(t, client)
}

func TestGMQ(t *testing.T) {
	EnableGMQTestMode(testHost)

	client := GetGMQClient(testHost, "", "", func(client Client, err error) {
		log.Fatal(err)
	})

	executeTestCase(t, client)
}

func TestLibMQTT(t *testing.T) {
	EnableLibMQTTTestMode(testHost)

	client := GetLibMQTTClient(testHost, "", "", func(client Client, err error) {
		log.Fatal(err)
	})

	executeTestCase(t, client)
}

func TestGlue(t *testing.T) {
	client := GetGlueClient(testHost, "", "", func(client Client, err error) {
		log.Fatal(err)
	})

	executeTestCase(t, client)
}

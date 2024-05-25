package mqtt_client

import (
	"fmt"
	"time"

	"github.com/initialed85/glue/pkg/endpoint"
	"github.com/initialed85/glue/pkg/topics"
)

type GlueClient struct {
	endpointManager *endpoint.Manager
}

func NewGlueClient(host, username, password string, errorHandler func(Client, error)) (*GlueClient, error) {
	_ = host
	_ = username
	_ = password
	_ = errorHandler

	endpointManager, err := endpoint.NewManagerSimple()
	if err != nil {
		return nil, err
	}

	c := GlueClient{
		endpointManager: endpointManager,
	}

	return &c, nil
}

func (c *GlueClient) Connect() error {
	endpointManager, err := endpoint.NewManagerSimple()
	if err != nil {
		return err
	}

	c.endpointManager = endpointManager
	c.endpointManager.Start()

	return nil
}

func (c *GlueClient) Publish(topic string, qos byte, retained bool, payload interface{}, quiet ...bool) error {
	stringPayload := payload.(string)
	bytePayload := []byte(stringPayload)

	return c.endpointManager.Publish(
		topic,
		"bytes",
		time.Second,
		bytePayload,
	)
}

func (c *GlueClient) Subscribe(topic string, qos byte, callback func(message Message)) error {
	if c.endpointManager == nil {
		return fmt.Errorf("endpointManager is nil (probably not connected)")
	}

	return c.endpointManager.Subscribe(
		topic,
		"bytes",
		func(topicsMessage *topics.Message) {
			message := Message{
				Received:  topicsMessage.Timestamp,
				Topic:     topicsMessage.TopicName,
				MessageID: uint16(topicsMessage.SequenceNumber),
				Payload:   string(topicsMessage.Payload),
			}

			callback(message)
		},
	)
}

func (c *GlueClient) Unsubscribe(topic string) error {
	if c.endpointManager == nil {
		return fmt.Errorf("endpointManager is nil (probably not connected)")
	}

	return c.endpointManager.Unsubscribe(topic)
}

func (c *GlueClient) Disconnect() error {
	if c.endpointManager == nil {
		return fmt.Errorf("endpointManager is nil (probably not connected)")
	}

	c.endpointManager.Stop()

	c.endpointManager = nil

	return nil
}

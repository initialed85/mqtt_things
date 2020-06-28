package mqtt_client

import (
	"fmt"
	"github.com/goiiot/libmqtt"
	"sync"
	"time"
)

var LibMQTTTestMode = false
var LibMQTTTestHost string

func EnableLibMQTTTestMode(host string) {
	LibMQTTTestMode = true
	LibMQTTTestHost = host
}

func getQosLevel(qos byte) (libmqtt.QosLevel, error) {
	if qos == 0 {
		return libmqtt.Qos0, nil
	} else if qos == 1 {
		return libmqtt.Qos1, nil
	} else if qos == 2 {
		return libmqtt.Qos2, nil
	} else {
		return libmqtt.Qos0, fmt.Errorf("invalid qos byte %+v", qos)
	}
}

type LibMQTTClient struct {
	clientID, host, username, password string
	client                             libmqtt.Client
	errorHandler                       func(Client, error)
}

func NewLibMQTTClient(host, username, password string, errorHandler func(Client, error)) (c *LibMQTTClient) {
	if LibMQTTTestMode {
		host = LibMQTTTestHost
	}

	clientID := getClientID("libmqtt")

	return &LibMQTTClient{
		clientID:     clientID,
		host:         host,
		username:     username,
		password:     password,
		errorHandler: errorHandler,
	}
}

func (c *LibMQTTClient) Connect() error {
	newClient, err := libmqtt.NewClient(
		libmqtt.WithDialTimeout(5),
		libmqtt.WithClientID(c.clientID),
		libmqtt.WithIdentity(c.username, c.password),
		libmqtt.WithKeepalive(5, 1.2),
		libmqtt.WithNetHandleFunc(func(_ libmqtt.Client, _ string, err error) {
			go func() {
				time.Sleep(time.Second)

				c.errorHandler(c, err)
			}()
		}),
	)
	if err != nil {
		return err
	}

	c.client = newClient

	var connErr error
	var wg sync.WaitGroup
	connHandleFunc := func(client libmqtt.Client, server string, code byte, err error) {
		if err != nil {
			connErr = err
		}

		if code != libmqtt.CodeSuccess {
			connErr = fmt.Errorf("failed to connect; code was %+v", code)
		}

		wg.Done()
	}

	wg.Add(1)

	err = c.client.ConnectServer(
		fmt.Sprintf("%v:1883", c.host),
		libmqtt.WithConnHandleFunc(connHandleFunc),
	)

	if err != nil {
		return err
	}

	wg.Wait()

	return connErr
}

func (c *LibMQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}

	qosLevel, err := getQosLevel(qos)
	if err != nil {
		return err
	}

	c.client.Publish(
		&libmqtt.PublishPacket{
			TopicName: topic,
			Qos:       qosLevel,
			IsRetain:  retained,
			Payload:   []byte(fmt.Sprintf("%v", payload)),
		},
	)

	return nil
}

func (c *LibMQTTClient) Subscribe(topic string, qos byte, callback func(message Message)) error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}

	qosLevel, err := getQosLevel(qos)
	if err != nil {
		return err
	}

	topicHandleFunc := func(client libmqtt.Client, topic string, qos libmqtt.QosLevel, msg []byte) {
		callback(
			Message{
				Received:  time.Now(),
				Topic:     topic,
				MessageID: 0,
				Payload:   string(msg),
			})
	}

	c.client.HandleTopic(topic, topicHandleFunc)

	c.client.Subscribe(&libmqtt.Topic{
		Name: topic,
		Qos:  qosLevel,
	})

	return nil
}

func (c *LibMQTTClient) Unsubscribe(topic string) error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}

	c.client.Unsubscribe(topic)

	return nil
}

func (c *LibMQTTClient) Disconnect() error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}

	c.client.Destroy(false)

	c.client = nil

	return nil
}

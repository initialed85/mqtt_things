package libmqtt_mqtt_client

import (
	"fmt"
	"github.com/goiiot/libmqtt"
	"github.com/google/uuid"
	"github.com/initialed85/mqtt_things/pkg/mqtt_common"
	"log"
	"sync"
	"time"
)

var TestMode = false
var TestHost string

func enableTestMode(host string) {
	TestMode = true
	TestHost = host
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

type Client struct {
	clientID, host, username, password string
	client                             libmqtt.Client
}

func New(host, username, password string) (c *Client) {
	if TestMode {
		host = TestHost
	}

	clientID := "libmqtt_"
	uuid4, err := uuid.NewRandom()
	if err != nil {
		clientID += fmt.Sprintf("unknown_%+v", time.Now().UnixNano())
	} else {
		clientID += uuid4.String()
	}

	c = &Client{
		clientID: clientID,
		host:     host,
		username: username,
		password: password,
	}

	log.Printf("created")

	return c
}

func (c *Client) Connect() error {
	log.Printf("connecting")

	newClient, err := libmqtt.NewClient(
		libmqtt.WithDialTimeout(5),
		libmqtt.WithClientID(c.clientID),
		libmqtt.WithIdentity(c.username, c.password),
		libmqtt.WithKeepalive(5, 1.2),
		libmqtt.WithLog(libmqtt.Debug),
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

func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	log.Printf("publishing %v to %v with qos %v and retained %v", topic, payload, qos, retained)

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

func (c *Client) Subscribe(topic string, qos byte, callback func(message mqtt_common.Message)) error {
	log.Printf("subscribing to %v callback %p and qos %v", topic, callback, qos)

	qosLevel, err := getQosLevel(qos)
	if err != nil {
		return err
	}

	topicHandleFunc := func(client libmqtt.Client, topic string, qos libmqtt.QosLevel, msg []byte) {
		callback(
			mqtt_common.Message{
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

func (c *Client) Unsubscribe(topic string) error {
	log.Printf("unsubscribing from %v", topic)

	c.client.Unsubscribe(topic)

	return nil
}

func (c *Client) Disconnect() error {
	log.Printf("disconnecting")

	c.client.Destroy(false)

	c.client = nil

	return nil
}

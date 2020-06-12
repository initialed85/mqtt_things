package gmq_mqtt_client

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/initialed85/mqtt_things/pkg/mqtt_common"
	"github.com/yosssi/gmq/mqtt/client"
	"log"
	"time"
)

var TestMode = false
var TestHost string

func enableTestMode(host string) {
	TestMode = true
	TestHost = host
}

type Client struct {
	connectOptions client.ConnectOptions
	options        client.Options
	client         *client.Client
}

func New(host, username, password string) (c *Client) {
	if TestMode {
		host = TestHost
	}

	clientID := "gmq_"
	uuid4, err := uuid.NewRandom()
	if err != nil {
		clientID += fmt.Sprintf("unknown_%+v", time.Now().UnixNano())
	} else {
		clientID += uuid4.String()
	}

	c = &Client{}

	c.connectOptions = client.ConnectOptions{
		Network:         "tcp",
		Address:         fmt.Sprintf("%v:1883", host),
		CONNACKTimeout:  time.Second * 5,
		PINGRESPTimeout: time.Second * 5,
		ClientID:        []byte(clientID),
		UserName:        []byte(username),
		Password:        []byte(password),
		KeepAlive:       1,
	}

	c.options = client.Options{
		ErrorHandler: func(err error) {
			panic(err)
		},
	}

	log.Printf("created %+v, %+v", c.connectOptions, c.options)

	return c
}

func (c *Client) Connect() error {
	log.Printf("connecting")

	c.client = client.New(&c.options)

	return c.client.Connect(&c.connectOptions)
}

func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	log.Printf("publishing %v to %v with qos %v and retained %v", topic, payload, qos, retained)

	return c.client.Publish(
		&client.PublishOptions{
			QoS:       qos,
			Retain:    retained,
			TopicName: []byte(topic),
			Message:   []byte(fmt.Sprintf("%v", payload)),
		},
	)
}

func (c *Client) Subscribe(topic string, qos byte, callback func(message mqtt_common.Message)) error {
	wrappedCallback := func(topicName, message []byte) {
		callback(mqtt_common.Message{
			Received:  time.Now(),
			Topic:     string(topicName),
			MessageID: 0,
			Payload:   string(message),
		})
	}

	log.Printf("subscribing to %v callback %p and qos %v", topic, callback, qos)

	return c.client.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			{
				TopicFilter: []byte(topic),
				QoS:         qos,
				Handler:     wrappedCallback,
			},
		},
	})
}

func (c *Client) Unsubscribe(topic string) error {
	log.Printf("unsubscribing from %v", topic)

	return c.client.Unsubscribe(&client.UnsubscribeOptions{
		TopicFilters: [][]byte{
			[]byte(topic),
		},
	})
}

func (c *Client) Disconnect() error {
	log.Printf("disconnecting")

	err := c.client.Disconnect()

	c.client = nil

	return err
}

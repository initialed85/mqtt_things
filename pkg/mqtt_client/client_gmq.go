package mqtt_client

import (
	"crypto/tls"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/yosssi/gmq/mqtt/client"
)

var GMQTestMode = false
var GMQTestHost string

func EnableGMQTestMode(host string) {
	GMQTestMode = true
	GMQTestHost = host
}

type GMQClient struct {
	connectOptions client.ConnectOptions
	options        client.Options
	client         *client.Client
	errorHandler   func(Client, error)
}

func NewGMQClient(host, username, password string, errorHandler func(Client, error)) (c *GMQClient) {
	port := 1883

	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		rawPort := parts[1]
		possiblePort, _ := strconv.ParseInt(rawPort, 10, 64)
		if possiblePort > 0 {
			host = parts[0]
			port = int(possiblePort)
		}
	}

	if GMQTestMode {
		host = GMQTestHost
	}

	clientID := getClientID("gmq")

	c = &GMQClient{
		errorHandler: errorHandler,
	}

	c.connectOptions = client.ConnectOptions{
		Network: "tcp",
		Address: fmt.Sprintf("%v:%v", host, port),
		// CONNACKTimeout:  time.Second * 10,
		// PINGRESPTimeout: time.Second * 10,
		ClientID: []byte(clientID),
		UserName: []byte(username),
		Password: []byte(password),
		// KeepAlive:       5,
	}

	if port == 8883 {
		c.connectOptions.TLSConfig = &tls.Config{}
	}

	c.options = client.Options{
		ErrorHandler: func(err error) {
			log.Printf("%+v caught; firing %p", err, c.errorHandler)

			go func() {
				time.Sleep(time.Second)

				c.errorHandler(c, err)
			}()
		},
	}

	return c
}

func (c *GMQClient) Connect() error {
	c.client = client.New(&c.options)

	return c.client.Connect(&c.connectOptions)
}

func (c *GMQClient) Publish(topic string, qos byte, retained bool, payload interface{}, quiet ...bool) error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}

	return c.client.Publish(
		&client.PublishOptions{
			QoS:       qos,
			Retain:    retained,
			TopicName: []byte(topic),
			Message:   []byte(fmt.Sprintf("%v", payload)),
		},
	)
}

func (c *GMQClient) Subscribe(topic string, qos byte, callback func(message Message)) error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}
	wrappedCallback := func(topicName, message []byte) {
		callback(Message{
			Received:  time.Now(),
			Topic:     string(topicName),
			MessageID: 0,
			Payload:   string(message),
		})
	}

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

func (c *GMQClient) Unsubscribe(topic string) error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}

	return c.client.Unsubscribe(&client.UnsubscribeOptions{
		TopicFilters: [][]byte{
			[]byte(topic),
		},
	})
}

func (c *GMQClient) Disconnect() error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}

	err := c.client.Disconnect()

	c.client = nil

	return err
}

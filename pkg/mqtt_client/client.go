package mqtt_client

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"time"
)

const (
	AtMostOnce  = byte(0)
	AtLeastOnce = byte(1)
	ExactlyOnce = byte(2)
)

var TestMode = false
var TestHost string

func enableTestMode(host string) {
	TestMode = true
	TestHost = host
}

type Message struct {
	Received  time.Time
	Duplicate bool
	Qos       byte
	Retained  bool
	Topic     string
	MessageID uint16
	Payload   string
	Ack       func()
}

type Client struct {
	clientOptions *mqtt.ClientOptions
	connectToken  mqtt.Token
	client        mqtt.Client
}

func New(host, username, password string) (client Client) {
	if TestMode {
		host = TestHost
	}

	client.clientOptions = mqtt.NewClientOptions()
	client.clientOptions.AddBroker(fmt.Sprintf("tcp://%v:1883", host))
	client.clientOptions.SetUsername(username)
	client.clientOptions.SetPassword(password)
	client.clientOptions.SetPingTimeout(time.Second * 1)
	client.clientOptions.SetConnectTimeout(time.Second * 5)
	client.clientOptions.SetWriteTimeout(time.Second * 5)
	client.clientOptions.SetAutoReconnect(true)
	client.clientOptions.SetResumeSubs(true)

	client.client = mqtt.NewClient(client.clientOptions)

	log.Printf("created %+v", client)

	return client
}

func (c *Client) Connect() error {
	log.Printf("connecting")

	c.connectToken = c.client.Connect()
	if c.connectToken == nil {
		return fmt.Errorf("nil token while connecting")
	}

	if !c.connectToken.WaitTimeout(c.clientOptions.ConnectTimeout) {
		return fmt.Errorf("connection timed out after %v", c.clientOptions.ConnectTimeout)
	}

	return c.connectToken.Error()
}

func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	log.Printf("publishing %v to %v with qos %v and retained %v", topic, payload, qos, retained)

	token := c.client.Publish(topic, qos, retained, payload)
	if token == nil {
		return fmt.Errorf("nil token while publishing (%v, %v, %v, %v)", topic, qos, retained, payload)
	}

	if !token.WaitTimeout(c.clientOptions.WriteTimeout) {
		return fmt.Errorf("publish timed out after %v", c.clientOptions.ConnectTimeout)
	}

	return token.Error()
}

func (c *Client) Subscribe(topic string, qos byte, callback func(Message)) error {
	wrappedCallback := func(client mqtt.Client, message mqtt.Message) {
		callback(Message{
			Received:  time.Now(),
			Duplicate: message.Duplicate(),
			Qos:       message.Qos(),
			Retained:  message.Retained(),
			Topic:     message.Topic(),
			MessageID: message.MessageID(),
			Payload:   string(message.Payload()),
			Ack:       message.Ack,
		})
	}

	log.Printf("subscribing to %v callback %p and qos %v", topic, callback, qos)

	token := c.client.Subscribe(topic, qos, wrappedCallback)
	if token == nil {
		return fmt.Errorf("nil token while subscribing (%v, %v)", topic, qos)
	}

	if !token.WaitTimeout(c.clientOptions.ConnectTimeout) {
		return fmt.Errorf("subscribe timed out after %v", c.clientOptions.ConnectTimeout)
	}

	return token.Error()
}

func (c *Client) Unsubscribe(topic string) error {
	log.Printf("unsubscribing from %v", topic)

	token := c.client.Unsubscribe(topic)
	if token == nil {
		return fmt.Errorf("nil token while unsubscribing (%v)", topic)
	}

	if !token.WaitTimeout(c.clientOptions.ConnectTimeout) {
		return fmt.Errorf("unsubscribe timed out after %v", c.clientOptions.ConnectTimeout)
	}

	return token.Error()
}

func (c *Client) Disconnect() error {
	log.Printf("disconnecting")

	if c.connectToken == nil {
		return fmt.Errorf("nil token while disconnecting")
	}

	c.client.Disconnect(1000)

	return nil
}

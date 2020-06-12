package paho_mqtt_client

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/initialed85/mqtt_things/pkg/mqtt_common"
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
	clientOptions *mqtt.ClientOptions
	connectToken  mqtt.Token
	client        mqtt.Client
}

func New(host, username, password string) (c *Client) {
	if TestMode {
		host = TestHost
	}

	clientID := "paho_"
	uuid4, err := uuid.NewRandom()
	if err != nil {
		clientID += fmt.Sprintf("unknown_%+v", time.Now().UnixNano())
	} else {
		clientID += uuid4.String()
	}

	c = &Client{}

	c.clientOptions = mqtt.NewClientOptions()
	c.clientOptions.AddBroker(fmt.Sprintf("tcp://%v:1883", host))
	c.clientOptions.SetClientID(string(clientID))
	c.clientOptions.SetUsername(username)
	c.clientOptions.SetPassword(password)
	c.clientOptions.SetKeepAlive(time.Second * 2)
	c.clientOptions.SetPingTimeout(time.Second * 2)
	c.clientOptions.SetConnectTimeout(time.Second * 2)
	c.clientOptions.SetWriteTimeout(time.Second * 2)
	c.clientOptions.SetMaxReconnectInterval(time.Second * 4)
	c.clientOptions.SetAutoReconnect(true)
	c.clientOptions.SetResumeSubs(true)

	log.Printf("created %+v", c.clientOptions)

	return c
}

func (c *Client) Connect() error {
	log.Printf("connecting")

	c.client = mqtt.NewClient(c.clientOptions)

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

func (c *Client) Subscribe(topic string, qos byte, callback func(message mqtt_common.Message)) error {
	wrappedCallback := func(client mqtt.Client, message mqtt.Message) {
		callback(mqtt_common.Message{
			Received:  time.Now(),
			Topic:     message.Topic(),
			MessageID: message.MessageID(),
			Payload:   string(message.Payload()),
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

	c.client = nil

	return nil
}

func (c *Client) Reconnect() error {
	_ = c.Disconnect()

	err := c.Connect()
	if err != nil {
		log.Printf("caught %+v trying to reconnect", err)
	}

	return err
}

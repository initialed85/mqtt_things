package mqtt_client

import (
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"time"
)

var PahoTestMode = false
var PahoTestHost string

func EnablePahoTestMode(host string) {
	PahoTestMode = true
	PahoTestHost = host
}

type PahoClient struct {
	clientOptions *paho.ClientOptions
	connectToken  paho.Token
	client        paho.Client
	errorHandler  func(Client, error)
}

func NewPahoClient(host, username, password string, errorHandler func(Client, error)) (c *PahoClient) {
	if PahoTestMode {
		host = PahoTestHost
	}

	clientID := getClientID("paho")

	c = &PahoClient{
		errorHandler: errorHandler,
	}

	c.clientOptions = paho.NewClientOptions()
	c.clientOptions.AddBroker(fmt.Sprintf("tcp://%v:1883", host))
	c.clientOptions.SetClientID(clientID)
	c.clientOptions.SetUsername(username)
	c.clientOptions.SetPassword(password)
	c.clientOptions.SetKeepAlive(time.Second * 2)
	c.clientOptions.SetPingTimeout(time.Second * 2)
	c.clientOptions.SetConnectTimeout(time.Second * 2)
	c.clientOptions.SetWriteTimeout(time.Second * 2)
	c.clientOptions.SetMaxReconnectInterval(time.Second * 4)
	c.clientOptions.SetAutoReconnect(true)
	c.clientOptions.SetResumeSubs(true)
	c.clientOptions.OnConnectionLost = func(client paho.Client, err error) {
		go func() {
			time.Sleep(time.Second)

			c.errorHandler(c, err)
		}()
	}

	return c
}

func (c *PahoClient) Connect() error {
	c.client = paho.NewClient(c.clientOptions)

	c.connectToken = c.client.Connect()
	if c.connectToken == nil {
		return fmt.Errorf("nil token while connecting")
	}

	if !c.connectToken.WaitTimeout(c.clientOptions.ConnectTimeout) {
		return fmt.Errorf("connection timed out after %v", c.clientOptions.ConnectTimeout)
	}

	return c.connectToken.Error()
}

func (c *PahoClient) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}

	token := c.client.Publish(topic, qos, retained, payload)
	if token == nil {
		return fmt.Errorf("nil token while publishing (%v, %v, %v, %v)", topic, qos, retained, payload)
	}

	if !token.WaitTimeout(c.clientOptions.WriteTimeout) {
		return fmt.Errorf("publish timed out after %v", c.clientOptions.ConnectTimeout)
	}

	return token.Error()
}

func (c *PahoClient) Subscribe(topic string, qos byte, callback func(message Message)) error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}

	wrappedCallback := func(client paho.Client, message paho.Message) {
		callback(Message{
			Received:  time.Now(),
			Topic:     message.Topic(),
			MessageID: message.MessageID(),
			Payload:   string(message.Payload()),
		})
	}

	token := c.client.Subscribe(topic, qos, wrappedCallback)
	if token == nil {
		return fmt.Errorf("nil token while subscribing (%v, %v)", topic, qos)
	}

	if !token.WaitTimeout(c.clientOptions.ConnectTimeout) {
		return fmt.Errorf("subscribe timed out after %v", c.clientOptions.ConnectTimeout)
	}

	return token.Error()
}

func (c *PahoClient) Unsubscribe(topic string) error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}

	token := c.client.Unsubscribe(topic)
	if token == nil {
		return fmt.Errorf("nil token while unsubscribing (%v)", topic)
	}

	if !token.WaitTimeout(c.clientOptions.ConnectTimeout) {
		return fmt.Errorf("unsubscribe timed out after %v", c.clientOptions.ConnectTimeout)
	}

	return token.Error()
}

func (c *PahoClient) Disconnect() error {
	if c.client == nil {
		return fmt.Errorf("client is nil (probably not connected)")
	}

	if c.connectToken == nil {
		return fmt.Errorf("nil token while disconnecting")
	}

	c.client.Disconnect(0)

	c.client = nil

	return nil
}

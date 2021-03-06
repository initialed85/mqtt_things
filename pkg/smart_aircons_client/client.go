package smart_aircons_client

import (
	"fmt"
	"log"
	"strings"

	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
)

type Client struct {
	router *Router
	model  *Model

	topicPrefix string
	host        string
	codes       string
	sendIR      func(string, string) error
	publish     func(topic string, qos byte, retained bool, payload interface{}) error
}

func NewClient(
	topicPrefix string,
	host string,
	codes string,
	sendIR func(string, string) error,
	publish func(topic string, qos byte, retained bool, payload interface{}) error,
) *Client {
	c := Client{
		topicPrefix: topicPrefix,
		host:        host,
		codes:       codes,
		sendIR:      sendIR,
		publish:     publish,
	}

	// model sets device state
	c.model = NewModel(
		c.setState,
	)

	// router receives messages from broker and calls model
	c.router = NewRouter(
		strings.TrimRight(topicPrefix, "/")+"/",
		c.model.SetOn,
		c.model.SetMode,
		c.model.SetTemperature,
	)

	// operate as normal by default
	c.DisableRestoreMode()

	return &c
}

func (c *Client) setState(on bool, mode string, temperature int64) error {
	var code string
	var err error

	log.Printf("setState(on=%#+v, mode=%#+v, temperature=%#+v)", on, mode, temperature)

	code, err = GetCode(c.codes, on, mode, temperature)
	if err != nil {
		return fmt.Errorf("cannot call setState(%#+v, %#+v, %#+v) because: %v", on, mode, temperature, err)
	}

	err = c.sendIR(c.host, code)
	if err != nil {
		return fmt.Errorf("cannot call setState(%#+v, %#+v, %#+v) because: %v", on, mode, temperature, err)
	}

	return nil
}

func (c *Client) EnableRestoreMode() {
	// honour messages on /get topics
	c.router.EnableGetCallbacks()
}

func (c *Client) DisableRestoreMode() {
	// ignore messages on /get topics
	c.router.DisableGetCallbacks()
}

func (c *Client) Handle(message mqtt.Message) {
	outgoingMessage, ok := c.router.Handle(message)
	if !ok {
		return
	}

	err := c.publish(outgoingMessage.Topic, mqtt.ExactlyOnce, true, outgoingMessage.Payload)
	if err != nil {
		log.Printf("warning: failed to publish %#+v because: %v", outgoingMessage, err)
	}
}

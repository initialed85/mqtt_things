package smart_aircons_client

import (
	"fmt"
	"log"
	"strings"
	"sync"

	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
)

type Client struct {
	mu *sync.Mutex

	router *Router
	model  *Model

	topicPrefix string
	host        string
	codes       string
	sendIR      func(string, []byte) error
	publish     func(topic string, qos byte, retained bool, payload interface{}, quiet ...bool) error
}

func NewClient(
	topicPrefix string,
	host string,
	codes string,
	sendIR func(string, []byte) error,
	publish func(topic string, qos byte, retained bool, payload interface{}, quiet ...bool) error,
) *Client {
	c := Client{
		mu:          new(sync.Mutex),
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
	var code []byte
	var err error

	log.Printf("setState(on=%#+v, mode=%#+v, temperature=%#+v)", on, mode, temperature)

	if !c.router.IsGetCallbacks() {
		if on && !c.model.on || ((mode == "cool" || mode == "heat") && mode != c.model.mode) {
			code, err = GetCode(c.codes, on, "fan_only", temperature)
			if err != nil {
				return fmt.Errorf("cannot call setState(%#+v, %#+v, %#+v) (on way to setState(%#+v, %#+v, %#+v)) because: %v", on, "fan_only", temperature, on, mode, temperature, err)
			}

			err = c.sendIR(c.host, code)
			if err != nil {
				return fmt.Errorf("cannot call setState(%#+v, %#+v, %#+v) (on way to setState(%#+v, %#+v, %#+v)) because: %v", on, "fan_only", temperature, on, mode, temperature, err)
			}
		}
	}

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

	c.mu.Lock()
	err := c.publish(outgoingMessage.Topic, mqtt.ExactlyOnce, true, outgoingMessage.Payload)
	c.mu.Unlock()

	if err != nil {
		log.Printf("warning: failed to publish %#+v because: %v", outgoingMessage, err)
	}
}

func (c *Client) Update() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	rawOn, mode, temperature := c.model.GetState()

	on := "OFF"
	if rawOn {
		on = "ON"
	}

	topicPrefix := strings.TrimRight(c.topicPrefix, "/") + "/"

	outgoingMessages := []mqtt.Message{
		{
			Topic:   fmt.Sprintf("%v%v/%v", topicPrefix, topicOnInfix, topicGetSuffix),
			Payload: fmt.Sprintf("%v", on),
		},
		{
			Topic:   fmt.Sprintf("%v%v/%v", topicPrefix, topicModeInfix, topicGetSuffix),
			Payload: fmt.Sprintf("%v", mode),
		},
		{
			Topic:   fmt.Sprintf("%v%v/%v", topicPrefix, topicTemperatureInfix, topicGetSuffix),
			Payload: fmt.Sprintf("%v", temperature),
		},
	}

	for _, outgoingMessage := range outgoingMessages {
		err := c.publish(outgoingMessage.Topic, mqtt.ExactlyOnce, true, outgoingMessage.Payload, true)
		if err != nil {
			return fmt.Errorf("failed to publish %#+v because: %v", outgoingMessage, err)
		}
	}

	return nil
}

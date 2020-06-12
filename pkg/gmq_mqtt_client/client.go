package gmq_mqtt_client

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/initialed85/mqtt_things/pkg/mqtt_common"
	"github.com/yosssi/gmq/mqtt/client"
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

type Subscription struct {
	topic    string
	qos      byte
	callback func(message mqtt_common.Message)
}

type Client struct {
	connectOptions                           client.ConnectOptions
	options                                  client.Options
	client                                   *client.Client
	subscriptionByTopicMutex, reconnectMutex sync.RWMutex
	subscriptionByTopic                      map[string]Subscription
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

	c = &Client{
		subscriptionByTopic: make(map[string]Subscription),
	}

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
			c.reconnectMutex.Lock()
			defer c.reconnectMutex.Unlock()
			log.Printf("error handler fired because %+v; backing off and reconnecting...", err)

			time.Sleep(time.Second * 5)

			err = c.Reconnect()
			if err != nil {
				panic(fmt.Errorf("reconnected after backoff failed because %+v; giving up", err))
			}
		},
	}

	log.Printf("created %+v, %+v", c.connectOptions, c.options)

	return c
}

func (c *Client) Connect() error {
	log.Printf("connecting")

	c.client = client.New(&c.options)

	err := c.client.Connect(&c.connectOptions)
	if err == nil {
		log.Printf("connected")
	}

	return err
}

func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	// TODO: this is a bit kludgy- block anyone trying to publish while we're reconnecting
	c.reconnectMutex.Lock()
	defer c.reconnectMutex.Unlock()

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
	// TODO: no reconnectMutex here because we need to subscribe during reconnect- it's fine it's fine

	wrappedCallback := func(topicName, message []byte) {
		callback(mqtt_common.Message{
			Received:  time.Now(),
			Topic:     string(topicName),
			MessageID: 0,
			Payload:   string(message),
		})
	}

	log.Printf("subscribing to %v callback %p and qos %v", topic, callback, qos)

	err := c.client.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			{
				TopicFilter: []byte(topic),
				QoS:         qos,
				Handler:     wrappedCallback,
			},
		},
	})

	if err == nil {
		c.subscriptionByTopicMutex.Lock()
		c.subscriptionByTopic[topic] = Subscription{
			topic,
			qos,
			callback,
		}
		c.subscriptionByTopicMutex.Unlock()
	}

	return err
}

func (c *Client) Unsubscribe(topic string) error {
	log.Printf("unsubscribing from %v", topic)

	err := c.client.Unsubscribe(&client.UnsubscribeOptions{
		TopicFilters: [][]byte{
			[]byte(topic),
		},
	})

	if err == nil {
		c.subscriptionByTopicMutex.Lock()
		delete(c.subscriptionByTopic, topic)
		c.subscriptionByTopicMutex.Unlock()
	}

	return err
}

func (c *Client) Disconnect() error {
	log.Printf("disconnecting")

	err := c.client.Disconnect()

	c.client = nil

	return err
}

func (c *Client) Reconnect() error {
	_ = c.Disconnect()

	err := c.Connect()
	if err != nil {
		return err
	}

	c.subscriptionByTopicMutex.Lock()
	for _, subscription := range c.subscriptionByTopic {
		err = c.Subscribe(subscription.topic, subscription.qos, subscription.callback)
		if err != nil {
			return err
		}
	}
	c.subscriptionByTopicMutex.Unlock()

	return nil
}

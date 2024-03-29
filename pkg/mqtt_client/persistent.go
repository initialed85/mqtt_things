package mqtt_client

import (
	"log"
	"sync"
	"time"
)

type Subscription struct {
	topic    string
	qos      byte
	callback func(message Message)
}

type PersistentClient struct {
	client                Client
	subscriptionByTopicMu sync.Mutex
	subscriptionByTopic   map[string]Subscription
	errorBeingHandledMu   sync.Mutex
	errorBeingHandled     bool
}

func NewPersistentClient() *PersistentClient {
	return &PersistentClient{
		subscriptionByTopic: make(map[string]Subscription),
	}
}

func (c *PersistentClient) SetClient(client Client) {
	c.client = client
}

func (c *PersistentClient) unsubscribeAll() {
	c.subscriptionByTopicMu.Lock()
	defer c.subscriptionByTopicMu.Unlock()

	for topic := range c.subscriptionByTopic {
		_ = c.client.Unsubscribe(topic)
	}
}

func (c *PersistentClient) resubscribeAll() error {
	c.subscriptionByTopicMu.Lock()
	defer c.subscriptionByTopicMu.Unlock()

	for _, subscription := range c.subscriptionByTopic {
		err := c.client.Subscribe(subscription.topic, subscription.qos, subscription.callback)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *PersistentClient) isErrorBeingHandled() bool {
	c.errorBeingHandledMu.Lock()
	defer c.errorBeingHandledMu.Unlock()

	return c.errorBeingHandled
}

func (c *PersistentClient) waitWhileErrorBeingHandled() {
	for {
		if c.isErrorBeingHandled() {
			log.Printf("error being handled; waiting...")

			time.Sleep(time.Second)

			continue
		}

		break
	}
}

func (c *PersistentClient) HandleError(client Client, err error) {
	log.Printf("handling %+v for %+v...", err, client)

	c.errorBeingHandledMu.Lock()
	if c.errorBeingHandled {
		c.errorBeingHandledMu.Unlock()
		log.Printf("another error is already in progress, ignoring this error.")
		return
	} else {
		c.errorBeingHandled = true
		c.errorBeingHandledMu.Unlock()
	}

	for {
		log.Printf("unsubscribing from all topics...")
		c.unsubscribeAll()

		log.Printf("disconnecting...")
		_ = c.Disconnect()

		log.Printf("sleeping...")
		time.Sleep(time.Second)

		log.Printf("reconnecting...")
		err = c.Connect()
		if err != nil {
			log.Printf("reconnect failed because %+v; trying again...", err)

			continue
		}

		log.Printf("resubscribing to all topics...")
		err = c.resubscribeAll()
		if err != nil {
			log.Printf("resubscribe failed because %+v; trying again...", err)

			continue
		}

		log.Printf("reconnected.")
		break
	}

	c.errorBeingHandledMu.Lock()
	c.errorBeingHandled = false
	c.errorBeingHandledMu.Unlock()
}

func (c *PersistentClient) Connect() error {
	log.Printf("connecting...")

	err := c.client.Connect()

	if err != nil {
		log.Printf("failed to connect because %+v", err)
	} else {
		log.Printf("connected")
	}

	return err
}

func (c *PersistentClient) Publish(topic string, qos byte, retained bool, payload interface{}, quiet ...bool) error {
	c.waitWhileErrorBeingHandled()

	actualQuiet := false
	if len(quiet) > 0 {
		actualQuiet = quiet[0]
	}

	if !actualQuiet {
		log.Printf("publishing %+v to %+v", payload, topic)
	}

	err := c.client.Publish(topic, qos, retained, payload)
	if err != nil {
		if !actualQuiet {
			log.Printf("failed to publish because %+v", err)
		}

		return err
	}

	return nil
}

func (c *PersistentClient) Subscribe(topic string, qos byte, callback func(message Message)) error {
	c.waitWhileErrorBeingHandled()

	log.Printf("subscribing to %+v with %p", topic, callback)

	err := c.client.Subscribe(topic, qos, callback)

	if err != nil {
		log.Printf("failed to subscribe because %+v", err)
	} else {
		c.subscriptionByTopicMu.Lock()
		c.subscriptionByTopic[topic] = Subscription{
			topic:    topic,
			qos:      qos,
			callback: callback,
		}
		c.subscriptionByTopicMu.Unlock()

		log.Printf("subscribed")
	}

	return err
}

func (c *PersistentClient) Unsubscribe(topic string) error {
	log.Printf("unsubscribing from %+v", topic)

	err := c.client.Unsubscribe(topic)

	if err != nil {
		log.Printf("failed to unsubscribe because %+v", err)
	} else {
		log.Printf("unsubscribed")
	}

	return err
}

func (c *PersistentClient) Disconnect() error {
	log.Printf("disconnecting...")

	_ = c.client.Disconnect()

	log.Printf("disconnected")

	return nil
}

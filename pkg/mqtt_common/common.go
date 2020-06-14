package mqtt_common

import "time"

const (
	AtMostOnce  = byte(0)
	AtLeastOnce = byte(1)
	ExactlyOnce = byte(2)
)

type Message struct {
	Received  time.Time
	Topic     string
	MessageID uint16
	Payload   string
}

type Client interface {
	Connect() error
	Publish(topic string, qos byte, retained bool, payload interface{}) error
	Subscribe(topic string, qos byte, callback func(message Message)) error
	Unsubscribe(topic string) error
	Disconnect() error
}

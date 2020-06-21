package mqtt_client

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



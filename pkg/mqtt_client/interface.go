package mqtt_client

type Client interface {
	Connect() error
	Publish(topic string, qos byte, retained bool, payload interface{}, quiet ...bool) error
	Subscribe(topic string, qos byte, callback func(message Message)) error
	Unsubscribe(topic string) error
	Disconnect() error
}

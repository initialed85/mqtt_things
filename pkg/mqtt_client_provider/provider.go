package mqtt_client_provider

import (
	"github.com/initialed85/mqtt_things/pkg/gmq_mqtt_client"
	"github.com/initialed85/mqtt_things/pkg/mqtt_common"
	"github.com/initialed85/mqtt_things/pkg/paho_mqtt_client"
	"os"
	"strings"
)

func GetMQTTClient(host, username, password string) (client mqtt_common.Client) {
	useGMQ := false
	usePaho := false
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if pair[0] == "MQTT_CLIENT_PROVIDER" {
			if strings.ToLower(pair[1]) == "gmq" {
				useGMQ = true
			} else if strings.ToLower(pair[1]) == "paho" {
				usePaho = true
			}
		}
	}

	if usePaho {
		client = paho_mqtt_client.New(host, username, password)
	} else if useGMQ {
		client = gmq_mqtt_client.New(host, username, password)
	} else {
		client = gmq_mqtt_client.New(host, username, password) // use GMQ as default
	}

	return client
}

package mqtt_client_provider

import (
	"github.com/initialed85/mqtt_things/pkg/gmq_mqtt_client"
	"github.com/initialed85/mqtt_things/pkg/libmqtt_mqtt_client"
	"github.com/initialed85/mqtt_things/pkg/mqtt_common"
	"github.com/initialed85/mqtt_things/pkg/paho_mqtt_client"
	"log"
	"os"
	"strings"
)

func GetMQTTClient(host, username, password string) (client mqtt_common.Client) {
	useGMQ := false
	usePaho := false
	useLibMQTT := false
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if pair[0] == "MQTT_CLIENT_PROVIDER" {
			if strings.ToLower(pair[1]) == "gmq" {
				useGMQ = true
			} else if strings.ToLower(pair[1]) == "paho" {
				usePaho = true
			} else if strings.ToLower(pair[1]) == "libmqtt" {
				useLibMQTT = true
			}
		}
	}

	if usePaho {
		log.Printf("using Paho")
		client = paho_mqtt_client.New(host, username, password)
	} else if useGMQ {
		log.Printf("using GMQ")
		client = gmq_mqtt_client.New(host, username, password)
	} else if useLibMQTT {
		log.Printf("using LibMQTT")
		client = libmqtt_mqtt_client.New(host, username, password)
	} else {
		log.Printf("using GMQ (because it's the default)")
		client = gmq_mqtt_client.New(host, username, password)
	}

	return client
}

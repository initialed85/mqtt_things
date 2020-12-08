package mqtt_client

import (
	"log"
	"os"
	"strings"
)

func GetPahoClient(host, username, password string, errorHandler func(Client, error)) (client Client) {
	return NewPahoClient(host, username, password, errorHandler)
}

func GetGMQClient(host, username, password string, errorHandler func(Client, error)) (client Client) {
	return NewGMQClient(host, username, password, errorHandler)
}

func GetLibMQTTClient(host, username, password string, errorHandler func(Client, error)) (client Client) {

	return NewLibMQTTClient(host, username, password, errorHandler)
}

func GetGlueClient(host, username, password string, errorHandler func(Client, error)) (client Client) {
	client, err := NewGlueClient(host, username, password, errorHandler)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func GetMQTTClient(host, username, password string) (client Client) {
	useGMQ := false
	usePaho := false
	useLibMQTT := false
	useGlue := false
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if pair[0] == "MQTT_CLIENT_PROVIDER" {
			if strings.ToLower(pair[1]) == "gmq" {
				useGMQ = true
			} else if strings.ToLower(pair[1]) == "paho" {
				usePaho = true
			} else if strings.ToLower(pair[1]) == "libmqtt" {
				useLibMQTT = true
			} else if strings.ToLower(pair[1]) == "glue" {
				useGlue = true
			}
		}
	}

	p := NewPersistentClient()

	if usePaho {
		log.Printf("using Paho")
		client = GetPahoClient(host, username, password, p.HandleError)
	} else if useGMQ {
		log.Printf("using GMQ")
		client = GetGMQClient(host, username, password, p.HandleError)
	} else if useLibMQTT {
		log.Printf("using LibMQTT")
		client = GetLibMQTTClient(host, username, password, p.HandleError)
	} else if useGlue {
		log.Printf("using Glue")
		client = GetGlueClient(host, username, password, p.HandleError)
	} else {
		log.Printf("using Glue (because it's the default)")
		client = GetGlueClient(host, username, password, p.HandleError)
	}

	p.SetClient(client)

	return p
}

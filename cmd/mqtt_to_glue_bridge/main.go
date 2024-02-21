package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/initialed85/glue/pkg/endpoint"
	"github.com/initialed85/glue/pkg/topics"

	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
)

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	mqttClient := mqtt.NewPersistentClient()
	gmqClient := mqtt.GetGMQClient(*hostPtr, *usernamePtr, *passwordPtr, mqttClient.HandleError)
	mqttClient.SetClient(gmqClient)

	err := mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	glueClient, err := endpoint.NewManagerSimple()
	if err != nil {
		log.Fatal(err)
	}

	glueClient.Start()

	c := make(chan os.Signal, 16)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err = mqttClient.Disconnect()
		if err != nil {
			log.Fatal(err)
		}

		glueClient.Stop()

		os.Exit(0)
	}()

	// feed stuff from mqtt into glue
	err = mqttClient.Subscribe(
		"+/#",
		mqtt.ExactlyOnce,
		func(message mqtt.Message) {
			err = glueClient.Publish(
				message.Topic,
				"__mqtt_to_glue_bridge__",
				time.Second,
				[]byte(message.Payload),
			)
			if err != nil {
				log.Printf("warning: %v", err)
			}
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// feed stuff from glue into mqtt
	err = glueClient.Subscribe(
		"#",
		"", // ignored for a wildcard subscription
		func(message topics.Message) {
			if message.TopicType == "__mqtt_to_glue_bridge__" {
				return
			}

			err = mqttClient.Publish(
				message.TopicName,
				mqtt.ExactlyOnce,
				false,
				string(message.Payload),
			)
			if err != nil {
				log.Printf("warning: %v", err)
			}
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(time.Second)
	}
}

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	glueClient := mqtt.GetGlueClient(
		"",
		"",
		"",
		func(client mqtt.Client, err error) {},
	)

	err = glueClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err = mqttClient.Disconnect()
		if err != nil {
			log.Fatal(err)
		}

		err = glueClient.Disconnect()
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	// feed stuff from mqtt into glue
	err = mqttClient.Subscribe(
		"#",
		mqtt.ExactlyOnce,
		func(message mqtt.Message) {
			err = glueClient.Publish(
				message.Topic,
				mqtt.ExactlyOnce,
				false,
				message.Payload,
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
		mqtt.ExactlyOnce,
		func(message mqtt.Message) {
			err = mqttClient.Publish(
				message.Topic,
				mqtt.ExactlyOnce,
				false,
				message.Payload,
			)
			if err != nil {
				log.Printf("warning: %v", err)
			}
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: the loops oh god the loops

	for {
		time.Sleep(time.Second)
	}
}

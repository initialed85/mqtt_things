package main

import (
	"flag"
	"github.com/initialed85/mqtt_things/pkg/mqtt_client_provider"
	"github.com/initialed85/mqtt_things/pkg/mqtt_common"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	modePtr := flag.String("mode", "sub", "pub / sub")
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")
	topicPtr := flag.String("topic", "", "mqtt topic")
	payloadPtr := flag.String("payload", "", "payload to publish")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *modePtr != "pub" && *modePtr != "sub" {
		log.Fatal("mode flag was neither pub nor sub")
	}

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if *topicPtr == "" {
		log.Fatal("topic flag empty")
	}

	mqttClient := mqtt_client_provider.GetMQTTClient(*hostPtr, *usernamePtr, *passwordPtr)
	err := mqttClient.Connect()
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

		os.Exit(0)
	}()

	if *modePtr == "pub" {
		if *payloadPtr == "" {
			log.Fatal("message flag empty")
		}

		err = mqttClient.Publish(*topicPtr, mqtt_common.ExactlyOnce, false, *payloadPtr)
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Second)
	} else if *modePtr == "sub" {
		callback := func(message mqtt_common.Message) {
			log.Printf("%+v\n", message)
		}

		err = mqttClient.Subscribe(*topicPtr, mqtt_common.ExactlyOnce, callback)
		if err != nil {
			log.Fatal(err)
		}

		select {}
	}

	err = mqttClient.Disconnect()
	if err != nil {
		log.Fatal(err)
	}
}

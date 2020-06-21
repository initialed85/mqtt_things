package main

import (
	"flag"

	"github.com/initialed85/mqtt_things/pkg/mqtt_action_router"
	"github.com/initialed85/mqtt_things/pkg/relays_client"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")
	portPtr := flag.String("port", "", "serial port")
	relayPtr := flag.Int64("relay", -1, "relay number")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if *portPtr == "" {
		log.Fatal("port flag empty")
	}

	if *relayPtr == -1 {
		log.Fatal("relay flag empty")
	}

	relaysClient, err := relays_client.New(*portPtr, 9600)
	if err != nil {
		log.Fatal(err)
	}

	actionable := relays_client.Actionable{
		Client: relaysClient,
	}

	mqttClient := mqtt.GetMQTTClient(*hostPtr, *usernamePtr, *passwordPtr)
	err = mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	actionRouter := mqtt_action_router.New(
		mqttClient,
		time.Millisecond*10,
		false,
	)

	relayNumbers := []int64{*relayPtr}

	for _, relayNumber := range relayNumbers {
		err = actionRouter.AddAction(
			"home/inside/heater/state/set",
			relays_client.Arguments{Relay: relayNumber},
			actionable.On,
			actionable.Off,
			mqtt_action_router.Off,
			"home/inside/heater/state/get",
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err = actionRouter.RemoveAllActions()
		if err != nil {
			log.Print(err)
		}

		err = mqttClient.Disconnect()
		if err != nil {
			log.Print(err)
		}

		os.Exit(0)
	}()

	ticker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <-ticker.C:
		}
	}
}

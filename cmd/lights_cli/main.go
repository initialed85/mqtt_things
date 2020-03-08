package main

import (
	"flag"
	"fmt"
	"github.com/initialed85/mqtt_things/pkg/lights_client"
	"github.com/initialed85/mqtt_things/pkg/mqtt_action_router"
	"github.com/initialed85/mqtt_things/pkg/mqtt_client"
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
	bridgeHost := flag.String("bridgeHost", "", "hue bridge host")
	apiKeyPtr := flag.String("apiKey", "", "hue api key")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if *apiKeyPtr == "" {
		log.Fatal("apiKey flag empty")
	}

	if *bridgeHost == "" {
		log.Fatal("bridgeHost flag empty")
	}

	lightsClient := lights_client.New(*bridgeHost, "smart_home", *apiKeyPtr)

	actionable := lights_client.Actionable{
		Client: lightsClient,
	}

	mqttClient := mqtt_client.New(*hostPtr, *usernamePtr, *passwordPtr)
	err := mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	actionRouter := mqtt_action_router.New(
		mqttClient,
		time.Millisecond*10,
		true,
	)

	lights, err := lightsClient.GetLights()
	if err != nil {
		log.Fatal(err)
	}

	for _, light := range lights {
		err = actionRouter.AddAction(
			fmt.Sprintf("home/inside/lights/globe/%v/state/set", light.Name),
			lights_client.Arguments{Name: light.Name},
			actionable.On,
			actionable.Off,
			mqtt_action_router.Off,
			fmt.Sprintf("home/inside/lights/globe/%v/state/get", light.Name),
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
			lights, err := lightsClient.GetLights()
			if err != nil {
				log.Fatal(err)
			}

			for _, light := range lights {
				err := mqttClient.Publish(
					fmt.Sprintf("home/inside/lights/globe/%v/state/get", light.Name),
					mqtt_client.ExactlyOnce,
					true,
					fmt.Sprintf("%v", light.State),
				)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

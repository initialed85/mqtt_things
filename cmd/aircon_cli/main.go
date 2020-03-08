package main

import (
	"flag"
	"github.com/initialed85/mqtt_things/pkg/aircon_client"
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
	airconHostPtr := flag.String("aircon", "", "aircon host")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if *airconHostPtr == "" {
		log.Fatal("aircon flag empty")
	}

	airconClient := aircon_client.New(*airconHostPtr)

	actionableAircon := aircon_client.Actionable{
		Client: airconClient,
	}

	mqttClient := mqtt_client.New(*hostPtr, *usernamePtr, *passwordPtr)
	err := mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	actionRouter := mqtt_action_router.New(
		mqttClient,
		time.Millisecond*10,
		false,
	)

	err = actionRouter.AddAction(
		"home/inside/aircon/state/set",
		aircon_client.Arguments{},
		actionableAircon.On,
		actionableAircon.Off,
		mqtt_action_router.Off,
		"home/inside/aircon/state/get",
	)
	if err != nil {
		log.Fatal(err)
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

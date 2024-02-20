package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/initialed85/mqtt_things/pkg/mqtt_action_router"
	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
	"github.com/initialed85/mqtt_things/pkg/relays_client"
)

type flagArrayInt64 []int64

func (f *flagArrayInt64) String() string {
	return fmt.Sprintf("%d", f)
}

func (f *flagArrayInt64) Set(value string) error {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}

	*f = append(*f, i)

	return nil
}

var relaysPtr flagArrayInt64

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")
	portPtr := flag.String("port", "", "serial port")
	flag.Var(&relaysPtr, "relay", "a relay to map to")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if *portPtr == "" {
		log.Fatal("port flag empty")
	}

	relayNumbers := make([]int64, 0)
	for _, v := range relaysPtr {
		relayNumbers = append(relayNumbers, v)
	}

	if len(relayNumbers) == 0 {
		log.Fatal("no relays specified")
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

	for _, relayNumber := range relayNumbers {
		err = actionRouter.AddAction(
			fmt.Sprintf("home/outside/sprinklers/bank/%v/state/set", relayNumber),
			relays_client.Arguments{Relay: relayNumber},
			actionable.On,
			actionable.Off,
			mqtt_action_router.Off,
			fmt.Sprintf("home/outside/sprinklers/bank/%v/state/get", relayNumber),
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	c := make(chan os.Signal, 16)
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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
	"github.com/initialed85/mqtt_things/pkg/smart_aircons_client"
)

type flagArrayString []string

func (f *flagArrayString) String() string {
	return strings.Join(*f, ", ")
}

func (f *flagArrayString) Set(value string) error {
	*f = append(*f, value)

	return nil
}

const (
	overallTopicPrefix = "home/inside/smart-aircons"
)

var (
	airconHosts      flagArrayString
	airconNames      flagArrayString
	airconCodesNames flagArrayString
)

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")
	flag.Var(&airconHosts, "airconHost", "a host for an aircon")
	flag.Var(&airconNames, "airconName", "a name for an aircon")
	flag.Var(&airconCodesNames, "airconCodesName", "a codes name for an aircon")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if len(airconHosts) == 0 {
		log.Fatal("no -airconHost flags specified")
	}

	if len(airconNames) == 0 {
		log.Fatal("no -airconName flags specified")
	}

	if len(airconCodesNames) == 0 {
		log.Fatal("no -airconCodesName flags specified")
	}

	if len(airconHosts) != len(airconNames) || len(airconNames) != len(airconCodesNames) {
		log.Fatal("unbalanced mixture of -airconName and -airconName and airconCodesName flags")
	}

	var err error

	mqttClient := mqtt.GetMQTTClient(*hostPtr, *usernamePtr, *passwordPtr)
	err = mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second)

	clients := make([]*smart_aircons_client.Client, 0)
	for i, airconHost := range airconHosts {
		airconName := airconNames[i]
		airconCodesName := airconCodesNames[i]

		topicPrefix := fmt.Sprintf("%v/%v", overallTopicPrefix, airconName)

		client := smart_aircons_client.NewClient(
			topicPrefix,
			airconHost,
			airconCodesName,
			smart_aircons_client.SendIR,
			mqttClient.Publish,
		)
		client.EnableRestoreMode()

		err = mqttClient.Subscribe(
			fmt.Sprintf("%v/#", topicPrefix),
			mqtt.ExactlyOnce,
			func(message mqtt.Message) {
				go client.Handle(message)
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		clients = append(clients, client)
	}

	time.Sleep(time.Second)

	for _, client := range clients {
		client.DisableRestoreMode()
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err = mqttClient.Disconnect()
		if err != nil {
			log.Print(err)
		}

		wg.Done()
	}()

	wg.Wait()
}

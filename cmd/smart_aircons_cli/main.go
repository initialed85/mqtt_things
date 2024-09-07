package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
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
			func(hostOrMac string, code []byte) error {
				if len(code) < 2 {
					return fmt.Errorf("code %#+v not at least 2 bytes long", code)
				}

				switch {
				case bytes.Equal(code[0:2], []byte{0x49, 0x52}):
					return smart_aircons_client.ZmoteSendIR(hostOrMac, code)
				case bytes.Equal(code[0:2], []byte{0x73, 0x65}), bytes.Equal(code[0:2], []byte{0x26, 0x0}):
					return smart_aircons_client.BroadlinkSendIR(hostOrMac, code)
				default:
				}

				return fmt.Errorf(
					"expected %#+v to be of have prefix []byte{0x73, 0x65} or []byte{0x49, 0x52} (for zmote) or []byte{0x26, 0x0} (for broadlink)",
					code,
				)
			},
			mqttClient.Publish,
		)

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

	for _, client := range clients {
		client.EnableRestoreMode()
	}

	time.Sleep(time.Second)

	for _, client := range clients {
		client.DisableRestoreMode()
	}

	time.Sleep(time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		t := time.NewTicker(time.Second * 1)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				for _, client := range clients {
					err = client.Update()
					if err != nil {
						log.Fatalf("failed to update %#+v; err: %v", client, err)
					}
				}
			}
		}
	}()

	c := make(chan os.Signal, 16)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err = mqttClient.Disconnect()
		if err != nil {
			log.Print(err)
		}

		cancel()
	}()

	<-ctx.Done()
}

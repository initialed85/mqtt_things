package main

import (
	"flag"
	"fmt"
	"github.com/initialed85/mqtt_things/pkg/mqtt_action_router"
	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
	"github.com/initialed85/mqtt_things/pkg/switches_client"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type flagArrayString []string

func (f *flagArrayString) String() string {
	return fmt.Sprintf("%d", f)
}

func (f *flagArrayString) Set(value string) error {
	*f = append(*f, value)

	return nil
}

var (
	names flagArrayString
	hosts flagArrayString
)

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")
	flag.Var(&hosts, "switchHost", "a host for a switch")
	flag.Var(&names, "switchName", "a name for a switch")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if len(hosts) == 0 {
		log.Fatal("no -switchHost flags specified")
	}

	if len(names) == 0 {
		log.Fatal("no -switchName flags specified")
	}

	if len(hosts) != len(names) {
		log.Fatal("unbalanced mixture of -switchHost and -switchName flags")
	}

	hostsAndName := make([]switches_client.HostAndName, 0)
	for i, name := range names {
		host := hosts[i]
		hostsAndName = append(hostsAndName, switches_client.HostAndName{
			Host: host,
			Name: name,
		})
	}

	switchesClient := switches_client.New(hostsAndName)

	actionable := switches_client.Actionable{
		Client: switchesClient,
	}

	mqttClient := mqtt.GetMQTTClient(*hostPtr, *usernamePtr, *passwordPtr)
	err := mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	actionRouter := mqtt_action_router.New(
		mqttClient,
		time.Millisecond*10,
		true,
	)

	switches, err := switchesClient.GetSwitches()
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range switches {
		err = actionRouter.AddAction(
			fmt.Sprintf("home/inside/switches/globe/%v/state/set", s.Name),
			switches_client.Arguments{Name: s.Name},
			actionable.On,
			actionable.Off,
			mqtt_action_router.Off,
			fmt.Sprintf("home/inside/switches/globe/%v/state/get", s.Name),
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

	lastStateByName := make(map[string]switches_client.State, 0)

	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-ticker.C:
			switches, err := switchesClient.GetSwitches()
			if err != nil {
				log.Fatal(err)
			}

			for _, s := range switches {
				_, ok := lastStateByName[s.Name]
				if !ok {
					lastStateByName[s.Name] = s.State
				} else if s.State == lastStateByName[s.Name] {
					continue
				}

				err := mqttClient.Publish(
					fmt.Sprintf("home/inside/switches/globe/%v/state/get", s.Name),
					mqtt.ExactlyOnce,
					true,
					fmt.Sprintf("%v", s.State),
				)
				if err != nil {
					log.Fatal(err)
				}

				lastStateByName[s.Name] = s.State
			}
		}
	}
}

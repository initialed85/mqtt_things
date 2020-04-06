package main

import (
	"flag"
	"fmt"
	"github.com/initialed85/mqtt_things/pkg/mqtt_action_router"
	"github.com/initialed85/mqtt_things/pkg/mqtt_client"
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
	flag.Var(&names, "switchName", "a name for a switch")
	flag.Var(&hosts, "switchHost", "a host for a switch")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if len(names) == 0 {
		log.Fatal("no -switchName flags specified")
	}

	if len(hosts) == 0 {
		log.Fatal("no -switchHost flags specified")
	}

	if len(names) != len(hosts) {
		log.Fatal("unbalanced mixture of -switchName and -switchHost flags")
	}

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	hostByName := make(map[string]string)
	for i, name := range names {
		host := hosts[i]
		hostByName[name] = host
	}

	switchesClient := switches_client.New(hostByName)

	actionable := switches_client.Actionable{
		Client: switchesClient,
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

	ticker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <-ticker.C:
			switches, err := switchesClient.GetSwitches()
			if err != nil {
				log.Fatal(err)
			}

			for _, s := range switches {
				err := mqttClient.Publish(
					fmt.Sprintf("home/inside/switches/globe/%v/state/get", s.Name),
					mqtt_client.ExactlyOnce,
					true,
					fmt.Sprintf("%v", s.State),
				)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

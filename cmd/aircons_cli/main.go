package main

import (
	"flag"
	"fmt"
	"github.com/initialed85/mqtt_things/pkg/aircons_client"
	"github.com/initialed85/mqtt_things/pkg/mqtt_action_router"
	"github.com/initialed85/mqtt_things/pkg/mqtt_client"
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
	names      flagArrayString
	hosts      flagArrayString
	codesNames flagArrayString
)

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")
	flag.Var(&hosts, "airconHost", "a host for an aircon")
	flag.Var(&names, "airconName", "a name for an aircon")
	flag.Var(&codesNames, "airconCodesName", "a codes name for an aircon")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if len(hosts) == 0 {
		log.Fatal("no -airconHost flags specified")
	}

	if len(names) == 0 {
		log.Fatal("no -airconName flags specified")
	}

	if len(codesNames) == 0 {
		log.Fatal("no -airconCodesName flags specified")
	}

	if len(names) != len(hosts) || len(names) != len(codesNames) {
		log.Fatal("unbalanced mixture of -airconName and -airconName and airconCodesName flags")
	}

	hostAndNameAndCodesName := make([]aircons_client.HostAndNameAndCodesName, 0)
	for i, name := range names {
		hostAndNameAndCodesName = append(hostAndNameAndCodesName, aircons_client.HostAndNameAndCodesName{
			Host:      hosts[i],
			Name:      name,
			CodesName: codesNames[i],
		})
	}

	airconsClient, err := aircons_client.New(hostAndNameAndCodesName)
	if err != nil {
		log.Fatal(err)
	}

	actionable := aircons_client.Actionable{
		Client: airconsClient,
	}

	mqttClient := mqtt_client.New(*hostPtr, *usernamePtr, *passwordPtr)
	err = mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	actionRouter := mqtt_action_router.New(
		mqttClient,
		time.Millisecond*10,
		true,
	)

	aircons, err := airconsClient.GetAircons()
	if err != nil {
		log.Fatal(err)
	}

	for _, a := range aircons {
		err = actionRouter.AddAction(
			fmt.Sprintf("home/inside/aircons/%v/state/set", a.Name),
			aircons_client.Arguments{Name: a.Name},
			actionable.On,
			actionable.Off,
			mqtt_action_router.Off,
			fmt.Sprintf("home/inside/aircons/%v/state/get", a.Name),
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

	lastStateByName := make(map[string]aircons_client.State, 0)

	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-ticker.C:
			aircons, err = airconsClient.GetAircons()
			if err != nil {
				log.Fatal(err)
			}

			for _, a := range aircons {
				_, ok := lastStateByName[a.Name]
				if !ok {
					lastStateByName[a.Name] = a.State
				} else if a.State == lastStateByName[a.Name] {
					continue
				}

				err := mqttClient.Publish(
					fmt.Sprintf("home/inside/aircons/%v/state/get", a.Name),
					mqtt_client.ExactlyOnce,
					true,
					fmt.Sprintf("%v", a.State),
				)
				if err != nil {
					log.Fatal(err)
				}

				lastStateByName[a.Name] = a.State
			}
		}
	}
}

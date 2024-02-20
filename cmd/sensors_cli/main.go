package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
	"github.com/initialed85/mqtt_things/pkg/sensors_client"
)

const (
	cyclePeriod      = time.Second * 1
	topicPrefix      = "home/inside/environment"
	topicSuffix      = "get"
	presenceAffix    = "presence"
	lightLevelAffix  = "light_level"
	darkAffix        = "dark"
	daylightAffix    = "daylight"
	temperatureAffix = "temperature"
)

func getIntStringFromBool(someBool bool) string {
	if someBool {
		return "1"
	} else {
		return "0"
	}
}

func getTopicFriendlyName(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, " ", "-"), "'", ""))
}

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

	mqttClient := mqtt.GetMQTTClient(*hostPtr, *usernamePtr, *passwordPtr)
	err := mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 16)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err = mqttClient.Disconnect()
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	sensorsClient := sensors_client.New(*bridgeHost, *apiKeyPtr)

	ticker := time.NewTicker(cyclePeriod)

	for {
		select {
		case <-ticker.C:
			sensors, err := sensorsClient.GetSensors()
			if err != nil {
				log.Print(err)
				break
			}

			log.Printf("sensors = %#+v\n", sensors)

			for _, sensor := range sensors {
				err = mqttClient.Publish(
					fmt.Sprintf(
						"%v/%v/%v/%v",
						topicPrefix,
						getTopicFriendlyName(sensor.Name),
						presenceAffix,
						topicSuffix,
					),
					mqtt.ExactlyOnce,
					false,
					getIntStringFromBool(sensor.Presence),
				)
				if err != nil {
					log.Print(err)
					break
				}

				err = mqttClient.Publish(
					fmt.Sprintf(
						"%v/%v/%v/%v",
						topicPrefix,
						getTopicFriendlyName(sensor.Name),
						lightLevelAffix,
						topicSuffix,
					),
					mqtt.ExactlyOnce,
					false,
					fmt.Sprintf("%v", sensor.LightLevel),
				)
				if err != nil {
					log.Print(err)
					break
				}

				err = mqttClient.Publish(
					fmt.Sprintf(
						"%v/%v/%v/%v",
						topicPrefix,
						getTopicFriendlyName(sensor.Name),
						darkAffix,
						topicSuffix,
					),
					mqtt.ExactlyOnce,
					false,
					getIntStringFromBool(sensor.Dark),
				)
				if err != nil {
					log.Print(err)
					break
				}

				err = mqttClient.Publish(
					fmt.Sprintf(
						"%v/%v/%v/%v",
						topicPrefix,
						getTopicFriendlyName(sensor.Name),
						daylightAffix,
						topicSuffix,
					),
					mqtt.ExactlyOnce,
					false,
					getIntStringFromBool(sensor.Daylight),
				)
				if err != nil {
					log.Print(err)
					break
				}

				err = mqttClient.Publish(
					fmt.Sprintf(
						"%v/%v/%v/%v",
						topicPrefix,
						getTopicFriendlyName(sensor.Name),
						temperatureAffix,
						topicSuffix,
					),
					mqtt.ExactlyOnce,
					false,
					fmt.Sprintf("%v", sensor.Temperature),
				)
				if err != nil {
					log.Print(err)
					break
				}
			}
		}
	}
}

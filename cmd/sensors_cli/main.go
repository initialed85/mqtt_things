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

	"github.com/initialed85/mqtt_things/pkg/broadlink_client"
	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
	"github.com/initialed85/mqtt_things/pkg/sensors_client"
)

const (
	cyclePeriod      = time.Millisecond * 1000
	topicPrefix      = "home/inside/environment"
	topicSuffix      = "get"
	presenceAffix    = "presence"
	lightLevelAffix  = "light_level"
	darkAffix        = "dark"
	daylightAffix    = "daylight"
	temperatureAffix = "temperature"
	humidityAffix    = "humidity"
)

func getIntStringFromBool(someBool bool) string {
	if someBool {
		return "1"
	} else {
		return "0"
	}
}

func getTopicFriendlyName(name string) string {
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "’", "") // what is this unicode bs
	name = strings.ToLower(name)

	return name
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

	broadlinkClient, err := broadlink_client.NewPersistentClient()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 1)

	mqttClient := mqtt.GetMQTTClient(*hostPtr, *usernamePtr, *passwordPtr)
	err = mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 1)

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
	time.Sleep(time.Second * 1)

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

			topicFriendlyNames := make(map[string]bool)

			for _, sensor := range sensors {
				topicFriendlyNames[getTopicFriendlyName(sensor.Name)] = true

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
				}
			}

			devices := broadlinkClient.GetDevices()

			for _, device := range devices {
				sensorData, err := device.GetSensorData(time.Second * 5)
				if err != nil {
					log.Printf("warning: failed to get sensor data for %v @ %v (%#+v): %v",
						device.MAC.String(), device.Addr.IP.String(), device.Name,
						err,
					)
					continue
				}

				topicFriendlyName := ""

				for i := 0; i < 64; i++ {
					possibleTopicFriendlyName := getTopicFriendlyName(device.Name)
					if i != 0 {
						possibleTopicFriendlyName = fmt.Sprintf("%v-%v", getTopicFriendlyName(device.Name), i)
					}

					if topicFriendlyNames[possibleTopicFriendlyName] {
						continue
					}

					topicFriendlyName = possibleTopicFriendlyName
					break
				}

				if topicFriendlyName == "" {
					log.Printf("warning: failed to find unconflicting topic friendly name for %#+v", device)
					continue
				}

				topicFriendlyNames[topicFriendlyName] = true

				err = mqttClient.Publish(
					fmt.Sprintf(
						"%v/%v/%v/%v",
						topicPrefix,
						topicFriendlyName,
						temperatureAffix,
						topicSuffix,
					),
					mqtt.ExactlyOnce,
					false,
					fmt.Sprintf("%v", sensorData.Temperature),
				)
				if err != nil {
					log.Print(err)
				}

				err = mqttClient.Publish(
					fmt.Sprintf(
						"%v/%v/%v/%v",
						topicPrefix,
						topicFriendlyName,
						humidityAffix,
						topicSuffix,
					),
					mqtt.ExactlyOnce,
					false,
					fmt.Sprintf("%v", sensorData.Humidity),
				)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}
}

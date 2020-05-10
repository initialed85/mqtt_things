package main

import (
	"flag"
	"fmt"
	"github.com/initialed85/mqtt_things/pkg/mqtt_client_provider"
	"github.com/initialed85/mqtt_things/pkg/mqtt_common"
	"github.com/initialed85/mqtt_things/pkg/weather_client"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	cyclePeriod        = time.Second * 1
	callsPerMinute     = 1
	permissibleAge     = 120 // seconds
	temperatureTopic   = "home/outside/weather/temperature/get"
	pressureTopic      = "home/outside/weather/pressure/get"
	humidityTopic      = "home/outside/weather/humidity/get"
	windSpeedTopic     = "home/outside/weather/wind-speed/get"
	windDirectionTopic = "home/outside/weather/wind-direction/get"
	sunriseTopic       = "home/outside/weather/sunrise/get"
	sunsetTopic        = "home/outside/weather/sunset/get"
)

var (
	allTopics = []string{
		temperatureTopic,
		pressureTopic,
		humidityTopic,
		windSpeedTopic,
		windDirectionTopic,
		sunriseTopic,
		sunsetTopic,
	}
)

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")

	latitudePtr := flag.Float64("latitude", -1337, "weather_client latitude")
	longitudePtr := flag.Float64("longitude", -1337, "weather_client longitude")
	appIDPtr := flag.String("appId", "", "app ID")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if *latitudePtr == -1337 {
		log.Fatal("latitude flag empty")
	}

	if *longitudePtr == -1337 {
		log.Fatal("longitude flag empty")
	}

	if *appIDPtr == "" {
		log.Fatal("appId flag empty")
	}

	mqttClient := mqtt_client_provider.GetMQTTClient(*hostPtr, *usernamePtr, *passwordPtr)
	err := mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		for _, topic := range allTopics {
			err = mqttClient.Publish(topic, mqtt_common.ExactlyOnce, false, "")
			if err != nil {
				log.Print(err)
			}
		}
		err = mqttClient.Disconnect()
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	weatherClient := weather_client.New(*latitudePtr, *longitudePtr, *appIDPtr, callsPerMinute, permissibleAge)

	ticker := time.NewTicker(cyclePeriod)

	for {
		select {
		case <-ticker.C:
			result, err := weatherClient.GetWeather()
			if err != nil {
				log.Print(err)
				break
			}

			log.Printf("result = %+v\n", result)

			err = mqttClient.Publish(
				temperatureTopic,
				mqtt_common.ExactlyOnce,
				false,
				fmt.Sprintf("%v", result.Temp),
			)
			if err != nil {
				log.Print(err)
				break
			}

			err = mqttClient.Publish(
				pressureTopic,
				mqtt_common.ExactlyOnce,
				false,
				fmt.Sprintf("%v", result.Pressure),
			)
			if err != nil {
				log.Print(err)
				break
			}

			err = mqttClient.Publish(
				humidityTopic,
				mqtt_common.ExactlyOnce,
				false,
				fmt.Sprintf("%v", result.Humidity),
			)
			if err != nil {
				log.Print(err)
				break
			}

			err = mqttClient.Publish(
				windSpeedTopic,
				mqtt_common.ExactlyOnce,
				false,
				fmt.Sprintf("%v", result.WindSpeed),
			)
			if err != nil {
				log.Print(err)
				break
			}

			err = mqttClient.Publish(
				windDirectionTopic,
				mqtt_common.ExactlyOnce,
				false,
				fmt.Sprintf("%v", result.WindDirection),
			)
			if err != nil {
				log.Print(err)
				break
			}

			err = mqttClient.Publish(
				sunriseTopic,
				mqtt_common.ExactlyOnce,
				false,
				fmt.Sprintf("%v", result.Sunrise.Unix()*1000),
			)
			if err != nil {
				log.Print(err)
				break
			}

			err = mqttClient.Publish(
				sunsetTopic,
				mqtt_common.ExactlyOnce,
				false,
				fmt.Sprintf("%v", result.Sunset.Unix()*1000),
			)
			if err != nil {
				log.Print(err)
				break
			}
		}
	}
}

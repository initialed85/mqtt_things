package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
	wunderground_weather_server "github.com/initialed85/mqtt_things/pkg/wunderground_weather_server"
)

const (
	timestampTopic         = "/home/outside/weather/timestamp/get"
	stationIDTopic         = "/home/outside/weather/station-id/get"
	latitudeTopic          = "/home/outside/weather/latitude/get"
	longitudeTopic         = "/home/outside/weather/longitude/get"
	temperatureTopic       = "/home/outside/weather/temperature/get"
	dewPointTopic          = "/home/outside/weather/dew-point/get"
	humidityTopic          = "/home/outside/weather/humidity/get"
	windSpeedTopic         = "/home/outside/weather/wind-speed/get"
	windDirectionTopic     = "/home/outside/weather/wind-direction/get"
	windGustTopic          = "/home/outside/weather/wind-gust/get"
	airPressureTopic       = "/home/outside/weather/air-pressure/get"
	rainLast60MinsTopic    = "/home/outside/weather/rain-last-60-mins/get"
	rainTodayTopic         = "/home/outside/weather/rain-today/get"
	temperatureIndoorTopic = "/home/inside/weather/temperature/get"
	humidityIndoorTopic    = "/home/inside/weather/humidity/get"
)

var (
	allTopics = []string{
		timestampTopic,
		stationIDTopic,
		latitudeTopic,
		longitudeTopic,
		temperatureTopic,
		dewPointTopic,
		humidityTopic,
		windSpeedTopic,
		windDirectionTopic,
		windGustTopic,
		airPressureTopic,
		rainLast60MinsTopic,
		rainTodayTopic,
		temperatureIndoorTopic,
		humidityIndoorTopic,
	}
)

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")
	portPtr := flag.Uint64("port", 0, "port")
	latitudePtr := flag.Float64("latitude", 0.0, "")
	longitudePtr := flag.Float64("longitude", 0.0, "")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if *portPtr == 0 {
		log.Fatal("port flag empty")
	}

	if *latitudePtr == 0.0 {
		log.Fatal("latitude flag empty")
	}

	if *longitudePtr == 0.0 {
		log.Fatal("longitude flag empty")
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
		for _, topic := range allTopics {
			err = mqttClient.Publish(topic, mqtt.ExactlyOnce, false, "")
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = wunderground_weather_server.Run(
		ctx,
		uint16(*portPtr),
		*latitudePtr,
		*longitudePtr,
		func(weather wunderground_weather_server.Weather) {
			err = mqttClient.Publish(
				timestampTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%v", weather.Timestamp.UnixNano()),
			)
			if err != nil {
				log.Printf("failed to publish Weather.Timestamp (%v): %v", weather.Timestamp, err)
			}

			err = mqttClient.Publish(
				stationIDTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%v", weather.StationID),
			)
			if err != nil {
				log.Printf("failed to publish Weather.StationID (%v): %v", weather.StationID, err)
			}

			err = mqttClient.Publish(
				latitudeTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%v", weather.Latitude),
			)
			if err != nil {
				log.Printf("failed to publish Weather.Latitude (%v): %v", weather.Latitude, err)
			}

			err = mqttClient.Publish(
				longitudeTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%v", weather.Longitude),
			)
			if err != nil {
				log.Printf("failed to publish Weather.Longitude (%v): %v", weather.Longitude, err)
			}

			err = mqttClient.Publish(
				temperatureTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%.2f", weather.Temperature),
			)
			if err != nil {
				log.Printf("failed to publish Weather.Temperature (%v): %v", weather.Temperature, err)
			}

			err = mqttClient.Publish(
				dewPointTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%.2f", weather.DewPoint),
			)
			if err != nil {
				log.Printf("failed to publish Weather.DewPoint (%v): %v", weather.DewPoint, err)
			}

			err = mqttClient.Publish(
				humidityTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%.2f", weather.Humidity),
			)
			if err != nil {
				log.Printf("failed to publish Weather.Humidity (%v): %v", weather.Humidity, err)
			}

			err = mqttClient.Publish(
				windSpeedTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%.2f", weather.WindSpeed),
			)
			if err != nil {
				log.Printf("failed to publish Weather.WindSpeed (%v): %v", weather.WindSpeed, err)
			}

			err = mqttClient.Publish(
				windDirectionTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%.2f", weather.WindDirection),
			)
			if err != nil {
				log.Printf("failed to publish Weather.WindDirection (%v): %v", weather.WindDirection, err)
			}

			err = mqttClient.Publish(
				windGustTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%.2f", weather.WindGust),
			)
			if err != nil {
				log.Printf("failed to publish Weather.WindGust (%v): %v", weather.WindGust, err)
			}

			err = mqttClient.Publish(
				airPressureTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%.2f", weather.AirPressure),
			)
			if err != nil {
				log.Printf("failed to publish Weather.AirPressure (%v): %v", weather.AirPressure, err)
			}

			err = mqttClient.Publish(
				rainLast60MinsTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%.2f", weather.RainLast60Mins),
			)
			if err != nil {
				log.Printf("failed to publish Weather.RainLast60Mins (%v): %v", weather.RainLast60Mins, err)
			}

			err = mqttClient.Publish(
				rainTodayTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%.2f", weather.RainToday),
			)
			if err != nil {
				log.Printf("failed to publish Weather.RainToday (%v): %v", weather.RainToday, err)
			}

			err = mqttClient.Publish(
				temperatureIndoorTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%.2f", weather.TemperatureIndoor),
			)
			if err != nil {
				log.Printf("failed to publish Weather.TemperatureIndoor (%v): %v", weather.TemperatureIndoor, err)
			}

			err = mqttClient.Publish(
				humidityIndoorTopic,
				mqtt.ExactlyOnce,
				false,
				fmt.Sprintf("%.2f", weather.HumidityIndoor),
			)
			if err != nil {
				log.Printf("failed to publish Weather.HumidityIndoor (%v): %v", weather.HumidityIndoor, err)
			}
		},
	)

	<-ctx.Done()
}

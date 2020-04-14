package main

import (
	"flag"
	"fmt"
	"github.com/initialed85/mqtt_things/pkg/circumstances_engine"
	"github.com/initialed85/mqtt_things/pkg/mqtt_client"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	cyclePeriod      = time.Second * 1
	temperatureTopic = "home/outside/weather/temperature/get"
	sunriseTopic     = "home/outside/weather/sunrise/get"
	sunsetTopic      = "home/outside/weather/sunset/get"
	prefix           = "home/circumstances"
)

var (
	mu                                    sync.Mutex
	temperature                           float64
	gotSunrise, gotSunset, gotTemperature bool
	sunrise, sunset                       time.Time
)

func parseNowForHoursMinutesSeconds(hoursMinutesSeconds string) (time.Time, error) {
	date := time.Now().Format("2006-01-02")

	return time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%v %v", date, strings.TrimSpace(hoursMinutesSeconds)))
}

func parseTemperature(payload string) (float64, error) {
	return strconv.ParseFloat(payload, 64)
}

func parseTimestamp(payload string) (time.Time, error) {
	unixTimestampNanoseconds, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	timestampUtc := time.Unix(unixTimestampNanoseconds/1000, 0)

	return time.Parse(
		"2006-01-02 15:04:05 MST",
		fmt.Sprintf(
			"%v %v",
			timestampUtc.Format("2006-01-02 15:04:05"),
			time.Now().Format("MST"),
		),
	)
}

func handleMessage(message mqtt_client.Message) {
	mu.Lock()
	defer mu.Unlock()

	log.Printf("callback called with %+v", message)

	if message.Topic == temperatureTopic {
		value, err := parseTemperature(message.Payload)
		if err != nil {
			log.Fatal(err)
		}

		temperature = value
		gotTemperature = true
		log.Printf("set temperature to %v", temperature)
	} else if message.Topic == sunriseTopic || message.Topic == sunsetTopic {
		timestamp, err := parseTimestamp(message.Payload)
		if err != nil {
			log.Fatal(err)
		}

		if message.Topic == sunriseTopic {
			sunrise = timestamp
			gotSunrise = true
			log.Printf("set sunrise to %v", sunrise)
		} else {
			sunset = timestamp
			gotSunset = true
			log.Printf("set sunset to %v", sunset)
		}
	} else {
		log.Fatalf("unsupported topic '%v'", message.Topic)
	}
}

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username (optional)")
	passwordPtr := flag.String("password", "", "mqtt password (optional)")
	bedtimePtr := flag.String("bedtime", "22:00:00", "bedtime HH:MM:SS (optional, default 22:00:00)")
	waketimePtr := flag.String("waketime", "06:00:00", "bedtime HH:MM:SS (optional, default 06:00:00)")
	hotEntryPtr := flag.Float64("hotEntry", 29, "hot entry deg C (optional, default 29)")
	hotExitPtr := flag.Float64("hotExit", 27, "hot exit deg C (optional, default 27)")
	coldEntryPtr := flag.Float64("coldEntry", 12, "cold entry deg C (optional, default 12)")
	coldExitPtr := flag.Float64("coldExit", 14, "cold exit deg C (optional, default 14)")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	_, err := parseNowForHoursMinutesSeconds(*bedtimePtr)
	if err != nil {
		log.Fatalf("failed to parse HH:MM:SS fom '%v'", *bedtimePtr)
	}

	_, err = parseNowForHoursMinutesSeconds(*waketimePtr)
	if err != nil {
		log.Fatalf("failed to parse HH:MM:SS fom '%v'", *waketimePtr)
	}

	mqttClient := mqtt_client.New(*hostPtr, *usernamePtr, *passwordPtr)
	err = mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err = mqttClient.Disconnect()
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	for _, topic := range []string{temperatureTopic, sunriseTopic, sunsetTopic} {
		err := mqttClient.Subscribe(topic, mqtt_client.ExactlyOnce, handleMessage)
		if err != nil {
			log.Fatal(err)
		}
	}

	configs := []struct {
		offset                   time.Duration
		lastCircumstancesByTopic map[string]string
		suffix                   string
		filter                   []string
	}{
		{
			time.Duration(0),
			make(map[string]string, 0),
			"",
			[]string{},
		},
		{
			time.Duration(15) * time.Minute,
			make(map[string]string, 0),
			"15m_early",
			[]string{"sunrise", "sunset", "bedtime", "waketime"},
		}, {
			-time.Duration(15) * time.Minute,
			make(map[string]string, 0),
			"15m_late",
			[]string{"sunrise", "sunset", "bedtime", "waketime"},
		},
	}

	ticker := time.NewTicker(cyclePeriod)
	for {
		select {
		case <-ticker.C:
			if !(gotTemperature && gotSunrise && gotSunset) {
				log.Print("temperature, sunrise or sunset not yet populated- deferring for now")

				continue
			}

			bedtime, err := parseNowForHoursMinutesSeconds(*bedtimePtr)
			if err != nil {
				log.Fatal(err)
			}

			waketime, err := parseNowForHoursMinutesSeconds(*waketimePtr)
			if err != nil {
				log.Fatal(err)
			}

			for _, config := range configs {
				circumstances := circumstances_engine.CalculateCircumstances(
					time.Now(),
					sunrise,
					sunset,
					bedtime,
					waketime,
					temperature,
					*hotEntryPtr,
					*hotExitPtr,
					*coldEntryPtr,
					*coldExitPtr,
					config.offset,
				)

				for _, circumstanceAndTopic := range circumstances_engine.GetTopicsAndCircumstances(circumstances, prefix, config.suffix) {
					_, ok := config.lastCircumstancesByTopic[circumstanceAndTopic.Topic]
					if !ok {
						config.lastCircumstancesByTopic[circumstanceAndTopic.Topic] = circumstanceAndTopic.Circumstance
					} else if circumstanceAndTopic.Circumstance == config.lastCircumstancesByTopic[circumstanceAndTopic.Topic] {
						continue
					}

					err = mqttClient.Publish(
						circumstanceAndTopic.Topic,
						mqtt_client.ExactlyOnce,
						true,
						circumstanceAndTopic.Circumstance,
					)
					if err != nil {
						log.Fatal(err)
					}

					log.Printf("published %v to %v", circumstanceAndTopic.Circumstance, circumstanceAndTopic.Topic)

					config.lastCircumstancesByTopic[circumstanceAndTopic.Topic] = circumstanceAndTopic.Circumstance
				}
			}
		}
	}
}

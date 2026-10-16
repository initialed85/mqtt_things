package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
	"github.com/relvacode/iso8601"
)

const (
	cyclePeriod = time.Millisecond * 1000
	topicPrefix = "home/inside/power"
	topicSuffix = "get"
)

/*
{
    "Time": "2026-10-16T14:21:06",
    "ENERGY": {
        "TotalStartTime": "2026-10-16T10:57:45",
        "Total": 0.518,
        "Yesterday": 0.000,
        "Today": 0.518,
        "Power": 248,
        "ApparentPower": 293,
        "ReactivePower": 157,
        "Factor": 0.85,
        "Voltage": 235,
        "Current": 1.246
    }
}
*/

type Payload struct {
	Time   iso8601.Time `json:"Time"`
	Energy struct {
		TotalStartTime iso8601.Time `json:"TotalStartTime"`
		Total          float64      `json:"Total"`
		Yesterday      float64      `json:"Yesterday"`
		Today          float64      `json:"Today"`
		Power          float64      `json:"Power"`
		ApparentPower  int64        `json:"ApparentPower"`
		ReactivePower  int64        `json:"ReactivePower"`
		Factor         float64      `json:"Factor"`
		Voltage        int64        `json:"Voltage"`
		Current        float64      `json:"Current"`
	} `json:"ENERGY"`
}

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	mqttClient := mqtt.GetMQTTClient(*hostPtr, *usernamePtr, *passwordPtr)
	err := mqttClient.Connect()
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

	mqttClient.Subscribe("tele/+/SENSOR", mqtt.ExactlyOnce, func(message mqtt.Message) {
		log.Printf(">>> %s => %s", message.Topic, message.Payload)

		var payload Payload
		err := json.Unmarshal([]byte(message.Payload), &payload)
		if err != nil {
			log.Printf("warning: failed to unmarshal %#+v because %s", message.Payload, err)
			return
		}

		topicParts := strings.Split(message.Topic, "/")
		if len(topicParts) < 2 {
			log.Printf("warning: topic %#+v not in the right structure", message.Topic)
			return
		}

		area := topicParts[1]

		topicBase := fmt.Sprintf("%s/%s", topicPrefix, area)

		err = mqttClient.Publish(fmt.Sprintf("%s/total/%s", topicBase, topicSuffix), mqtt.ExactlyOnce, false, payload.Energy.Total)
		if err != nil {
			log.Printf("warning: %s", err)
		}

		err = mqttClient.Publish(fmt.Sprintf("%s/yesterday/%s", topicBase, topicSuffix), mqtt.ExactlyOnce, false, payload.Energy.Yesterday)
		if err != nil {
			log.Printf("warning: %s", err)
		}

		err = mqttClient.Publish(fmt.Sprintf("%s/today/%s", topicBase, topicSuffix), mqtt.ExactlyOnce, false, payload.Energy.Today)
		if err != nil {
			log.Printf("warning: %s", err)
		}

		err = mqttClient.Publish(fmt.Sprintf("%s/power/%s", topicBase, topicSuffix), mqtt.ExactlyOnce, false, payload.Energy.Power)
		if err != nil {
			log.Printf("warning: %s", err)
		}

		err = mqttClient.Publish(fmt.Sprintf("%s/apparent-power/%s", topicBase, topicSuffix), mqtt.ExactlyOnce, false, payload.Energy.ApparentPower)
		if err != nil {
			log.Printf("warning: %s", err)
		}

		err = mqttClient.Publish(fmt.Sprintf("%s/reactive-power/%s", topicBase, topicSuffix), mqtt.ExactlyOnce, false, payload.Energy.ReactivePower)
		if err != nil {
			log.Printf("warning: %s", err)
		}

		err = mqttClient.Publish(fmt.Sprintf("%s/factor/%s", topicBase, topicSuffix), mqtt.ExactlyOnce, false, payload.Energy.Factor)
		if err != nil {
			log.Printf("warning: %s", err)
		}

		err = mqttClient.Publish(fmt.Sprintf("%s/voltage/%s", topicBase, topicSuffix), mqtt.ExactlyOnce, false, payload.Energy.Voltage)
		if err != nil {
			log.Printf("warning: %s", err)
		}

		err = mqttClient.Publish(fmt.Sprintf("%s/current/%s", topicBase, topicSuffix), mqtt.ExactlyOnce, false, payload.Energy.Current)
		if err != nil {
			log.Printf("warning: %s", err)
		}

	})

	ticker := time.NewTicker(cyclePeriod)

	for range ticker.C {
		// noop
	}
}

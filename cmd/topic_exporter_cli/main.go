package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
)

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

	mu := sync.Mutex{}
	gaugeByTopic := make(map[string]prometheus.Gauge, 0)

	setGauge := func(topic string, value float64) {
		log.Printf("setting guage for %#+v to %#+v", topic, value)

		mu.Lock()
		gauge, ok := gaugeByTopic[topic]
		if !ok {
			gauge = promauto.NewGauge(prometheus.GaugeOpts{
				Name: "mqtt_topic_float64_value",
				Help: "",
				ConstLabels: map[string]string{
					"topic": topic,
				},
			})

			gaugeByTopic[topic] = gauge
		}
		mu.Unlock()

		gauge.Set(value)

	}

	err = mqttClient.Subscribe(
		"+/+/#",
		mqtt.ExactlyOnce,
		func(message mqtt.Message) {
			topic := message.Topic
			payload := strings.TrimSpace(strings.ToLower(message.Payload))

			if strings.Contains(topic, "home/inside/smart-aircons/") && strings.HasSuffix(topic, "/mode/get") {
				heatTopic := strings.ReplaceAll(topic, "/mode/get", "/heat/get")
				coolTopic := strings.ReplaceAll(topic, "/mode/get", "/cool/get")
				fanTopic := strings.ReplaceAll(topic, "/mode/get", "/fan/get")
				if strings.HasPrefix(payload, "off") {
					setGauge(fanTopic, 0.0)
					setGauge(coolTopic, 0.0)
					setGauge(heatTopic, 0.0)
				} else if strings.HasPrefix(payload, "fan_") {
					setGauge(fanTopic, 1.0)
					setGauge(coolTopic, 0.0)
					setGauge(heatTopic, 0.0)
				} else if strings.HasPrefix(payload, "cool_") {
					setGauge(fanTopic, 0.0)
					setGauge(coolTopic, 1.0)
					setGauge(heatTopic, 0.0)
				} else if strings.HasPrefix(payload, "heat_") {
					setGauge(fanTopic, 0.0)
					setGauge(coolTopic, 0.0)
					setGauge(heatTopic, 1.0)
				}

				return
			}

			var value float64

			switch payload {
			case "yes", "true", "y", "t", "on":
				value = 1.0
			case "no", "false", "n", "f", "off":
				value = 0.0
			default:
				value, err = strconv.ParseFloat(message.Payload, 64)
				if err != nil {
					log.Printf("couldn't parse %#+v from %#+v to a float64", message.Payload, topic)
					return
				}
			}

			setGauge(topic, value)
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/metrics", promhttp.Handler())

	err = http.ListenAndServe(":9137", nil)
	if err != nil {
		log.Fatal(err)
	}
}

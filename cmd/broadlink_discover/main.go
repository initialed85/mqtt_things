package main

import (
	"fmt"
	"log"
	"time"

	"github.com/initialed85/mqtt_things/internal/hack"
	"github.com/initialed85/mqtt_things/pkg/broadlink_client"
)

func main() {
	c, err := broadlink_client.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	t := time.NewTicker(time.Second * 1)

	for range t.C {
		devices, err := c.Discover(time.Second * 5)
		if err != nil {
			log.Printf("warning: %s", err)
			continue
		}

		if len(devices) == 0 {
			continue
		}

		for _, device := range devices {
			fmt.Print("\n")

			log.Printf("%s", hack.UnsafeJSONPrettyFormat(device))

			err = device.Auth(time.Second * 5)
			if err != nil {
				log.Printf("warning: %s", err)
				continue
			}

			sensorData, err := device.GetSensorData(time.Second * 5)
			if err != nil {
				log.Printf("warning: %s", err)
				continue
			}

			log.Printf("%s", hack.UnsafeJSONPrettyFormat(sensorData))
		}
	}
}

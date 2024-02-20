package main

import (
	"fmt"
	"log"
	"time"

	"github.com/initialed85/mqtt_things/pkg/broadlink_client"
)

func main() {
	c, err := broadlink_client.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = c.Close()
	}()

	devices, err := c.Discover(time.Second * 2)
	if err != nil {
		log.Fatal(err)
	}

	for _, device := range devices {
		log.Printf("discovered: %v, %v, %v", device.Name, device.MAC, device.Addr.String())

		if device.MAC.String() != "0b:ae:0c:74:46" {
			continue
		}

		err = device.Auth(time.Second * 2)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("authenticated.")

		sensorData, err := device.GetSensorData(time.Second * 2)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("sensors: %#+v", sensorData)

		log.Printf("alright, start punching buttons...")

		lastCode := make([]byte, 0)
		for i := 18; i <= 30; i++ {
			var code []byte

		retry:
			for {
				code, err = device.Learn(time.Second * 60)
				if err != nil {
					log.Fatal(err)
				}

				if len(lastCode) != 0 && len(code) < len(lastCode) {
					continue
				}

				lastCode = code

				break retry
			}

			fmt.Printf("\"heat_%v\": %#+v,\n", i, code)
		}

		// err = device.SendIR(code, time.Second*2)
		// if err != nil {
		// 	log.Fatal(err)
		// }
	}
}

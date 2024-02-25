package main

import (
	"log"
	"time"

	"github.com/initialed85/mqtt_things/pkg/broadlink_client"
)

func main() {
	devices, err := broadlink_client.Discover(time.Second * 5)
	if err != nil {
		log.Fatal(err)
	}

	for _, device := range devices {
		log.Printf("%v\t%v\t%v", device.MAC.String(), device.Addr.IP.String(), device.Name)
	}
}

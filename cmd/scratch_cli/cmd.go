package main

import (
	"log"
	"time"

	"github.com/initialed85/mqtt_things/pkg/broadlink_client"
)

func main() {
	c, err := broadlink_client.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	ds, err := c.Discover(time.Second * 5)
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range ds {
		log.Printf("%v - %v - %v", d.Name, d.MAC.String(), d.Addr.IP)
	}
}

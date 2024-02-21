package main

import (
	"log"
	"time"

	"github.com/initialed85/mqtt_things/pkg/broadlink_client"
)

func main() {
	client, err := broadlink_client.NewPersistentClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	for {
		time.Sleep(time.Second * 1)
	}
}

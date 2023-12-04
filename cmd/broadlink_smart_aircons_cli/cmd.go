package main

import (
	"github.com/initialed85/mqtt_things/pkg/broadlink_smart_aircons_client"
	"log"
)

func main() {
	discoveryResponses, err := broadlink_smart_aircons_client.Discover()
	if err != nil {
		log.Fatal(err)
	}

	for _, discoveryResponse := range discoveryResponses {
		log.Printf("discoveryResponse=%#+v", discoveryResponse)

		authorizationResponse, err := broadlink_smart_aircons_client.Authorize(
			discoveryResponse.Type,
			discoveryResponse.IP,
			discoveryResponse.Port,
		)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("authorizationResponse=%#+v", authorizationResponse)
	}
}

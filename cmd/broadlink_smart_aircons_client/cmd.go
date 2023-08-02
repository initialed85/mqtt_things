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

	log.Printf("discoveryResponses=%#+v", discoveryResponses)

	authorizationResponse, err := broadlink_smart_aircons_client.Authorize(
		discoveryResponses[0].Type,
		discoveryResponses[0].IP,
		discoveryResponses[0].Port,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("authorizationResponse=%#+v", authorizationResponse)
}

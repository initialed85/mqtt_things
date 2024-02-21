package smart_aircons_client

import (
	"fmt"
	"log"
	"strings"
	"time"

	broadlink_client "github.com/initialed85/mqtt_things/pkg/broadlink_client"
)

var (
	client *broadlink_client.PersistentClient
)

func init() {
	var err error

	client, err = broadlink_client.NewPersistentClient()
	if err != nil {
		log.Fatal(err)
	}
}

func BroadlinkSendIR(hostOrMac string, rawCode any) error {
	hostOrMac = strings.ToLower(strings.TrimSpace(hostOrMac))

	log.Printf("sending %#+v to broadlink %v", rawCode, hostOrMac)

	code, ok := rawCode.([]byte)
	if !ok {
		return fmt.Errorf(
			"expected %#+v to be of type []byte; IR code fed to wrong function?",
			rawCode,
		)
	}

	device, err := client.GetDeviceForIP(hostOrMac)
	if err != nil {
		device, err = client.GetDeviceForMac(hostOrMac)
		if err != nil {
			return fmt.Errorf("failed to find device for %#+v: %v", hostOrMac, err)
		}
	}

	err = device.SendIR(code, time.Second*5)
	if err != nil {
		return fmt.Errorf("failed to to send IR to %#+v: %v", hostOrMac, err)
	}

	return nil
}

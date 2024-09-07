package smart_aircons_client

import (
	"fmt"
	"log"
	"strings"
	"time"

	broadlink_client "github.com/initialed85/mqtt_things/pkg/broadlink_client"
)

func BroadlinkSendIR(host string, rawCode []byte) error {
	host = strings.ToLower(strings.TrimSpace(host))

	code := rawCode

	log.Printf("sending %#+v to broadlink %v", code, host)

	device, err := broadlink_client.FromHost(host)
	if err != nil {
		return err
	}

	err = device.SendIR(code, time.Second*5)
	if err != nil {
		return fmt.Errorf("failed to to send IR to %#+v: %v", host, err)
	}

	return nil
}

func BroadlinkLearnIR(host string) ([]byte, error) {
	host = strings.ToLower(strings.TrimSpace(host))

	device, err := broadlink_client.FromHost(host)
	if err != nil {
		return nil, err
	}

	code, err := device.Learn(time.Second * 5)
	if err != nil {
		return nil, fmt.Errorf("failed to to learn IR from %#+v: %v", host, err)
	}

	return code, nil
}

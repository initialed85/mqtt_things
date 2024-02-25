package smart_aircons_client

import (
	"fmt"
	"log"
	"strings"
	"time"

	broadlink_client "github.com/initialed85/mqtt_things/pkg/broadlink_client"
)

func BroadlinkSendIR(host string, rawCode any) error {
	host = strings.ToLower(strings.TrimSpace(host))

	log.Printf("sending %#+v to broadlink %v", rawCode, host)

	code, ok := rawCode.([]byte)
	if !ok {
		return fmt.Errorf(
			"expected %#+v to be of type []byte; IR code fed to wrong function?",
			rawCode,
		)
	}

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

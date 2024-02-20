package smart_aircons_client

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	broadlink_client "github.com/initialed85/mqtt_things/pkg/broadlink_client"
)

var (
	client    *broadlink_client.Client
	startupMu = new(sync.Mutex)
	mu        = new(sync.Mutex)
	devices   []*broadlink_client.Device
)

func init() {
	var err error

	client, err = broadlink_client.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	startupMu.Lock()
	go func() {
		time.Sleep(time.Second * 11)
		startupMu.Unlock()
	}()

	go func() {
		for {
			possibleDevices, err := client.Discover(time.Second * 10)
			if err != nil {
				log.Printf("warning: %v", err)
			}

			if len(possibleDevices) > 0 {
				mu.Lock()
				macs := make([]string, 0)
				for _, device := range devices {
					macs = append(macs, device.MAC.String())
				}
				slices.Sort(macs)
				mu.Unlock()

				possibleMACs := make([]string, 0)
				for _, possibleDevice := range possibleDevices {
					possibleMACs = append(possibleMACs, possibleDevice.MAC.String())
				}
				slices.Sort(possibleMACs)

				if !slices.Equal[[]string](macs, possibleMACs) {
					for _, possibleDevice := range devices {
						log.Printf("discovered %#+v", possibleDevice)

						err = possibleDevice.Auth(time.Second * 1)
						if err != nil {
							log.Printf("warning: %v", err)
							continue
						}

						log.Printf("authenticated %#+v", possibleDevice)
					}

					mu.Lock()
					devices = possibleDevices
					mu.Unlock()
				}

				devices = possibleDevices
			}
		}
	}()
}

func BroadlinkSendIR(hostOrMac string, rawCode any) error {
	hostOrMac = strings.ToLower(strings.TrimSpace(hostOrMac))

	startupMu.Lock()
	defer startupMu.Unlock()

	log.Printf("sending %#+v to broadlink %v", rawCode, hostOrMac)

	code, ok := rawCode.([]byte)
	if !ok {
		return fmt.Errorf(
			"expected %#+v to be of type []byte; IR code fed to wrong function?",
			rawCode,
		)
	}

	mu.Lock()
	possibleDevices := devices
	mu.Unlock()

	var device *broadlink_client.Device

	summaries := make([]string, 0)

	for _, possibleDevice := range possibleDevices {
		summaries = append(summaries, fmt.Sprintf("%v (%v)", possibleDevice.MAC.String(), possibleDevice.Addr.IP.String()))

		if strings.ToLower(strings.TrimSpace(possibleDevice.MAC.String())) == hostOrMac || strings.ToLower(strings.TrimSpace(possibleDevice.Addr.IP.String())) == hostOrMac {
			device = possibleDevice
			break
		}
	}

	if device == nil {
		return fmt.Errorf("failed to find device for %#+v in %v", hostOrMac, strings.Join(summaries, ", "))
	}

	err := device.SendIR(code, time.Second*1)
	if err != nil {
		authErr := device.Auth(time.Second * 1)
		if authErr != nil {
			return fmt.Errorf("got %v trying to auth after getting %v", authErr, err)
		}

		err = device.SendIR(code, time.Second*1)
		if err != nil {
			return err
		}
	}

	return nil
}

package broadlink_client

import (
	"time"
)

func Discover(timeout time.Duration) ([]*StatelessDevice, error) {
	c, err := NewClient()
	if err != nil {
		return nil, err
	}

	devices, err := c.Discover(timeout)
	if err != nil {
		return nil, err
	}

	statelessDevices := make([]*StatelessDevice, 0)

	var lastErr error

	for _, device := range devices {
		statelessDevice, err := FromDevice(device)
		if err != nil {
			lastErr = err
			continue
		}

		statelessDevices = append(statelessDevices, statelessDevice)
	}

	if len(devices) > 0 && len(statelessDevices) == 0 {
		return nil, lastErr
	}

	return statelessDevices, nil
}

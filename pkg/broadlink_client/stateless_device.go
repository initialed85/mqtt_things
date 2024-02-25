package broadlink_client

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type StatelessDevice struct {
	device   *Device
	Name     string
	Type     uint16
	MAC      net.HardwareAddr
	Addr     net.UDPAddr
	LastSeen time.Time
	ID       int32
	Key      []byte
}

func FromDevice(device *Device) (*StatelessDevice, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	d := &StatelessDevice{
		device: &Device{
			mu:       new(sync.Mutex),
			client:   client,
			Name:     device.Name,
			Type:     device.Type,
			MAC:      device.MAC,
			Addr:     device.Addr,
			LastSeen: device.LastSeen,
			ID:       device.ID,
			Key:      device.Key,
		},
		Name:     device.Name,
		Type:     device.Type,
		MAC:      device.MAC,
		Addr:     device.Addr,
		LastSeen: device.LastSeen,
		ID:       device.ID,
		Key:      device.Key,
	}

	err = d.Auth(time.Second * 5)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func FromHost(host string) (*StatelessDevice, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%v:80", host))
	if err != nil {
		return nil, err
	}

	d := &StatelessDevice{
		device: &Device{
			mu:       new(sync.Mutex),
			client:   client,
			Name:     "",
			Type:     0x00,
			MAC:      nil,
			Addr:     *addr,
			LastSeen: time.Now(),
			ID:       0x00,
			Key:      make([]byte, 0),
		},
		Name:     "",
		Type:     0x00,
		MAC:      nil,
		Addr:     *addr,
		LastSeen: time.Now(),
		ID:       0x00,
		Key:      make([]byte, 0),
	}

	err = d.Discover(time.Second * 5)
	if err != nil {
		return nil, err
	}

	err = d.Auth(time.Second * 5)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *StatelessDevice) Discover(timeout time.Duration) error {
	err := d.device.Discover(timeout)
	if err != nil {

		return err
	}

	d.Name = d.device.Name
	d.Type = d.device.Type
	d.MAC = d.device.MAC
	d.LastSeen = d.device.LastSeen

	return nil
}

func (d *StatelessDevice) Auth(timeout time.Duration) error {
	err := d.device.Auth(timeout)
	if err != nil {

		return err
	}

	d.Key = d.device.Key

	return nil
}

func (d *StatelessDevice) GetSensorData(timeout time.Duration) (*SensorData, error) {
	sensorData, err := d.device.GetSensorData(timeout)
	if err != nil {

		return nil, err
	}

	return sensorData, nil
}

func (d *StatelessDevice) Learn(timeout time.Duration) ([]byte, error) {
	data, err := d.device.Learn(timeout)
	if err != nil {

		return nil, err
	}

	return data, nil
}

func (d *StatelessDevice) SendIR(data []byte, timeout time.Duration) error {
	err := d.device.SendIR(data, timeout)
	if err != nil {

		return err
	}

	return nil
}

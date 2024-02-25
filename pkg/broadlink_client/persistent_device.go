package broadlink_client

import (
	"net"
	"time"
)

type PersistentDevice struct {
	device           *Device
	persistentClient *PersistentClient
	Name             string
	Type             uint16
	MAC              net.HardwareAddr
	Addr             net.UDPAddr
	LastSeen         time.Time
	ID               int32
	Key              []byte
}

func FromDeviceAndPersistentClient(device *Device, persistentClient *PersistentClient) *PersistentDevice {
	return &PersistentDevice{
		device:           device,
		persistentClient: persistentClient,
		Name:             device.Name,
		Type:             device.Type,
		MAC:              device.MAC,
		Addr:             device.Addr,
		LastSeen:         device.LastSeen,
		ID:               device.ID,
		Key:              device.Key,
	}
}

func (d *PersistentDevice) Discover(timeout time.Duration) error {
	d.persistentClient.waitForAnyRestart()

	err := d.device.Discover(timeout)
	if err != nil {
		d.persistentClient.requestRestart()
		return err
	}

	d.Key = d.device.Key

	return nil
}

func (d *PersistentDevice) Auth(timeout time.Duration) error {
	d.persistentClient.waitForAnyRestart()

	err := d.device.Auth(timeout)
	if err != nil {
		d.persistentClient.requestRestart()
		return err
	}

	d.Key = d.device.Key

	return nil
}

func (d *PersistentDevice) GetSensorData(timeout time.Duration) (*SensorData, error) {
	d.persistentClient.waitForAnyRestart()

	err := d.device.Auth(timeout)
	if err != nil {
		d.persistentClient.requestRestart()
		return nil, err
	}

	sensorData, err := d.device.GetSensorData(timeout)
	if err != nil {
		d.persistentClient.requestRestart()
		return nil, err
	}

	return sensorData, nil
}

func (d *PersistentDevice) Learn(timeout time.Duration) ([]byte, error) {
	d.persistentClient.waitForAnyRestart()

	err := d.device.Auth(timeout)
	if err != nil {
		d.persistentClient.requestRestart()
		return nil, err
	}

	data, err := d.device.Learn(timeout)
	if err != nil {
		d.persistentClient.requestRestart()
		return nil, err
	}

	return data, nil
}

func (d *PersistentDevice) SendIR(data []byte, timeout time.Duration) error {
	d.persistentClient.waitForAnyRestart()

	err := d.device.Auth(timeout)
	if err != nil {
		d.persistentClient.requestRestart()
		return err
	}

	err = d.device.SendIR(data, timeout)
	if err != nil {
		d.persistentClient.requestRestart()
		return err
	}

	return nil
}

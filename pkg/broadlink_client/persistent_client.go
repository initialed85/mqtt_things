package broadlink_client

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/maps"
)

type PersistentClient struct {
	ctx         context.Context
	cancel      context.CancelFunc
	mu          *sync.Mutex
	client      *Client
	restart     chan bool
	deviceByMAC map[string]*Device
}

func NewPersistentClient() (*PersistentClient, error) {
	c := PersistentClient{
		mu:          new(sync.Mutex),
		restart:     make(chan bool),
		deviceByMAC: make(map[string]*Device),
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())

	err := c.init()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-c.restart:
			}

			for i := 0; i < 10; i++ {
				err = c.init()
				if err != nil {
					if i < 9 {
						log.Printf(
							"warning: attempt %v/%v to init failed: %v; trying again...",
							i+1, 10, err,
						)
						time.Sleep(time.Second * 1)
						continue
					}

					log.Panicf(
						"error: attempt %v/%v to init failed: %v; crashing out...",
						i+1, 10, err,
					)
				}
			}
		}
	}()

	go func() {
		timeout := time.Millisecond * 100

		for {
			select {
			case <-c.ctx.Done():
				return
			default:
			}

			c.mu.Lock()
			client := c.client
			c.mu.Unlock()

			if client == nil {
				time.Sleep(timeout)
				continue
			}

			hadAnySuccess := false

			devices, err := c.client.Discover(timeout)
			if err == nil {
				for _, device := range devices {
					c.mu.Lock()
					existingDevice, ok := c.deviceByMAC[device.MAC.String()]
					if !ok {
						log.Printf("discovered %v @ %v (%#+v)",
							device.MAC.String(), device.Addr.IP.String(), device.Name,
						)
					}
					c.mu.Unlock()

					err = device.Auth(time.Second * 5)
					if err != nil {
						log.Printf(
							"warning: failed to auth %v @ %v (%#+v): %v; ignoring...",
							device.MAC.String(), device.Addr.IP.String(), device.Name,
							err,
						)
						continue
					} else {
						if existingDevice == nil || !slices.Equal(existingDevice.Key, device.Key) {
							log.Printf("authenticated %v @ %v (%#+v) - key is %#+v",
								device.MAC.String(), device.Addr.IP.String(), device.Name, device.Key,
							)
						}
					}

					device.requestRestart = c.RequestRestart

					c.mu.Lock()
					c.deviceByMAC[device.MAC.String()] = device
					c.mu.Unlock()

					hadAnySuccess = true
				}
			}

			if len(devices) > 0 && !hadAnySuccess {
				log.Printf("warning: found %v devices but auth'd none of them; restarting...", len(devices))
				select {
				case c.restart <- true:
					time.Sleep(time.Second * 1)
				default:
					time.Sleep(time.Second * 5)
				}
			}

			c.mu.Lock()
			existingDevices := maps.Values(c.deviceByMAC)
			for _, device := range existingDevices {
				if time.Since(device.LastSeen) > time.Second*30 {
					delete(c.deviceByMAC, device.MAC.String())
					log.Printf("expired %v @ %v (%#+v)",
						device.MAC.String(), device.Addr.IP.String(), device.Name,
					)
				}
			}
			c.mu.Unlock()

			if timeout < time.Second*10 {
				timeout += time.Millisecond * 100
			}
		}
	}()

	return &c, nil
}

func (c *PersistentClient) init() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error

	if c.client != nil {
		log.Printf("closing existing client...")
		_ = c.client.Close()
		c.client = nil
	}

	log.Printf("opening new client...")
	c.client, err = NewClient()
	if err != nil {
		return err
	}

	log.Printf("init done.")

	return nil
}

func (c *PersistentClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cancel()

	if c.client != nil {
		_ = c.client.Close()
		c.client = nil
	}
}

func (c *PersistentClient) RequestRestart() {
	select {
	case c.restart <- true:
	default:
	}
}

func (c *PersistentClient) GetDevices() []*Device {
	c.mu.Lock()
	defer c.mu.Unlock()

	return maps.Values(c.deviceByMAC)
}

func (c *PersistentClient) GetDeviceForMac(mac string) (*Device, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	device, ok := c.deviceByMAC[mac]
	if !ok {
		return nil, fmt.Errorf("failed to find device for MAC %v; know about %v", mac, strings.Join(maps.Keys(c.deviceByMAC), ", "))
	}

	return device, nil
}

func (c *PersistentClient) GetDeviceForIP(ip string) (*Device, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ips := make([]string, 0)
	for _, device := range c.deviceByMAC {
		if device.Addr.IP.String() == ip {
			return device, nil
		}

		ips = append(ips, device.Addr.IP.String())
	}

	return nil, fmt.Errorf("failed to find device for IP %v; know about %v", ip, strings.Join(ips, ", "))
}

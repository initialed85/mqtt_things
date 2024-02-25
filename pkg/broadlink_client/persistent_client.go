package broadlink_client

import (
	"context"
	"fmt"
	"log"
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
	restartMu   *sync.Mutex
	deviceByMAC map[string]*PersistentDevice
}

func NewPersistentClient() (*PersistentClient, error) {
	c := PersistentClient{
		mu:          new(sync.Mutex),
		restart:     make(chan bool),
		deviceByMAC: make(map[string]*PersistentDevice),
		restartMu:   new(sync.Mutex),
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

			log.Printf("handling requested restart...")

			c.restartMu.Lock()

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

				log.Printf(
					"attempt %v/%v to init succeeded; sleeping a little...",
					i+1, 10,
				)

				break
			}

			time.Sleep(time.Second * 1)

			c.restartMu.Unlock()

			log.Printf("requested restart done")
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

			devices, err := c.client.Discover(timeout)
			if err == nil {
				for _, device := range devices {
					c.mu.Lock()
					persistentDevice, ok := c.deviceByMAC[device.MAC.String()]
					c.mu.Unlock()

					if !ok {
						persistentDevice = FromDeviceAndPersistentClient(device, &c)

						log.Printf("discovered %v @ %v (%#+v)",
							persistentDevice.MAC.String(), persistentDevice.Addr.IP.String(), persistentDevice.Name,
						)

						err = persistentDevice.Auth(time.Second * 5)
						if err != nil {
							log.Printf(
								"warning: failed to auth %v @ %v (%#+v): %v; ignoring...",
								persistentDevice.MAC.String(), persistentDevice.Addr.IP.String(), persistentDevice.Name,
								err,
							)
							continue
						}

						log.Printf("authenticated %v @ %v (%#+v) - key is %#+v",
							persistentDevice.MAC.String(), persistentDevice.Addr.IP.String(), persistentDevice.Name, persistentDevice.Key,
						)

						c.mu.Lock()
						c.deviceByMAC[device.MAC.String()] = persistentDevice
						c.mu.Unlock()
					}

					persistentDevice.device.Addr = device.Addr
					persistentDevice.device.Name = device.Name
					persistentDevice.device.LastSeen = device.LastSeen

					persistentDevice.Addr = device.Addr
					persistentDevice.Name = device.Name
					persistentDevice.LastSeen = device.LastSeen
				}
			} else {
				log.Printf("warning: failed to discover devices: %v", err)
				c.requestRestart()
			}

			c.mu.Lock()
			existingDevices := maps.Values(c.deviceByMAC)
			for _, persistentDevice := range existingDevices {
				if time.Since(persistentDevice.LastSeen) > time.Second*30 {
					delete(c.deviceByMAC, persistentDevice.MAC.String())
					log.Printf("expired %v @ %v (%#+v)",
						persistentDevice.MAC.String(), persistentDevice.Addr.IP.String(), persistentDevice.Name,
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

	for _, persistentDevice := range c.deviceByMAC {
		persistentDevice.device.client = c.client
	}

	log.Printf("init done.")

	return nil
}

func (c *PersistentClient) requestRestart() {
	select {
	case c.restart <- true:
		log.Printf("accepted restart request")
	default:
		log.Printf("rejected restart request (must be one in-progress)")
	}
}

func (c *PersistentClient) waitForAnyRestart() {
	c.mu.Lock()
	func() {}() // noop
	c.mu.Unlock()
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

func (c *PersistentClient) GetDevices() []*PersistentDevice {
	c.mu.Lock()
	defer c.mu.Unlock()

	return maps.Values(c.deviceByMAC)
}

func (c *PersistentClient) GetDeviceForMac(mac string) (*PersistentDevice, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	device, ok := c.deviceByMAC[mac]
	if !ok {
		return nil, fmt.Errorf("failed to find device for MAC %v; know about %v", mac, strings.Join(maps.Keys(c.deviceByMAC), ", "))
	}

	return device, nil
}

func (c *PersistentClient) GetDeviceForIP(ip string) (*PersistentDevice, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ips := make([]string, 0)
	for _, persistentDevice := range c.deviceByMAC {
		if persistentDevice.device.Addr.IP.String() == ip {
			return persistentDevice, nil
		}

		ips = append(ips, persistentDevice.device.Addr.IP.String())
	}

	return nil, fmt.Errorf("failed to find device for IP %v; know about %v", ip, strings.Join(ips, ", "))
}

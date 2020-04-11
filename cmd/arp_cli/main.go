package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/initialed85/mqtt_things/pkg/mqtt_client"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// copied heavily from https://raw.githubusercontent.com/google/gopacket/master/examples/arpscan/arpscan.go
//
// therefore subject to the following license:
//
// ----
// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.
// ----
//
// for clarity this refers to https://github.com/google/gopacket/blob/master/LICENSE

// unrelated warning: here be dragons

type flagArrayString []string

func (f *flagArrayString) String() string {
	return fmt.Sprintf("%d", f)
}

func (f *flagArrayString) Set(value string) error {
	*f = append(*f, value)

	return nil
}

const (
	topicPrefix     = "home/arp"
	timeoutDuration = time.Second * 5
)

var (
	arpIPs        flagArrayString
	mqttClient    mqtt_client.Client
	lastSeenByIP  = make(map[string]time.Time, 0)
	lastStateByIP = make(map[string]string, 0)
)

func readARP(handle *pcap.Handle, iface *net.Interface, stop chan struct{}) {
	src := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	in := src.Packets()
	for {
		var packet gopacket.Packet
		select {
		case <-stop:
			return
		case packet = <-in:
			arpLayer := packet.Layer(layers.LayerTypeARP)
			if arpLayer == nil {
				continue
			}

			arp := arpLayer.(*layers.ARP)
			if arp.Operation != layers.ARPReply || bytes.Equal([]byte(iface.HardwareAddr), arp.SourceHwAddress) {
				continue
			}

			ip := net.IP(arp.SourceProtAddress)
			mac := net.HardwareAddr(arp.SourceHwAddress)

			found := false
			for _, compareIP := range arpIPs {
				if compareIP == ip.String() {
					found = true

					break
				}
			}

			if !found {
				continue
			}

			log.Printf("found %v at %v", ip, mac)

			ipString := ip.String()

			lastSeenByIP[ipString] = time.Now()
			lastStateByIP[ipString] = "1"
		}
	}
}

func writeARP(handle *pcap.Handle, iface *net.Interface, addr *net.IPNet) error {
	eth := layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}

	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(iface.HardwareAddr),
		SourceProtAddress: []byte(addr.IP),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	ips := make([]net.IP, 0)
	for _, ip := range arpIPs {
		ips = append(ips, net.ParseIP(ip))
	}

	log.Printf("arping for %+v", ips)
	for _, ip := range ips {
		arp.DstProtAddress = ip[len(ip)-4:]
		err := gopacket.SerializeLayers(buf, opts, &eth, &arp)
		if err != nil {
			log.Fatalf("failed SerializeLayers with %+v: %v", arp, err)
			return err
		}

		err = handle.WritePacketData(buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func scan(iface *net.Interface) error {
	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}

	var addr *net.IPNet
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				addr = &net.IPNet{
					IP:   ip4,
					Mask: ipnet.Mask[len(ipnet.Mask)-4:],
				}

				break
			}
		}
	}

	if addr == nil {
		return errors.New("no good IP network found")
	} else if addr.IP[0] == 127 {
		return errors.New("skipping localhost")
	}

	handle, err := pcap.OpenLive(iface.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	stop := make(chan struct{})
	go readARP(handle, iface, stop)
	defer close(stop)

	for {
		err := writeARP(handle, iface, addr)
		if err != nil {
			log.Printf("error writing packets on %v: %v", iface.Name, err)

			return err
		}

		time.Sleep(time.Second * 1)
	}

	return nil
}

func expire() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			for ip, lastSeen := range lastSeenByIP {
				timeout := lastSeen.Add(timeoutDuration)
				if timeout.Before(now) {
					lastStateByIP[ip] = "0"
				}
			}
		}
	}
}

func publish() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			for ip, state := range lastStateByIP {
				err := mqttClient.Publish(
					fmt.Sprintf("%v/%v/get", topicPrefix, ip),
					mqtt_client.ExactlyOnce,
					false,
					state,
				)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")
	interfaceName := flag.String("interfaceName", "", "interface to capture on")
	flag.Var(&arpIPs, "arpIP", "a host to ARP")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if *interfaceName == "" {
		log.Fatal("no -interfaceName flags specified")
	}

	if len(arpIPs) == 0 {
		log.Fatal("no -arpIP flags specified")
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, iface := range ifaces {
		if iface.Name != *interfaceName {
			continue
		}

		wg.Add(1)
		go func(iface net.Interface) {
			defer wg.Done()
			if err := scan(&iface); err != nil {
				log.Printf("interface %v: %v", iface.Name, err)
			}
		}(iface)

		break
	}

	mqttClient = mqtt_client.New(*hostPtr, *usernamePtr, *passwordPtr)
	err = mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	for _, ip := range arpIPs {
		lastSeenByIP[ip] = time.Now().Add(-timeoutDuration)
		lastStateByIP[ip] = "0"
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		for _, ip := range arpIPs {
			err := mqttClient.Publish(
				fmt.Sprintf("%v/%v/get", topicPrefix, ip),
				mqtt_client.ExactlyOnce,
				false,
				"0",
			)
			if err != nil {
				log.Fatal(err)
			}
		}

		err = mqttClient.Disconnect()
		if err != nil {
			log.Print(err)
		}

		os.Exit(0)
	}()

	go expire()

	go publish()

	wg.Wait()
}
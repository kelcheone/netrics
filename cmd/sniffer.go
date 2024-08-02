package sniffer

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type DataUsage struct {
	TotalBytes    int64
	IncomingBytes int64
	OutGoingBytes int64
	Timestamp     time.Time
	PIDUsage      map[int]int64
}

type Usage map[string]*DataUsage

type LogUsage func(usage Usage, ip string, bytes int64, incoming bool, timestamp time.Time)

func Sniff(logUsage LogUsage) error {
	iface, err := getDefaultInterface()
	if err != nil {
		return err
	}
	log.Printf("monitoring the network on interface: %s ..............", iface)
	handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	usage := make(Usage)

	localIP, err := getLocalIP()
	if err != nil {
		return err
	}

	localNet := &net.IPNet{
		IP:   localIP,
		Mask: net.CIDRMask(24, 32),
	}

	for packet := range packetSource.Packets() {
		processPacket(packet, usage, logUsage, localNet)
	}
	return nil
}

func getLocalIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP, nil
			}
		}
	}
	return nil, fmt.Errorf("no local IP address found")
}

func processPacket(packet gopacket.Packet, usage Usage, logUsage LogUsage, localNet *net.IPNet) {
	IPLayer := packet.Layer(layers.LayerTypeIPv4)
	if IPLayer == nil {
		return
	}

	ip, ok := IPLayer.(*layers.IPv4)
	if !ok {
		return
	}

	srcIP := ip.SrcIP
	dstIP := ip.DstIP
	packetLength := int64(len(packet.Data()))
	timestamp := time.Now()

	if localNet.Contains(srcIP) {
		updateUsage(usage, dstIP.String(), packetLength, false, timestamp)
		logUsage(usage, dstIP.String(), packetLength, false, timestamp)
	}
	if localNet.Contains(dstIP) {
		updateUsage(usage, srcIP.String(), packetLength, true, timestamp)
		logUsage(usage, srcIP.String(), packetLength, true, timestamp)
	}
}

func updateUsage(usage Usage, ip string, bytes int64, incoming bool, timestamp time.Time) {
	if _, exists := usage[ip]; !exists {
		usage[ip] = &DataUsage{Timestamp: timestamp}
	}
	usage[ip].TotalBytes += bytes
	if incoming {
		usage[ip].IncomingBytes += bytes
	} else {
		usage[ip].OutGoingBytes += bytes
	}
	usage[ip].Timestamp = timestamp
}

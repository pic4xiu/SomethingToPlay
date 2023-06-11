package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type DNSPacket struct {
	ID      uint16 `gorm:"primary_key"`
	Queries string
	Answers string
}

func main() {
	// Open the device for capturing
	handle, err := pcap.OpenLive("en0", 65535, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Set a filter to capture only DNS packets
	filter := "udp and port 53"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}

	// Open a connection to the MySQL database
	db, err := gorm.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/database?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the DNSPacket table if it doesn't exist
	db.AutoMigrate(&DNSPacket{})

	// Start capturing packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// Parse the DNS packet
		dnsLayer := packet.Layer(layers.LayerTypeDNS)
		if dnsLayer != nil {
			dnsPacket := dnsLayer.(*layers.DNS)
			if dnsPacket.QR == false && dnsPacket.OpCode == layers.DNSOpCodeQuery {
				// Extract the queries from the DNS packet
				queries := make([]string, len(dnsPacket.Questions))

				for i, question := range dnsPacket.Questions {
					queries[i] = string(question.Name)
					fmt.Println((question))
				}

				// Check if the DNSPacket already exists in the database
				var count int
				db.Model(&DNSPacket{}).Where("id = ?", dnsPacket.ID).Count(&count)
				if count == 0 {
					// Recursively resolve the DNS queries
					answers := make([]string, 0)
					for _, query := range queries {
						// fmt.Println(query)
						ips, err := resolveDNS(query)
						if err == nil {
							answers = append(answers, ips...)
						}
					}

					// Create a new DNSPacket object and save it to the database
					dnsPacket := DNSPacket{
						ID:      dnsPacket.ID,
						Queries: strings.Join(queries, ","),
						Answers: strings.Join(answers, ","),
					}
					db.Create(&dnsPacket)
				}
			}
		}
	}
}

func resolveDNS(query string) ([]string, error) {
	// Resolve the DNS query using the system resolver
	ips := make([]string, 0)
	addrs, err := net.LookupHost(query)
	if err != nil {
		return ips, err
	}
	for _, addr := range addrs {
		ips = append(ips, addr)
	}
	return ips, nil
}

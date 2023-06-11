package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func main() {
	// 打开数据库连接
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/database")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 创建TCP数据包表
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS tcp_packets (id INT AUTO_INCREMENT PRIMARY KEY, src_ip VARCHAR(15), dst_ip VARCHAR(15), src_port INT, dst_port INT)")
	if err != nil {
		log.Fatal(err)
	}

	// 打开网络接口
	handle, err := pcap.OpenLive("en0", 65535, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// 设置过滤器，只捕获TCP流量
	filter := "tcp"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}

	// 循环读取TCP数据包
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// 解析TCP数据包
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			srcIP := packet.NetworkLayer().NetworkFlow().Src().String()
			dstIP := packet.NetworkLayer().NetworkFlow().Dst().String()
			srcPort := tcp.SrcPort
			dstPort := tcp.DstPort
			// 检查TCP数据包是否包含SYN标志，但不包含ACK标志
			if tcp.SYN && !tcp.ACK {
				// 将五元组信息存储到数据库中
				_, err = db.Exec("INSERT INTO tcp_packets (src_ip, dst_ip, src_port, dst_port) VALUES (?, ?, ?, ?)", srcIP, dstIP, srcPort, dstPort)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

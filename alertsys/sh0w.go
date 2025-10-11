package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/cors"
)

type Packet struct {
	Timestamp       int64  `json:"timestamp"`
	SourceIP        string `json:"source_ip"`
	DestinationIP   string `json:"destination_ip"`
	SourcePort      int    `json:"source_port"`
	DestinationPort int    `json:"destination_port"`
	Payload         []byte `json:"payload"`
}

func main() {
	// 连接MySQL数据库
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/database")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 创建HTTP处理程序
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 查询数据库中的数据
		rows, err := db.Query("SELECT timestamp, source_ip, destination_ip, source_port, destination_port, payload FROM packets")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		// 将数据转换为Packet结构体，并将其编码为JSON格式
		packets := make([]Packet, 0)
		for rows.Next() {
			var packet Packet
			err := rows.Scan(&packet.Timestamp, &packet.SourceIP, &packet.DestinationIP, &packet.SourcePort, &packet.DestinationPort, &packet.Payload)
			if err != nil {
				log.Fatal(err)
			}
			packets = append(packets, packet)
		}
		jsonBytes, err := json.Marshal(packets)
		if err != nil {
			log.Fatal(err)
		}

		// 将JSON响应发送回客户端
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	// 添加CORS标头
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
	})
	handler = c.Handler(handler).(http.HandlerFunc)

	// 启动HTTP服务器
	log.Fatal(http.ListenAndServe(":8080", handler))
}

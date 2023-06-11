package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/go-redis/redis"
	"github.com/streadway/amqp"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// 连接到RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// 创建一个通道
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 声明要消费的队列
	q, err := ch.QueueDeclare(
		"requests", // 队列名称
		false,      // 是否持久化
		false,      // 是否自动删除
		false,      // 是否具有排他性
		false,      // 是否阻塞
		nil,        // 额外参数
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// 消费消息
	// 消费消息
	msgs, err := ch.Consume(
		q.Name, // 队列名称
		"",     // 消费者名称
		true,   // 是否自动应答
		false,  // 是否具有排他性
		false,  // 是否阻塞
		false,  // 额外参数
		nil,    // 额外参数
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// 处理消息
	for msg := range msgs {
		// 解析Session和URL字段
		var Session, URL string
		fields := strings.Split(string(msg.Body), ",")
		for _, field := range fields {
			if strings.HasPrefix(field, "Session:") {
				Session = strings.TrimSpace(strings.TrimPrefix(field, "Session:"))
			}
		}
		re := regexp.MustCompile(`URL:\s*([^,]+)`)
		match := re.FindStringSubmatch(string(msg.Body))
		if len(match) > 1 {
			URL = match[1]
			fmt.Printf("URL: %s\n", URL)
		}
		// 打印结果
		fmt.Printf("%s %s\n", Session, URL)
		// Get value from Redis hash
		key := "cache"
		field := URL
		val, _ := redisClient.HGet(key, field).Result()
		// Check if value exists
		if val == "" {
			// Value does not exist
			fmt.Println("Value does not exist")
			err = ch.Publish(
				"",        // 交换机名称
				"newurls", // 队列名称
				false,     // 是否强制发送到队列
				false,     // 是否等待服务器确认
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(fmt.Sprintf("Session: %s, URL: %s", Session, URL)),
				},
			)
			if err != nil {
				log.Fatalf("Failed to publish a message: %v", err)
			}
		} else if val == "1" {
			// Value exists, add session to BL set
			setKey := "blacklist"
			err := redisClient.SAdd(setKey, Session).Err()
			if err != nil {
				log.Fatal(err)
			}
		}

	}
}

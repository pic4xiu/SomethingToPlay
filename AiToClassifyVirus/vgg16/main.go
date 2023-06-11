package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
)

// 处理查询请求
func getResultHandler(w http.ResponseWriter, r *http.Request) {
	// 检查请求方法是否为GET
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析查询参数
	md5 := r.URL.Query().Get("md5")

	// 连接到MySQL数据库
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/mydatabase")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 查询文件处理结果
	var result int
	err = db.QueryRow("SELECT result FROM files WHERE md5 = ?", md5).Scan(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回查询结果
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	tpl := template.Must(template.ParseFiles("result.html"))
	tpl.Execute(w, map[string]interface{}{
		"MD5":    md5,
		"Result": result,
	})
}

// 处理文件上传请求
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// 检查请求方法是否为POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析表单数据
	err := r.ParseMultipartForm(32 << 20) // 限制上传文件大小为32MB
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 将文件保存到本地
	fileDir := "./uploads"
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		os.Mkdir(fileDir, 0755)
	}
	filePath := filepath.Join(fileDir, handler.Filename)
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// 计算文件的MD5值
	hash := md5.New()
	if _, err := io.Copy(f, io.TeeReader(file, hash)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	md5Hash := fmt.Sprintf("%x", hash.Sum(nil))
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/mydatabase")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 查询文件处理结果
	var result int
	err = db.QueryRow("SELECT result FROM files WHERE md5 = ?", md5Hash).Scan(&result)
	if err == nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		tpl := template.Must(template.ParseFiles("result.html"))
		tpl.Execute(w, map[string]interface{}{
			"MD5":    md5Hash,
			"Result": result,
		})
		return
	}

	// 将文件名和MD5值发送到RabbitMQ队列中
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"file_queue", // 队列名称
		true,         // 是否持久化
		false,        // 是否自动删除
		false,        // 是否具有排他性
		false,        // 是否阻塞
		nil,          // 额外参数
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	body := []byte(fmt.Sprintf("%s|%s", handler.Filename, md5Hash))
	err = ch.Publish(
		"",     // 交换机名称
		q.Name, // 队列名称
		false,  // 是否强制
		false,  // 是否立即
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish a message: %v", err)
	}
	// 获取队列中的消息数量
	queue, err := ch.QueueInspect(q.Name)
	if err != nil {
		log.Fatalf("Failed to inspect queue: %v", err)
	}
	messageCount := queue.Messages

	// 返回上传成功的响应
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	tpl := template.Must(template.ParseFiles("index.html"))
	tpl.Execute(w, map[string]interface{}{
		"Message":      "Upload successful",
		"MessageCount": messageCount,
		"MD5":          md5Hash,
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	tpl := template.Must(template.ParseFiles("index.html"))
	tpl.Execute(w, nil)
}

func main() {
	// 注册文件上传处理函数
	http.HandleFunc("/getresult", getResultHandler)
	http.HandleFunc("/upload", uploadHandler)

	// 注册根路由处理函数
	http.HandleFunc("/", indexHandler)

	// 启动HTTP服务器
	log.Fatal(http.ListenAndServe(":7777", nil))
}

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
)

// User represents a user in the system
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims represents the JWT claims for a user
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func connect() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/mydb")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	db, err := connect()
	if err != nil {

		return
	}
	defer db.Close()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare a queue
	q, err := ch.QueueDeclare(
		"requests", // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Handle login requests
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var user1 User

		// Check the username and password
		_ = db.QueryRow("SELECT username, password FROM users WHERE username = ?", user.Username).Scan(&user1.Username, &user1.Password)

		if user != user1 {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// if (user.Username == "admin" && user.Password == "password") || (user.Username == "test" && user.Password == "test") {
		// 	log.Printf("%s login in ", user.Username)
		// } else {
		// 	http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		// 	return
		// }

		// Create a JWT token
		claims := &Claims{
			Username: user.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: jwt.TimeFunc().Add(time.Hour * 24).Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("secret"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the token in the response header
		w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))

		// Send the response
		w.Write([]byte("Hello World"))
	})

	// http.HandleFunc("/invalidate", func(w http.ResponseWriter, r *http.Request) {
	// 	if r.FormValue("key") != "qweasd" {
	// 		http.Error(w, "Invalid request", http.StatusBadRequest)
	// 		return
	// 	}
	// 	// 从请求头中获取 token 字符串
	// 	authHeader := r.Header.Get("Authorization")
	// 	if authHeader == "" {
	// 		http.Error(w, "Authorization header is missing", http.StatusBadRequest)
	// 		return
	// 	}
	// 	tokenString := authHeader[len("Bearer "):]

	// 	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
	// 		return []byte("secret"), nil
	// 	})
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusUnauthorized)
	// 		return
	// 	}

	// 	// 找username
	// 	claims, ok := token.Claims.(*Claims)
	// 	if !ok {
	// 		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
	// 		return
	// 	}

	// 	redisClient.SAdd("blacklist", claims.Username).Err()

	// })

	// Handle all other requests
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get the JWT token from the request header
		fmt.Println(r.URL)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := authHeader[len("Bearer "):]

		// fmt.Print(you)
		// Parse the JWT token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Get the username from the JWT claims
		claims, ok := token.Claims.(*Claims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		if redisClient.SIsMember("blacklist", claims.Username).Val() {
			http.Error(w, "Sry U R in blacklist", http.StatusUnauthorized)
			return
		}
		// Send the session and URL to RabbitMQ
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(fmt.Sprintf("Session: %s, URL: %s", claims.Username, r.URL)),
			})

		if err != nil {
			log.Printf("Failed to publish message: %v", err)
		}

		// Send the response
		w.Write([]byte("Hello World"))
	})

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", nil))
}

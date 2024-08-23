package main

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

func main() {
	ctx := context.Background()

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

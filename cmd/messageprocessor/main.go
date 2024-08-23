package main

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

const messageExchange = "messages"

type message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

type redisKey struct {
	s string `json:"s"`
	r string `json:"r"`
}

func main() {
	ctx := context.Background()

	// setup redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// setup RabbitMQ connection
	conn, err := amqp.Dial("amqp://user:password@localhost:7001/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		messageExchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare RabbitMQ exchange: %v", err)
	}

	// temporary queue for receiving fanout messages,
	// bind it to the exchange
	q, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare temporary RabbitMQ queue: %v", err)
	}

	err = ch.QueueBind(
		q.Name,
		"",
		messageExchange,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind temporary RabbitMQ queue to message exchange: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)

			var messageBody message

			err := json.Unmarshal(d.Body, &messageBody)
			if err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			// store message in Redis
			key := redisKey{s: messageBody.Sender, r: messageBody.Receiver}
			keyBytes, err := json.Marshal(key)
			if err != nil {
				panic("failed to marshal key - this should never happen")
			}

			err = rdb.LPush(ctx, string(keyBytes), messageBody.Message).Err()
			if err != nil {
				log.Printf("Failed to store message in Redis: %v", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}

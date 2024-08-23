package main

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const messageExchange = "messages"

type messageRequest struct {
	Sender   string `json:"sender" binding:"required"`
	Receiver string `json:"receiver" binding:"required"`
	Message  string `json:"message" binding:"required"`
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

	r := gin.Default()

	r.GET("/message", func(c *gin.Context) {
		var req messageRequest

		err := c.BindJSON(&req)
		if err != nil {
			// 400 Bad Request automatically returned by BindWith
			return
		}

		// marshal again with guaranteed success and only the fields we need
		rabbitmqBody, err := json.Marshal(req)
		if err != nil {
			panic("failed to marshal request - this should never happen")
		}

		// forward message to RabbitMQ
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		err = ch.PublishWithContext(
			ctx,
			messageExchange,
			"",
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        rabbitmqBody,
			})

		c.JSON(http.StatusOK, gin.H{})
	})

	r.Run()
}

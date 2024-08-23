package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)

	r := gin.Default()

	r.GET("/message", func(c *gin.Context) {
		var req messageRequest

		err := c.BindJSON(&req)
		if err != nil {
			// 400 Bad Request automatically returned by BindWith
			return
		}

		// forward message to RabbitMQ

		c.JSON(http.StatusOK, gin.H{})
	})

	r.Run()
}

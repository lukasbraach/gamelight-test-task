package main

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"net/http"

	"github.com/gin-gonic/gin"
)

type messageListRequest struct {
	Sender   string `json:"sender" binding:"required"`
	Receiver string `json:"receiver" binding:"required"`
}

type messageResponse struct {
	Sender   string `json:"sender" binding:"required"`
	Receiver string `json:"receiver" binding:"required"`
	Message  string `json:"message" binding:"required"`
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

	r := gin.Default()

	r.GET("/message/list", func(c *gin.Context) {
		var req messageListRequest
		err := c.BindJSON(&req)
		if err != nil {
			// 400 Bad Request automatically returned by BindWith
			return
		}

		// Create a key for the sender and receiver
		key := redisKey{s: req.Sender, r: req.Receiver}
		keyBytes, err := json.Marshal(key)
		if err != nil {
			panic("failed to marshal key - this should never happen")
		}

		// Get matching messages from Redis
		values, err := rdb.LRange(ctx, string(keyBytes), 0, -1).Result()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Convert messages to response format
		list := make([]messageResponse, len(values))

		for i, v := range values {
			// no time to overengineer stuff.
			// If I had used a Redis stream, things would have been easier...
			list[len(values)-i-1] = messageResponse{
				Sender:   req.Sender,
				Receiver: req.Receiver,
				Message:  v,
			}

		}

		c.JSON(http.StatusOK, list)
	})

	r.Run()
}

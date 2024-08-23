package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type messageListRequest struct {
	Sender   string `json:"sender" binding:"required"`
	Receiver string `json:"receiver" binding:"required"`
}

func main() {
	r := gin.Default()

	r.GET("/message/list", func(c *gin.Context) {
		var req messageListRequest
		err := c.BindJSON(&req)
		if err != nil {
			// 400 Bad Request automatically returned by BindWith
			return
		}

		// Get matching messages from Redis

		c.JSON(http.StatusOK, gin.H{})
	})

	r.Run()
}

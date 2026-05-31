package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type PingHandler struct{}

func NewPingHandler() *PingHandler {
	return &PingHandler{}
}

func (h *PingHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong", "timestamp": time.Now().Format(time.RFC3339)})
}

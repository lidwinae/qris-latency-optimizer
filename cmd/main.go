package main

import (
	"qris-latency-optimizer/delivery/handler"
	"github.com/gin-gonic/gin"
)

func init() {
	// database.LoadEnv()
	// database.ConnectDB()
}

func main() {
	r := gin.Default()
	handler.Rest(r)

	r.Run()
}
package main

import (
	"qris-latency-optimizer/delivery/handler"
	"qris-latency-optimizer/repository/database"
	"qris-latency-optimizer/repository/redis"

	"github.com/gin-gonic/gin"
)

func init() {
	database.LoadEnv()
	database.ConnectDB()
	redis.ConnectRedis()
	redis.WarmUpCache()
}

func main() {
	r := gin.Default()
	handler.CorsHandler(r)
	handler.Rest(r)

	r.Run()
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qris-latency-optimizer/delivery/handler"
	"qris-latency-optimizer/repository/database"
	"qris-latency-optimizer/repository/rabbitmq"
	"qris-latency-optimizer/repository/redis"
	"qris-latency-optimizer/usecase/service"
	"qris-latency-optimizer/worker"
	"qris-latency-optimizer/internal/websocket" // TAMBAHKAN

	"github.com/gin-gonic/gin"
)

func main() {
	// --- Startup Sequence ---
	fmt.Println("=== QRIS Latency Optimizer Starting ===")

	// 1. Load environment
	database.LoadEnv()

	// 2. Connect to PostgreSQL + auto-migrate + seed
	database.ConnectDB()
	fmt.Println("✓ PostgreSQL connected & migrated")

	// 3. Connect to Redis + warm cache
	redis.ConnectRedis()
	redis.WarmUpCache()

	// 4. Connect to RabbitMQ
	rabbitmq.ConnectRabbitMQ()
	defer rabbitmq.Close()

	// 5. Initialize WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run() // Run hub dalam goroutine
	fmt.Println("✓ WebSocket Hub initialized")

	// 6. Start the RabbitMQ consumer worker (processes async notifications)
	worker.StartPaymentConsumer()

	// --- HTTP Server ---
	r := gin.Default()
	handler.CorsHandler(r)
	r.Use(service.LatencyTracker()) // Track latency for all API requests
	handler.Rest(r, wsHub) // DIUBAH: pass wsHub ke handler

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		fmt.Println("=== Server running on :8080 ===")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// --- Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n=== Shutting down gracefully ===")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	rabbitmq.Close()
	fmt.Println("=== Shutdown complete ===")
}
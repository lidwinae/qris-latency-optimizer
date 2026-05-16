package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"qris-latency-optimizer/models"
	"qris-latency-optimizer/repository/database"
	"qris-latency-optimizer/repository/rabbitmq"
	"qris-latency-optimizer/repository/redis"
)

// StartPaymentConsumer runs a background goroutine to process async payment confirmations
func StartPaymentConsumer() {
	msgs, err := rabbitmq.Channel.Consume(
		rabbitmq.Queue.Name, // queue
		"",                  // consumer
		true,                // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		log.Fatalf("Failed to register RabbitMQ consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			processStart := time.Now()

			var event map[string]string
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("[Worker] Error unmarshalling message: %v | Body: %s", err, string(d.Body))
				continue
			}

			transactionID := event["transaction_id"]
			if transactionID == "" {
				log.Printf("[Worker] Skipping message with empty transaction_id")
				continue
			}

			// 1. Update status to SUCCESS in PostgreSQL
			if err := database.DB.Model(&models.Transaction{}).
				Where("id = ?", transactionID).
				Update("status", "SUCCESS").Error; err != nil {
				log.Printf("[Worker] Failed to update transaction %s: %v", transactionID, err)
				continue
			}

			// 2. Invalidate Redis cache so the next poll fetches fresh SUCCESS status
			cacheKey := fmt.Sprintf("transaction:%s", transactionID)
			redis.Delete(cacheKey)

			elapsed := time.Since(processStart)
			log.Printf("[Worker] ✓ Confirmed payment %s in %v", transactionID, elapsed)
		}
	}()

	fmt.Println("✓ RabbitMQ Worker is running and waiting for messages...")
}

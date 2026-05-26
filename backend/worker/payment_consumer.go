package worker

import (
	"encoding/json"
	"log"

	"qris-latency-optimizer/repository/rabbitmq"
)

type NotificationPayload struct {
	TransactionID string    `json:"transaction_id"`
	MerchantID    string    `json:"merchant_id"`
	MerchantName  string    `json:"merchant_name"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	Timestamp     string    `json:"timestamp"`
}

// StartPaymentConsumer - start consuming messages dari RabbitMQ
func StartPaymentConsumer() {
	go func() {
		// ✨ DIUBAH: Get notification queue dari package
		channel := rabbitmq.Channel
		if channel == nil {
			log.Println("⚠ RabbitMQ channel not available, consumer not started")
			return
		}

		// ✨ DIUBAH: Use NotificationQueue yang sudah dideklarasi
		q := rabbitmq.GetNotificationQueue()

		msgs, err := channel.Consume(
			q.Name,       // queue name
			"consumer",   // consumer tag
			false,        // auto-ack (manual ack)
			false,        // exclusive
			false,        // no-local
			false,        // no-wait
			nil,          // args
		)
		if err != nil {
			log.Fatalf("❌ Failed to consume: %v", err)
		}

		log.Println("✓ Consumer worker started, listening for notifications...")

		for msg := range msgs {
			var payload NotificationPayload
			err := json.Unmarshal(msg.Body, &payload)
			if err != nil {
				log.Printf("❌ Failed to unmarshal: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Printf("📨 Processing notification [TX: %s, Merchant: %s]", 
				payload.TransactionID, payload.MerchantName)

			// TODO: Push ke WebSocket hub di sini
			// consumer.Hub.SendToMerchant(payload.MerchantID, notification)

			msg.Ack(false) // acknowledge success
		}
	}()
}
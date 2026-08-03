package workers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/TanishValesha/FlashSend-Notifier/internal/db"
	"github.com/TanishValesha/FlashSend-Notifier/internal/models"
	email "github.com/TanishValesha/FlashSend-Notifier/internal/notify/email"
	sms "github.com/TanishValesha/FlashSend-Notifier/internal/notify/sms"
	rabbitmq "github.com/TanishValesha/FlashSend-Notifier/internal/rabbitMQ"
	"github.com/rabbitmq/amqp091-go"
)

func StartEmailWorker() {
	for {
		ch := rabbitmq.GetChannel()
		if ch == nil {
			time.Sleep(3 * time.Second)
			continue
		}

		// Ensure queues exist before consuming (idempotent, safe on reconnect)
		rabbitmq.SetupQueue()

		msgs, err := ch.Consume("email_queue", "", false, false, false, false, nil)
		if err != nil {
			log.Printf("Failed to consume from email_queue: %v, retrying in 3s...", err)
			time.Sleep(3 * time.Second)
			continue
		}

		log.Println("Email worker consuming from email_queue")
		for msg := range msgs {
			processEmailMessage(msg)
		}

		// If we exit the range loop, the channel died; reconnect and re-consume
		log.Println("Email worker channel closed, will reconnect...")
		time.Sleep(3 * time.Second)
	}
}

func processEmailMessage(msg amqp091.Delivery) {
	ch := rabbitmq.GetChannel()
	if ch == nil {
		// Can't ack or publish — nack and requeue
		msg.Nack(false, true)
		log.Println("No RabbitMQ channel available, message requeued")
		return
	}

	var payload rabbitmq.QueueMessage
	json.Unmarshal(msg.Body, &payload)
	log.Printf("Processing email message: %s", msg.Body)

	if payload.NotificationID == 0 {
		log.Println("Invalid notification ID (0)")
		msg.Ack(false)
		return
	}

	var entry models.Notification
	errDB := db.DB.First(&entry, payload.NotificationID).Error
	if errDB != nil {
		log.Println("Notification not found:", payload.NotificationID)
		msg.Ack(false)
		return
	}

	// Idempotency: skip if already in a terminal state
	if entry.Status == models.StatusSent || entry.Status == models.StatusFailed {
		log.Printf("Idempotency: notification %d already %s, skipping", payload.NotificationID, entry.Status)
		msg.Ack(false)
		return
	}

	entry.Status = models.StatusProcessing
	db.DB.Save(&entry)

	err := email.SendEmail(payload.To, payload.Subject, payload.Body)
	if err != nil {
		entry.Retries++
		db.DB.Save(&entry)

		if entry.Retries > 3 {
			entry.Status = models.StatusFailed
			entry.Error = err.Error()
			db.DB.Save(&entry)

			ch.Publish(
				"",
				"email_dlq",
				false,
				false,
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        msg.Body,
				},
			)

			msg.Ack(false)
			return
		}

		// Use DLX-based retry: publish to retry exchange with per-message TTL.
		// RabbitMQ will route it back to email_queue when TTL expires.
		// No in-process goroutine — zero data-loss risk on process restart.
		backoff := time.Second * time.Duration(entry.Retries*entry.Retries)
		nextRetry := time.Now().Add(backoff)

		entry.Status = models.StatusRetrying
		entry.NextAttemptAt = &nextRetry
		entry.Error = err.Error()
		db.DB.Save(&entry)

		if retryErr := rabbitmq.PublishRetry(payload, backoff); retryErr != nil {
			log.Printf("Failed to publish retry for notification %d: %v", payload.NotificationID, retryErr)
			// Fall back: leave message unacked so it requeues
			msg.Nack(false, true)
			return
		}

		msg.Ack(false)
		return
	}

	entry.Status = models.StatusSent
	entry.Error = ""
	db.DB.Save(&entry)
	msg.Ack(false)
}

func StartSMSWorker() {
	for {
		ch := rabbitmq.GetChannel()
		if ch == nil {
			time.Sleep(3 * time.Second)
			continue
		}

		// Ensure queues exist before consuming (idempotent, safe on reconnect)
		rabbitmq.SetupQueue()

		msgs, err := ch.Consume("sms_queue", "", false, false, false, false, nil)
		if err != nil {
			log.Printf("Failed to consume from sms_queue: %v, retrying in 3s...", err)
			time.Sleep(3 * time.Second)
			continue
		}

		log.Println("SMS worker consuming from sms_queue")
		for msg := range msgs {
			processSMSMessage(msg)
		}

		log.Println("SMS worker channel closed, will reconnect...")
		time.Sleep(3 * time.Second)
	}
}

func processSMSMessage(msg amqp091.Delivery) {
	ch := rabbitmq.GetChannel()
	if ch == nil {
		msg.Nack(false, true)
		log.Println("No RabbitMQ channel available, message requeued")
		return
	}

	var payload rabbitmq.QueueMessage
	json.Unmarshal(msg.Body, &payload)
	log.Printf("Processing SMS message: %s", msg.Body)

	if payload.NotificationID == 0 {
		log.Println("Invalid notification ID (0)")
		msg.Ack(false)
		return
	}

	var entry models.Notification
	db.DB.First(&entry, payload.NotificationID)

	// Idempotency: skip if already in a terminal state
	if entry.Status == models.StatusSent || entry.Status == models.StatusFailed {
		log.Printf("Idempotency: notification %d already %s, skipping", payload.NotificationID, entry.Status)
		msg.Ack(false)
		return
	}

	entry.Status = models.StatusProcessing
	db.DB.Save(&entry)

	err := sms.SendSMSWithFailover(payload.To, payload.Body)
	if err != nil {
		entry.Retries++
		db.DB.Save(&entry)

		if entry.Retries > 3 {
			entry.Status = models.StatusFailed
			entry.Error = err.Error()
			db.DB.Save(&entry)

			ch.Publish(
				"",
				"sms_dlq",
				false,
				false,
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        msg.Body,
				},
			)

			msg.Ack(false)
			return
		}

		// Use DLX-based retry: publish to retry exchange with per-message TTL.
		// RabbitMQ will route it back to sms_queue when TTL expires.
		backoff := time.Second * time.Duration(entry.Retries*entry.Retries)
		nextRetry := time.Now().Add(backoff)

		entry.Status = models.StatusRetrying
		entry.NextAttemptAt = &nextRetry
		entry.Error = err.Error()
		db.DB.Save(&entry)

		if retryErr := rabbitmq.PublishRetry(payload, backoff); retryErr != nil {
			log.Printf("Failed to publish retry for notification %d: %v", payload.NotificationID, retryErr)
			msg.Nack(false, true)
			return
		}

		msg.Ack(false)
		return
	}

	entry.Status = models.StatusSent
	entry.Error = ""
	db.DB.Save(&entry)
	msg.Ack(false)
}

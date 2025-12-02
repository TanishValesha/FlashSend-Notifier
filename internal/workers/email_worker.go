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
	msgs, _ := rabbitmq.Ch.Consume("email_queue", "", false, false, false, false, nil)

	for msg := range msgs {
		var payload rabbitmq.QueueMessage
		json.Unmarshal(msg.Body, &payload)
		log.Printf("Messages: %s", msg.Body)

		var entry models.Notification
		db.DB.First(&entry, payload.NotificationID)

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

				rabbitmq.Ch.Publish(
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
				continue
			}

			backoff := time.Second * time.Duration(entry.Retries*entry.Retries)
			nextRetry := time.Now().Add(backoff)

			entry.Status = models.StatusRetrying
			entry.NextAttemptAt = &nextRetry
			entry.Error = err.Error()
			db.DB.Save(&entry)

			go func(payload rabbitmq.QueueMessage, delay time.Duration) {
				time.Sleep(delay)

				body, _ := json.Marshal(payload)

				rabbitmq.Ch.Publish(
					"",
					"email_queue",
					false,
					false,
					amqp091.Publishing{
						ContentType: "application/json",
						Body:        body,
					},
				)
			}(payload, backoff)

			msg.Ack(false)
			continue
		}

		entry.Status = models.StatusSent
		entry.Error = ""
		db.DB.Save(&entry)
		msg.Ack(false)
	}
}

func StartSMSWorker() {
	msgs, _ := rabbitmq.Ch.Consume("sms_queue", "", false, false, false, false, nil)

	for msg := range msgs {
		var payload rabbitmq.QueueMessage
		json.Unmarshal(msg.Body, &payload)
		log.Printf("Messages: %s", msg.Body)

		var entry models.Notification
		db.DB.First(&entry, payload.NotificationID)

		entry.Status = models.StatusProcessing
		db.DB.Save(&entry)

		err := sms.SendSMS(payload.To, payload.Body)
		if err != nil {
			entry.Retries++
			db.DB.Save(&entry)

			if entry.Retries > 3 {
				entry.Status = models.StatusFailed
				entry.Error = err.Error()
				db.DB.Save(&entry)

				rabbitmq.Ch.Publish(
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
				continue
			}

			backoff := time.Second * time.Duration(entry.Retries*entry.Retries)
			nextRetry := time.Now().Add(backoff)

			entry.Status = models.StatusRetrying
			entry.NextAttemptAt = &nextRetry
			entry.Error = err.Error()
			db.DB.Save(&entry)

			go func(payload rabbitmq.QueueMessage, delay time.Duration) {
				time.Sleep(delay)

				body, _ := json.Marshal(payload)

				rabbitmq.Ch.Publish(
					"",
					"sms_queue",
					false,
					false,
					amqp091.Publishing{
						ContentType: "application/json",
						Body:        body,
					},
				)
			}(payload, backoff)

			msg.Ack(false)
			continue
		}

		entry.Status = models.StatusSent
		entry.Error = ""
		db.DB.Save(&entry)
		msg.Ack(false)
	}
}

package workers

import (
	"log"
	"time"

	"github.com/TanishValesha/FlashSend-Notifier/internal/db"
	"github.com/TanishValesha/FlashSend-Notifier/internal/models"
	rabbitmq "github.com/TanishValesha/FlashSend-Notifier/internal/rabbitMQ"
)

func StartScheduledWorker() {
	ticker := time.NewTicker(1 * time.Minute)

	for range ticker.C {
		processDueScheduledNotifications()
	}
}

func processDueScheduledNotifications() {
	now := time.Now()

	var dueNotifications []models.Notification

	err := db.DB.Where("is_scheduled = ? AND scheduled_at <= ? AND status = ?", true, now, models.StatusScheduled).Find(&dueNotifications).Error

	if err != nil {
		log.Println("Scheduler DB error:", err)
		return
	}

	for _, job := range dueNotifications {
		enqueueScheduledJob(&job)
	}
}

func enqueueScheduledJob(job *models.Notification) {
	var entry models.Notification
	db.DB.Where("id = ? AND status = ?", job.ID, models.StatusScheduled).First(&entry)

	entry.Status = models.StatusQueued
	db.DB.Save(&entry)

	var msg rabbitmq.QueueMessage

	switch entry.Channel {
	case "sms":
		msg = rabbitmq.QueueMessage{
			NotificationID:      entry.ID,
			NotificationChannel: rabbitmq.ChannelSMS,
			To:                  entry.To,
			Body:                entry.Body,
		}
	case "email":
		msg = rabbitmq.QueueMessage{
			NotificationID:      entry.ID,
			NotificationChannel: rabbitmq.ChannelEmail,
			To:                  entry.To,
			Subject:             *entry.Subject,
			Body:                entry.Body,
		}
	}

	err := rabbitmq.PublishMessageToQueue(msg)
	if err != nil {
		log.Println("Error enqueueing scheduled job:", err)
		entry.Status = models.StatusFailed
		entry.Error = err.Error()
		db.DB.Save(&entry)
		return
	}

	log.Println("Scheduled job queued:", job.ID)
}

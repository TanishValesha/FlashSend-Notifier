package rabbitmq

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

func PublishMessageToQueue(msg QueueMessage) error {
	switch msg.NotificationChannel {
	case ChannelEmail:
		body, err := json.Marshal(msg)

		if err != nil {
			log.Println("Failed to marshal message:", err)
			return err
		}
		return Ch.Publish(
			"",
			"email_queue", // routing key
			false,
			false,
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
	case ChannelSMS:
		body, _ := json.Marshal(msg)

		return Ch.Publish(
			"",
			"sms_queue", // routing key
			false,
			false,
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
	}

	return nil

}

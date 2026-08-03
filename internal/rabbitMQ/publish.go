package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func PublishMessageToQueue(msg QueueMessage) error {
	ch := GetChannel()
	if ch == nil {
		return amqp091.ErrClosed
	}

	switch msg.NotificationChannel {
	case ChannelEmail:
		body, err := json.Marshal(msg)

		if err != nil {
			log.Println("Failed to marshal message:", err)
			return err
		}
		return ch.Publish(
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

		return ch.Publish(
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

// PublishRetry sends a failed message to the retry queue with a per-message
// TTL (expiration). RabbitMQ will hold it until TTL elapses, then route it
// back to the main queue via the retry queue's dead-letter exchange config.
// This eliminates the data-loss risk from in-process retry goroutines.
func PublishRetry(msg QueueMessage, backoff time.Duration) error {
	ch := GetChannel()
	if ch == nil {
		return amqp091.ErrClosed
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	var exchange, routingKey string
	switch msg.NotificationChannel {
	case ChannelEmail:
		exchange = "email_retry_exchange"
		routingKey = "email_retry"
	case ChannelSMS:
		exchange = "sms_retry_exchange"
		routingKey = "sms_retry"
	default:
		return fmt.Errorf("unknown notification channel: %s", msg.NotificationChannel)
	}

	expirationMs := fmt.Sprintf("%d", backoff.Milliseconds())

	return ch.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Expiration:  expirationMs,
		},
	)
}

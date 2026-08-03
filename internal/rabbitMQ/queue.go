package rabbitmq

import (
	"github.com/rabbitmq/amqp091-go"
)

func SetupQueue() error {
	ch := GetChannel()
	if ch == nil {
		return nil
	}

	queues := []string{"email", "sms"}

	for _, q := range queues {
		mainQueue := q + "_queue"
		dlqQueue := q + "_dlq"
		retryExchange := q + "_retry_exchange"
		retryQueue := q + "_retry_queue"
		retryRoutingKey := q + "_retry"

		// 1. Main queue (durable)
		_, err := ch.QueueDeclare(mainQueue, true, false, false, false, nil)
		if err != nil {
			return err
		}

		// 2. Dead Letter Queue (terminal failures)
		_, err = ch.QueueDeclare(dlqQueue, true, false, false, false, nil)
		if err != nil {
			return err
		}

		// 3. Retry exchange (direct type)
		err = ch.ExchangeDeclare(retryExchange, "direct", true, false, false, false, nil)
		if err != nil {
			return err
		}

		// 4. Retry queue with DLX routing back to the main queue.
		//    Messages published with per-message TTL (expiration) will
		//    sit here, then be routed back to main_queue when TTL expires.
		_, err = ch.QueueDeclare(
			retryQueue,
			true,  // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			amqp091.Table{
				"x-dead-letter-exchange":    "",        // default exchange
				"x-dead-letter-routing-key": mainQueue, // route back to main queue
			},
		)
		if err != nil {
			return err
		}

		// 5. Bind retry queue to retry exchange
		err = ch.QueueBind(retryQueue, retryRoutingKey, retryExchange, false, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

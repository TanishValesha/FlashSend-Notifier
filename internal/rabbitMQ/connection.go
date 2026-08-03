package rabbitmq

import (
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

var (
	conn   *amqp091.Connection
	ch     *amqp091.Channel
	mu     sync.RWMutex
	url    string
	closed bool
)

func InitRabbitMQ(amqpURL string) error {
	url = amqpURL
	return connect()
}

func connect() error {
	mu.Lock()
	defer mu.Unlock()

	var err error
	conn, err = amqp091.Dial(url)
	if err != nil {
		return err
	}

	ch, err = conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	// Register close notifications for both connection and channel
	connClose := conn.NotifyClose(make(chan *amqp091.Error))
	chClose := ch.NotifyClose(make(chan *amqp091.Error))

	log.Println("RabbitMQ connected")

	// Background goroutine that monitors either close signal and triggers reconnect
	go func() {
		select {
		case err := <-connClose:
			if err != nil {
				log.Printf("RabbitMQ connection closed: %v", err)
			}
		case err := <-chClose:
			if err != nil {
				log.Printf("RabbitMQ channel closed: %v", err)
			}
		}

		if closed {
			return
		}

		log.Println("RabbitMQ connection lost, starting reconnection loop...")
		for {
			if closed {
				return
			}
			time.Sleep(3 * time.Second)
			if err := connect(); err != nil {
				log.Printf("RabbitMQ reconnect failed: %v, retrying in 3s...", err)
				continue
			}
			log.Println("RabbitMQ reconnected successfully")
			return
		}
	}()

	return nil
}

// GetChannel returns the current channel in a thread-safe manner.
// Returns nil if not connected.
func GetChannel() *amqp091.Channel {
	mu.RLock()
	defer mu.RUnlock()
	return ch
}

// GetConnection returns the current connection in a thread-safe manner.
func GetConnection() *amqp091.Connection {
	mu.RLock()
	defer mu.RUnlock()
	return conn
}

// IsConnected returns true if the channel is ready.
func IsConnected() bool {
	mu.RLock()
	defer mu.RUnlock()
	return ch != nil && !conn.IsClosed()
}

// IsClosed returns true if Close() has been called (graceful shutdown).
func IsClosed() bool {
	mu.RLock()
	defer mu.RUnlock()
	return closed
}

// Close shuts down the RabbitMQ connection gracefully.
func Close() {
	mu.Lock()
	defer mu.Unlock()

	closed = true

	if ch != nil {
		ch.Close()
	}
	if conn != nil && !conn.IsClosed() {
		conn.Close()
	}
	log.Println("RabbitMQ connection closed")
}

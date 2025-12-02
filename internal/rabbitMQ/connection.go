package rabbitmq

import (
	"log"

	"github.com/rabbitmq/amqp091-go"
)

var Conn *amqp091.Connection
var Ch *amqp091.Channel

func InitRabbitMQ(url string) error {
	var err error

	Conn, err = amqp091.Dial(url)
	if err != nil {
		return err
	}

	Ch, err = Conn.Channel()
	if err != nil {
		return err
	}

	log.Println("RabbitMQ connected")
	return nil
}

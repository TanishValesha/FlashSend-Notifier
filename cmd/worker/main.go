package main

import (
	"log"

	"github.com/TanishValesha/FlashSend-Notifier/internal/config"
	"github.com/TanishValesha/FlashSend-Notifier/internal/db"
	rabbitmq "github.com/TanishValesha/FlashSend-Notifier/internal/rabbitMQ"
	"github.com/TanishValesha/FlashSend-Notifier/internal/workers"
)

func main() {
	config.Load()
	db.Init()
	rabbitmq.InitRabbitMQ(config.Cfg.AMQPURL)

	log.Println("Starting Email Worker...")
	go workers.StartEmailWorker()

	log.Println("Starting SMS Worker...")
	go workers.StartSMSWorker()

	log.Println("Starting Scheduled Notification Worker...")
	go workers.StartScheduledWorker()

	select {}
}

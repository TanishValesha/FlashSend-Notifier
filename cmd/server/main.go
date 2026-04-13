package main

import (
	"log"

	"github.com/TanishValesha/FlashSend-Notifier/internal/config"
	"github.com/TanishValesha/FlashSend-Notifier/internal/db"
	rabbitmq "github.com/TanishValesha/FlashSend-Notifier/internal/rabbitMQ"
	"github.com/TanishValesha/FlashSend-Notifier/internal/router"
)

func main() {
	config.Load()

	db.Init()
	db.CreateEnums()
	db.AutoMigrate()

	rabbitmq.InitRabbitMQ(config.Cfg.AMQPURL)
	rabbitmq.SetupQueue()

	router := router.Init()

	address := config.Cfg.BindAddr
	log.Printf("Server Runinng on %s", address)

	err := router.Run(address)
	if err != nil {
		log.Fatalf("Error Running Server: %s", err)
	}
}

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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

	log.Println("All workers started. Waiting for shutdown signal...")

	// Block until shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Received signal %s, shutting down workers...", sig)

	// Close RabbitMQ connection to stop consuming
	rabbitmq.Close()

	log.Println("All workers stopped. Exiting.")
}

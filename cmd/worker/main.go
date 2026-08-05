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
	if err := rabbitmq.InitRabbitMQ(config.Cfg.AMQPURL); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	if err := rabbitmq.SetupQueue(); err != nil {
		log.Fatalf("Failed to setup queues: %v", err)
	}

	log.Println("Starting Email Worker...")
	go workers.StartEmailWorker()

	log.Println("Starting SMS Worker...")
	go workers.StartSMSWorker()

	log.Println("Starting Scheduled Notification Worker...")
	go workers.StartScheduledWorker()

	log.Println("Starting Health Server...")
	workers.StartHealthServer()

	log.Println("All workers started. Waiting for shutdown signal...")

	// Block until shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Received signal %s, shutting down workers...", sig)

	// Close RabbitMQ connection to stop consuming
	rabbitmq.Close()

	// Shut down the health server
	workers.ShutdownHealthServer()

	log.Println("All workers stopped. Exiting.")
}

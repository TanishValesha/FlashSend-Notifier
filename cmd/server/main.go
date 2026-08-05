package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TanishValesha/FlashSend-Notifier/internal/config"
	"github.com/TanishValesha/FlashSend-Notifier/internal/db"
	rabbitmq "github.com/TanishValesha/FlashSend-Notifier/internal/rabbitMQ"
	"github.com/TanishValesha/FlashSend-Notifier/internal/router"
)

// Injected at build time via ldflags
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	config.Load()

	db.Init()
	db.CreateEnums()
	db.AutoMigrate()

	if err := rabbitmq.InitRabbitMQ(config.Cfg.AMQPURL); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	if err := rabbitmq.SetupQueue(); err != nil {
		log.Fatalf("Failed to setup queues: %v", err)
	}

	r := router.Init(Version, BuildTime)

	srv := &http.Server{
		Addr:         config.Cfg.BindAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so we can listen for shutdown signals
	go func() {
		log.Printf("Server running on %s (version=%s built=%s)", config.Cfg.BindAddr, Version, BuildTime)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %s", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Received signal %s, shutting down gracefully...", sig)

	// Give outstanding requests 15 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}

	log.Println("Server exited cleanly")
}

package workers

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	rabbitmq "github.com/TanishValesha/FlashSend-Notifier/internal/rabbitMQ"
)

var healthServer *http.Server

func StartHealthServer() {
	addr := os.Getenv("HEALTH_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if !rabbitmq.IsConnected() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","rabbitmq":"unhealthy"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","rabbitmq":"ok"}`))
	})

	healthServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		log.Printf("Worker health server listening on %s", addr)
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Worker health server error: %v", err)
		}
	}()
}

func ShutdownHealthServer() {
	if healthServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := healthServer.Shutdown(ctx); err != nil {
			log.Printf("Worker health server shutdown error: %v", err)
		}
	}
}

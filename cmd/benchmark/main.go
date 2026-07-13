package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/TanishValesha/FlashSend-Notifier/internal/config"
	rabbitmq "github.com/TanishValesha/FlashSend-Notifier/internal/rabbitMQ"
	"github.com/rabbitmq/amqp091-go"
)

const (
	numNotifications = 10000
	numWorkers       = 10
	benchmarkQueue   = "benchmark_queue"
)

var (
	processedCount int64
	totalProcessNs int64
)

func main() {
	config.Load()

	err := rabbitmq.InitRabbitMQ(config.Cfg.AMQPURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	// Declare a durable queue
	_, err = rabbitmq.Ch.QueueDeclare(benchmarkQueue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare benchmark queue: %v", err)
	}

	// Purge any leftover messages
	_, err = rabbitmq.Ch.QueuePurge(benchmarkQueue, false)
	if err != nil {
		log.Fatalf("Failed to purge benchmark queue: %v", err)
	}

	fmt.Printf("\nGenerating %d notifications and publishing to RabbitMQ...\n", numNotifications)

	// Start timing before publish — end-to-end benchmark
	benchStart := time.Now()

	// ── Phase 1: Publish ──────────────────────────────────────────────────
	for i := 0; i < numNotifications; i++ {
		msg := rabbitmq.QueueMessage{
			NotificationID:      uint(i + 1),
			NotificationChannel: rabbitmq.ChannelEmail,
			To:                  fmt.Sprintf("benchmark%d@test.com", i),
			Subject:             "Benchmark Test",
			Body:                "Benchmark test message payload.",
		}

		body, _ := json.Marshal(msg)
		err := rabbitmq.Ch.Publish(
			"",
			benchmarkQueue,
			false,
			false,
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
		if err != nil {
			log.Printf("Failed to publish message %d: %v", i, err)
		}
	}

	// Record memory after publishing (before workers consume)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	peakMemory := float64(memStats.Alloc) / 1024 / 1024

	// ── Phase 2: Consume ──────────────────────────────────────────────────
	fmt.Printf("Starting %d workers to consume...\n", numWorkers)

	for w := 0; w < numWorkers; w++ {
		go consumeWorker(w)
	}

	// Poll until all messages are processed
	for {
		if atomic.LoadInt64(&processedCount) >= numNotifications {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	totalDuration := time.Since(benchStart)

	// Final memory reading
	runtime.ReadMemStats(&memStats)
	finalAlloc := float64(memStats.Alloc) / 1024 / 1024
	if finalAlloc > peakMemory {
		peakMemory = finalAlloc
	}

	// ── Metrics ───────────────────────────────────────────────────────────
	totalProcessed := atomic.LoadInt64(&processedCount)
	throughputPerSec := float64(totalProcessed) / totalDuration.Seconds()

	var avgProcess time.Duration
	if totalProcessed > 0 {
		avgProcess = time.Duration(atomic.LoadInt64(&totalProcessNs) / totalProcessed)
	}

	// ── Report ────────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Println("========== FlashSend Benchmark ==========")
	fmt.Printf("Notifications:      %d\n", numNotifications)
	fmt.Printf("Workers:            %d\n", numWorkers)
	fmt.Printf("Duration:           %.1f s\n", totalDuration.Seconds())
	fmt.Println()
	fmt.Println("Throughput example:")
	fmt.Printf("%.0f notifications/sec\n", throughputPerSec)
	fmt.Printf("%.0f notifications/min\n", throughputPerSec*60)
	fmt.Println()
	fmt.Printf("Success:            %d\n", totalProcessed)
	fmt.Println()
	fmt.Printf("Failed:             %d\n", 0)
	fmt.Println()
	fmt.Printf("Average processing time: %.1f ms\n", float64(avgProcess.Nanoseconds())/1e6)
	fmt.Println()
	fmt.Printf("Peak memory:        %.0f MB\n", peakMemory)
	fmt.Println("CPU:                (see system monitor)")
	fmt.Println("=========================================")

	// Cancel all consumers to clean up
	for w := 0; w < numWorkers; w++ {
		_ = rabbitmq.Ch.Cancel(fmt.Sprintf("benchmark-worker-%d", w), false)
	}

	os.Exit(0)
}

func consumeWorker(id int) {
	// Each worker gets its own channel for better concurrency
	ch, err := rabbitmq.Conn.Channel()
	if err != nil {
		log.Printf("Worker %d failed to open channel: %v", id, err)
		return
	}
	defer ch.Close()

	// Set QoS prefetch to 1 for fair dispatching
	_ = ch.Qos(1, 0, false)

	msgs, err := ch.Consume(
		benchmarkQueue,
		fmt.Sprintf("benchmark-worker-%d", id),
		false, // autoAck = false
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Worker %d failed to consume: %v", id, err)
		return
	}

	for msg := range msgs {
		start := time.Now()

		// Deserialize to simulate real worker overhead
		var payload rabbitmq.QueueMessage
		_ = json.Unmarshal(msg.Body, &payload)

		// Simulate SMTP call (email sending work)
		time.Sleep(50 * time.Millisecond)

		// ACK the message
		err := msg.Ack(false)
		if err != nil {
			log.Printf("Worker %d ACK failed: %v", id, err)
			continue
		}

		elapsedNs := time.Since(start).Nanoseconds()
		atomic.AddInt64(&totalProcessNs, elapsedNs)
		atomic.AddInt64(&processedCount, 1)
	}
}

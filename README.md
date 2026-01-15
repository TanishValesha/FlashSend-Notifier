# FlashSend — Distributed Notification Platform

FlashSend is a **Go-based distributed notification delivery system** designed to reliably send SMS, Email, and scheduled notifications at scale. It uses a **RabbitMQ-powered asynchronous worker architecture** with retries, dead-letter queues (DLQ), DB-backed scheduling, and automatic SMS provider failover (Twilio → Vonage) to guarantee delivery even during traffic spikes and third-party outages.

---

## Features

* Send SMS and Email notifications
* Schedule notifications for future delivery
* RabbitMQ asynchronous processing
* Exponential retry pipelines
* Dead Letter Queues (DLQ) for fault isolation
* Automatic SMS provider failover (Twilio → Vonage)
* Horizontally scalable distributed workers
* JWT authentication & API-key based access
* PostgreSQL-backed persistence
* Status-driven notification lifecycle
* Human-readable 24-hour time scheduling

---

## System Flow

```
Client → API (Gin) → RabbitMQ Queue → Distributed Workers → Providers (SMS/Email)
                                  ↘ DLQ (failed messages)
```

Scheduled notifications are stored in PostgreSQL and scanned by a cron-like scheduler built using Go’s `time.Ticker`.

---

## ⚙ Tech Stack

| Layer         | Technology               |
| ------------- | ------------------------ |
| Language      | Go                       |
| API Framework | Gin                      |
| Queue         | RabbitMQ                 |
| Database      | PostgreSQL               |
| SMS Providers | Twilio, Vonage           |
| Auth          | JWT                      |
| Scheduler     | Goroutines + time.Ticker |

---

## Notification Lifecycle

```
queued → processing → retrying → sent → dead
```

Retries use exponential backoff. Notifications that exceed retry limits are routed to DLQ.

---

## Setup

### 1️⃣ Clone Repository

```bash
git clone https://github.com/TanishValesha/FlashSend-Notifier
cd FlashSend-Notifier
```

---

### 2️⃣ Environment Variables

Create a `.env` file:

```env
BIND_ADDR=:8080

JWT_SECRET=your-jwt-secret
APIKEY_HMAC_SECRET=your-apikey-secret
JWT_EXPIRATION_HOURS=24

DATABASE_URL=postgres://user:password@localhost:5432/flashsend?sslmode=disable

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=your-email@gmail.com
SMTP_APP_PASSWORD=your-smtp-app-password

TWILIO_ACCOUNT_SID=your_twilio_sid
TWILIO_AUTH_TOKEN=your_twilio_auth_token
TWILIO_PHONE_NUMBER=your_twilio_phone

VONAGE_API_KEY=your_vonage_api_key
VONAGE_API_SECRET=your_vonage_api_secret
VONAGE_FROM=FlashSend
```

---

### 3️⃣ Start Dependencies

```bash
docker-compose up -d postgres rabbitmq
```

---

### 4️⃣ Run Server

```bash
go run ./cmd/server/main.go
```

### 5️⃣ Run Workers

```bash
go run ./cmd/worker/main.go
```
---

# API Documentation

## Authentication APIs

### Register

`POST /api/auth/register`

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response**

```json
{
  "message": "User registered successfully"
}
```

---

### Login

`POST /api/auth/login`

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response**

```json
{
  "token": "<jwt-token>"
}
```

---

## API Key Management

### Create API Key

`POST /api/apikeys`

**Response**

```json
{
  "key": "fs_98af3c1..."
}
```

---

### List API Keys

`GET /api/apikeys`

---

### Delete API Key

`DELETE /api/apikeys/:id`

---

## Send Notification

### Send Instant Notification

`POST /api/notify/send`

```json
{
  "channel": "sms",
  "to": "+919999999999",
  "subject": "optional for email",
  "body": "Hello from FlashSend"
}
```

**Response**

```json
{
  "message": "Notification queued",
  "id": 12
}
```

---

## Schedule Notification

### Schedule Notification

`POST /api/notify/schedule`

```json
{
  "channel": "sms",
  "to": "+919999999999",
  "subject": "optional",
  "body": "Reminder",
  "scheduleAt": "2025-02-20 18:30"
}
```

**Response**

```json
{
  "message": "Scheduled notification created",
  "id": 34
}
```

---

## Notification History

### List Notifications

`GET /api/notifications`

---

### Get Single Notification

`GET /api/notifications/:id`

---

## Health Check

`GET /ping`

---


## Why FlashSend?

FlashSend demonstrates real-world **distributed systems design patterns** such as:

* Competing consumer worker pools
* Queue-based backpressure handling
* Fault-tolerant retry pipelines
* Scheduled job orchestration
* Provider failover and resiliency design

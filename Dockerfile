# ---- Stage 1: Build ----
FROM golang:1.24-alpine AS builder

ARG BINARY=server

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .

ARG VERSION=dev
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o /app/bin/flashsend \
    ./cmd/${BINARY}

# ---- Stage 2: Runtime ----
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/bin/flashsend /app/flashsend

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/flashsend"]
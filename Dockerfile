# Gunakan versi Go terbaru atau install air versi yang kompatibel
FROM golang:1.23-alpine AS base

# Add essential build tools
RUN apk add --no-cache gcc musl-dev make git

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Development stage
FROM base AS development

# Install Air versi yang kompatibel dengan Go 1.23
RUN go install github.com/air-verse/air@v1.52.3

# Copy the entire project
COPY . .

# Command to run Air
CMD ["air", "-c", ".air.toml"]

# Builder stage
FROM base AS builder

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Production stage
FROM alpine:3.18 AS production

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/main .


CMD ["./main"]
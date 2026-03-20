FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/server ./server
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./server"]

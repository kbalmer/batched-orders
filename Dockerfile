FROM golang:alpine3.21 AS builder

RUN CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o batched-orders ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/batched-orders .

RUN chmod +x batched-orders

EXPOSE 8000

CMD ["./batched-orders"]
FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o logging-service ./cmd/logging-service

EXPOSE 8084

CMD ["./logging-service"]

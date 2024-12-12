FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ads-integration-service ./cmd/ads-integration-service

EXPOSE 8081

CMD ["./ads-integration-service"]

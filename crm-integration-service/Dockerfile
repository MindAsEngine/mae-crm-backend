FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o crm-integration-service ./cmd/crm-integration-service

EXPOSE 8082

CMD ["./crm-integration-service"]

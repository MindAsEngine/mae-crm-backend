package workers

import (
	interfaces "ads-integration-service/internal/services/interfaces"
	"log"
	"strconv"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
    Channel *amqp.Channel
}

func InitRabbitMQ() *RabbitMQ {
    conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
 
    ch, err := conn.Channel()
    if err != nil {
        log.Fatalf("Failed to open a channel: %v", err)
    }

    return &RabbitMQ{Channel: ch}
}

func StartLoader(mq *RabbitMQ, handler interfaces.TaskHandler) {
    msgs, err := mq.Channel.Consume(
        "UploadMessages", // Queue
        "",      // Consumer
        true,    // Auto-ack
        false,   // Exclusive
        false,   // No-local
        false,   // No-wait
        nil,     // Args
    )
    if err != nil {
        log.Fatalf("Failed to register consumer: %v", err)
    }

    for msg := range msgs {
        IntegrationID, err := strconv.Atoi(string(msg.Body))
        if err != nil {
            log.Printf("Invalid UploadMsg ID: %s", msg.Body)
            continue
        }

        log.Printf("Received UploadMsg ID: %d", IntegrationID)
        if err := handler.ProcessUploadMsg(IntegrationID); err != nil {
            log.Printf("Failed to process UploadMsg %d: %v", IntegrationID, err)
        }
    }
}

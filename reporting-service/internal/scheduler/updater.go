package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"reporting-service/internal/repository"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

// AudienceUpdateTask представляет задачу на обновление аудитории
type AudienceUpdateTask struct {
    AudienceID uuid.UUID `json:"audience_id"`
    Timestamp  time.Time `json:"timestamp"`
}

type AudienceUpdateService struct {
    repo           *repository.AudienceRepo
    rabbitChannel  *amqp.Channel
    rabbitQueue    amqp.Queue
}

func NewAudienceUpdateService(
    repo *repository.AudienceRepo, 
    ch *amqp.Channel,
) (*AudienceUpdateService, error) {
    // Объявляем очередь
    q, err := ch.QueueDeclare(
        "audience_updates", // имя очереди
        true,               // долговечность
        false,              // автоудаление
        false,              // эксклюзивность
        false,              // без ожидания
        nil,                // аргументы
    )
    if err != nil {
        return nil, fmt.Errorf("failed to declare RabbitMQ queue: %w", err)
    }

    return &AudienceUpdateService{
        repo:           repo,
        rabbitChannel:  ch,
        rabbitQueue:    q,
    }, nil
}

// CreateDailyUpdateTasks создает задачи на обновление для всех аудиторий
func (s *AudienceUpdateService) CreateDailyUpdateTasks(ctx context.Context) error {
    // Находим все существующие аудитории
    query := `SELECT id FROM audiences`
    var audienceIDs []uuid.UUID

    err := s.repo.DB.SelectContext(ctx, &audienceIDs, query)
    if err != nil {
        return fmt.Errorf("failed to fetch audiences: %w", err)
    }

    // Создаем и отправляем задачи для каждой аудитории
    for _, audienceID := range audienceIDs {
        task := AudienceUpdateTask{
            AudienceID: audienceID,
            Timestamp:  time.Now(),
        }

        // Сериализуем задачу
        body, err := json.Marshal(task)
        if err != nil {
            log.Printf("Failed to marshal task for audience %v: %v", audienceID, err)
            continue
        }

        // Отправляем задачу в RabbitMQ
        err = s.rabbitChannel.Publish(
            "",                     // exchange
            s.rabbitQueue.Name,     // routing key
            false,                  // mandatory
            false,                  // immediate
            amqp.Publishing{
                ContentType: "application/json",
                Body:        body,
            })

        if err != nil {
            log.Printf("Failed to publish task for audience %v: %v", audienceID, err)
        }
    }

    return nil
}

// StartDailyUpdateScheduler запускает ежедневное создание задач
func (s *AudienceUpdateService) StartDailyUpdateScheduler(ctx context.Context) {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            err := s.CreateDailyUpdateTasks(ctx)
            if err != nil {
                log.Printf("Daily update task creation failed: %v", err)
            }
        case <-ctx.Done():
            return
        }
    }
}

// WorkerForUpdateTasks воркер для обработки задач обновления аудиторий
func (s *AudienceUpdateService) WorkerForUpdateTasks(ctx context.Context) error {
    msgs, err := s.rabbitChannel.Consume(
        s.rabbitQueue.Name, // queue
        "",                 // consumer
        false,              // auto-ack
        false,              // exclusive
        false,              // no-local
        false,              // no-wait
        nil,                // args
    )
    if err != nil {
        return fmt.Errorf("failed to register consumer: %w", err)
    }

    for msg := range msgs {
        var task AudienceUpdateTask
        err := json.Unmarshal(msg.Body, &task)
        if err != nil {
            log.Printf("Failed to unmarshal task: %v", err)
            msg.Nack(false, false) // отрицательное подтверждение
            continue
        }

        // Обновляем аудиторию
        err = s.repo.UpdateAudienceRequests(ctx, task.AudienceID)
        if err != nil {
            log.Printf("Failed to update audience %v: %v", task.AudienceID, err)
            msg.Nack(false, true) // отрицательное подтверждение с возвратом в очередь
        } else {
            msg.Ack(false) // успешное подтверждение
        }
    }

    return nil
}
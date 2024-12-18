package audience

import (
	"context"
	"encoding/json"
	//"encoding/json"
	"fmt"
	"time"

	//"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go" // Fixed import
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"reporting-service/internal/domain"
	MysqlRepo "reporting-service/internal/repository/mysql"
	PostgreRepo "reporting-service/internal/repository/postgre"
)

const (
    exchangeName = "audiences"
    queueName    = "audience.updates"
)

type AudienceUpdateService struct {
    audienceRepo PostgreRepo.PostgresAudienceRepository
    mysqlRepo    MysqlRepo.MySQLAudienceRepository
    rabbitMQ     *amqp.Channel
    logger       *zap.Logger
    scheduler    *cron.Cron
}

type MessagePublisher interface {
    Publish(ctx context.Context, routingKey string, msg interface{}) error
}

type AudienceMessage struct {
    AudienceID      int64           `json:"audience_id"`
    UpdatedAt       time.Time       `json:"updated_at"`
    RequestCount    int             `json:"request_count"`
    LastRequestID   int64           `json:"last_request_id,omitempty"`
}

func NewAudienceUpdateService(
    audienceRepo PostgreRepo.PostgresAudienceRepository,
    mysqlRepo MysqlRepo.MySQLAudienceRepository,
    rabbitMQ *amqp.Channel,
    logger *zap.Logger,
) *AudienceUpdateService {
    return &AudienceUpdateService{
        audienceRepo: audienceRepo,
        mysqlRepo:    mysqlRepo,
        rabbitMQ:     rabbitMQ,
        logger:       logger.With(zap.String("service", "audience_updater")),
        scheduler:    cron.New(cron.WithLocation(time.UTC)),
    }
}

func (s *AudienceUpdateService) Start(ctx context.Context) error {
    _, err := s.scheduler.AddFunc("0 1 * * *", func() {
        if err := s.processAllAudiences(ctx); err != nil {
            s.logger.Error("failed to process audiences", zap.Error(err))
        }
    })
    if err != nil {
        return fmt.Errorf("failed to schedule updates: %w", err)
    }

    s.scheduler.Start()
    s.logger.Info("audience updater started")

    <-ctx.Done()
    s.scheduler.Stop()
    return nil
}

func (s *AudienceUpdateService) processAllAudiences(ctx context.Context) error {
    audiences, err := s.audienceRepo.List(ctx)
    if err != nil {
        return fmt.Errorf("list audiences: %w", err)
    }

    for _, audience := range audiences {
        if err := s.processAudience(ctx, &audience); err != nil {
            s.logger.Error("process audience failed",
                zap.String("audience_id", string(audience.ID)),
                zap.Error(err))
            continue
        }
    }
    return nil
}

func (s *AudienceUpdateService) processAudience(ctx context.Context, audience *domain.Audience) error {
    requests, err := s.mysqlRepo.GetApplicationsByAudienceFilter(ctx, audience.Filter)
    if err != nil {
        return fmt.Errorf("get requests: %w", err)
    }

    if err := s.audienceRepo.UpdateApplication(ctx, audience.ID, requests); err != nil {
        return fmt.Errorf("update requests: %w", err)
    }

    // Prepare message for RabbitMQ
    msg := domain.AudienceMessage{
        AudienceID:    audience.ID,
        Applications: requests,
        Filter:       audience.Filter,
    }

    // Serialize message
    body, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("marshal message: %w", err)
    }

    // Publish to RabbitMQ
    err = s.rabbitMQ.PublishWithContext(ctx,
        exchangeName,
        queueName,
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:       body,
        })
    if err != nil {
        return fmt.Errorf("publish message: %w", err)
    }

    s.logger.Info("audience processed and message published",
        zap.String("audience_id", string(audience.ID)),
        zap.Int("request_count", len(requests)))

    return nil
}
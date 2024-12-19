package audience

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"reporting-service/internal/domain"
	MysqlRepo "reporting-service/internal/repository/mysql"
	PostgreRepo "reporting-service/internal/repository/postgre"
)

type Service struct {
	audienceRepo PostgreRepo.PostgresAudienceRepository
	mysqlRepo    MysqlRepo.MySQLAudienceRepository
	logger       *zap.Logger
	amqpChan     *amqp.Channel
	config       Config
	exporter     *ExcelExporter
}

type Config struct {
	UpdateTime string `yaml:"update_time"`
	BatchSize  int    `yaml:"batch_size"`
	ExportPath string `yaml:"export_path"`
}

func NewService(
	cfg Config,
	mysqlRepo *MysqlRepo.MySQLAudienceRepository,
	audienceRepo *PostgreRepo.PostgresAudienceRepository,
	amqpChan *amqp.Channel,
	logger *zap.Logger,) *Service {
	s := &Service{
		audienceRepo: *audienceRepo,
		mysqlRepo:    *mysqlRepo,
		amqpChan:     amqpChan,
		logger:       logger,
		config:       cfg,
		exporter:     NewExcelExporter(*audienceRepo, *mysqlRepo, logger),
	}

	if err := s.setupRabbitMQ(); err != nil {
		logger.Fatal("Failed to setup RabbitMQ", zap.Error(err))
	}

	return s
}

func (s *Service) setupRabbitMQ() error {
	err := s.amqpChan.ExchangeDeclare(
		"audiences", // name
		"direct",    // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	_, err = s.amqpChan.QueueDeclare(
		"audience.updates", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	return s.amqpChan.QueueBind(
		"audience.updates", // queue name
		"audience.updates", // routing key
		"audiences",        // exchange
		false,
		nil,
	)
}

func (s *Service) GetById(ctx context.Context, id int64) (*domain.AudienceResponse, error) {
	audience, err := s.audienceRepo.GetByID(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("get audience: %w", err)
	}

	var response = domain.AudienceResponse{
		ID:           audience.ID,
		Name:         audience.Name,
		Integrations: audience.Integrations,
		CreatedAt:    audience.CreatedAt,
		UpdatedAt:    audience.UpdatedAt,
	}

	return &response, nil
}

func (s *Service) List(ctx context.Context) ([]domain.AudienceResponse, error) {
	audiences, err := s.audienceRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("get audiences: %w", err)
	}

	var response []domain.AudienceResponse
	for _, a := range audiences {
		response = append(response, domain.AudienceResponse{
			ID:           a.ID,
			Name:         a.Name,
			Integrations: a.Integrations,
			CreatedAt:    a.CreatedAt,
			UpdatedAt:    a.UpdatedAt,
		})
	}
	return response, nil
}

func (s *Service) CreateIntegrations(ctx context.Context, req domain.IntegrationsCreateRequest) (*domain.IntegrationsCreateResponse, error) {
	integrations := make([]domain.Integration, 0, len(req.AudienceIds))
	for _, id := range req.AudienceIds {
		integration := &domain.Integration{
			AudienceID:  id,
			CabinetName: req.CabinetName,
		}
		s.audienceRepo.CreateIntegration(ctx, integration, id)
		integrations = append(integrations, *integration)
	}
	return &domain.IntegrationsCreateResponse{
		Integrations: integrations,
	}, nil
}

func (s *Service) Create(ctx context.Context, req domain.AudienceCreateRequest) (*domain.AudienceResponse, error) {
	audience := &domain.Audience{
		Name:      req.Name,
		Filter:    req.Filter,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	applications, err := s.mysqlRepo.GetApplicationsByAudienceFilter(ctx, req.Filter)

	if err != nil {
		return nil, fmt.Errorf("get applications: %w", err)
	}

	audience.Applications = applications

	if err := s.audienceRepo.Create(ctx, audience); err != nil {
		return nil, fmt.Errorf("create audience: %w", err)
	}

	return &domain.AudienceResponse{
		ID:           audience.ID,
		Name:         audience.Name,
		Integrations: audience.Integrations,
		CreatedAt:    audience.CreatedAt,
		UpdatedAt:    audience.UpdatedAt,
	}, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.audienceRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete audience: %w", err)
	}
	return nil
}

func (s *Service) Export(ctx context.Context, id int64) (string, error) {
	return s.exporter.ExportAudience(ctx, id)
}

func (s *Service) DisconnectAll(ctx context.Context, id int64) error {
	if err := s.audienceRepo.RemoveAllIntegrations(ctx, id); err != nil {
		return fmt.Errorf("remove integrations: %w", err)
	}
	return nil
}

func (s *Service) UpdateAudience(ctx context.Context, id int64, application_ids []int64) error {
	audience, err := s.audienceRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get audience: %w", err)
	}

	requests, err := s.mysqlRepo.GetNewApplicationsByAudience(ctx, audience, application_ids)
	if err != nil {
		return fmt.Errorf("get requests: %w", err)
	}

	if err := s.audienceRepo.UpdateApplicationsForAudience(ctx, id, requests); err != nil {
		return fmt.Errorf("update requests: %w", err)
	}

	return nil
}

func (s *Service) ProcessAllAudiences(ctx context.Context) error {
	audiences, err := s.audienceRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list audiences: %w", err)
	}
	for _, audience := range audiences {
		s.logger.Info("processing all audiences", zap.Int64("audience:", audience.ID))

		//Получаем фильтр по аудитории
		filter, err := s.audienceRepo.GetFilterByAudienceId(ctx, audience.ID)
		if err != nil {
			s.logger.Error("get filter by audience id failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}
		s.logger.Info("filter", zap.Any("filter", filter))

		audience.Filter = domain.AudienceFilter{
			CreationDateFrom:     filter.CreationDateFrom,
			CreationDateTo:       filter.CreationDateTo,
			StatusNames:          filter.StatusNames,
			RegectionReasonNames: filter.RegectionReasonNames,
			NonTargetReasonNames: filter.NonTargetReasonNames,
		}

		//Получаем текущие заявки по аудитории
		current_applications, err := s.audienceRepo.GetApplicationIdsByAdienceId(ctx, audience.ID)
		if err != nil {
			s.logger.Error("get applications by audience filter failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}
		s.logger.Info("current applications", zap.Any("current_applications", current_applications))

		//Получаем заявки, которые изменили статус
		changed_applications, err := s.mysqlRepo.GetChangedApplicationIds(ctx, &audience.Filter, current_applications)
		if err != nil {
			s.logger.Error("get changed applications failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}
		s.logger.Info("changed applications", zap.Any("changed_applications", changed_applications))

		//Удаляем заявки с измененными статусами
		if err := s.audienceRepo.DeleteApplications(ctx, audience.ID, changed_applications); err != nil {
			s.logger.Error("delete applications with changed statuses failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}
		s.logger.Info("applications deleted", zap.Any("applications", changed_applications))

		//Получаем заявки которые не изменили статус
		current_applications, err = s.audienceRepo.GetApplicationIdsByAdienceId(ctx, audience.ID)
		if err != nil {
			s.logger.Error("get changed applications failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}
		s.logger.Info("current applications", zap.Any("current_applications", current_applications))

		//Получаем обновленные заявки которые ещё не в аудитории
		requests, err := s.mysqlRepo.GetNewApplicationsByAudience(ctx, &audience, current_applications)
		if err != nil {
			s.logger.Error("get requests: ", zap.Error(err))
			continue
		}

		if requests != nil {
			if err := s.audienceRepo.UpdateApplicationsForAudience(ctx, audience.ID, requests); err != nil {
				s.logger.Error("update requests: ", zap.Error(err))
				continue
			}
		}

		//TODO: change on production
		//if requests == nil && changed_applications == nil {
		//	s.logger.Info("no changed or new requests found so nothing pushed to rabbit", zap.Any("audience_id", audience.ID))
		//} else {
			req_ids, err := s.audienceRepo.GetApplicationIdsByAdienceId(ctx, audience.ID)
		
			if err != nil {
				s.logger.Error("get application ids by audience id: ", zap.Error(err))
				continue
			}
			
			if len(req_ids) == 0 {
				s.logger.Info("no requests found", zap.Any("audience_id", audience.ID))
				continue
			}
	
			requests, err = s.mysqlRepo.ListApplicationsByIds(ctx, req_ids)
			if err != nil {
				s.logger.Error("get audience: ", zap.Error(err))
				continue
			}
	
			audience.Applications = requests
	
			if err := s.pushAudienceToRabbit(ctx, &audience); err != nil {
				s.logger.Error("process audience failed",
					zap.String("audience_id", string(audience.ID)),
					zap.Error(err))
				continue
			}
		//}
	}
	return nil
}

func (s *Service) pushAudienceToRabbit(ctx context.Context, audience *domain.Audience) error {
	var lastRequestId int64
	if len(audience.Applications) > 0 {
		lastRequestId = audience.Applications[len(audience.Applications)-1].ID
	}

	message := domain.AudienceMessage{
		AudienceID:   audience.ID,
		Applications: audience.Applications,
		Filter:       audience.Filter,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	err = s.amqpChan.PublishWithContext(
		ctx,
		"audiences",        // exchange
		"audience.updates", // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			MessageId:    fmt.Sprintf("%d-%d", audience.ID, time.Now().Unix()),
			DeliveryMode: amqp.Persistent,
		},
	)

	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	s.logger.Info("audience update message published",
		zap.Int64("audience_id", audience.ID),
		zap.Int("request_count", len(audience.Applications)),
		zap.Int64("last_request_id", lastRequestId))

	return nil
}

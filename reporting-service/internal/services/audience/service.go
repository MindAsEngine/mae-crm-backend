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
	rabbitMQ     *amqp.Channel
	exporter     *ExcelExporter
	config       Config
}

type Config struct {
	UpdateTime string `yaml:"update_time"`
	BatchSize  int    `yaml:"batch_size"`
	ExportPath string `yaml:"export_path"`
}

func NewService(cfg Config,
	mysqlRepo *MysqlRepo.MySQLAudienceRepository,
	audienceRepo *PostgreRepo.PostgresAudienceRepository,
	logger *zap.Logger) *Service {

	return &Service{
		audienceRepo: *audienceRepo,
		mysqlRepo:    *mysqlRepo,
		logger:       logger,
		exporter:     NewExcelExporter(*audienceRepo, logger),
		config:       cfg,
	}
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

	if err := s.audienceRepo.UpdateApplication(ctx, id, requests); err != nil {
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
		filter, err := s.audienceRepo.GetFilterByAudienceId(ctx, audience.ID)
		if err != nil {
			s.logger.Error("get filter by audience id failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}
		audience.Filter = domain.AudienceFilter{
			CreationDateFrom:     filter.CreationDateFrom,
			CreationDateTo:       filter.CreationDateTo,
			StatusNames:          filter.StatusNames,
			RegectionReasonNames: filter.RegectionReasonNames,
			NonTargetReasonNames: filter.NonTargetReasonNames,
		}
		current_applications, err := s.audienceRepo.GetApplicationIdsByAdienceId(ctx, audience.ID)
		if err != nil {
			s.logger.Error("get applications by audience filter failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue	
		}
		changed_applications, err := s.mysqlRepo.GetChangedApplicationIds(ctx, &audience.Filter, current_applications)

		if err != nil {
			s.logger.Error("get changed applications failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}

		if err := s.audienceRepo.DeleteApplications(ctx, audience.ID, changed_applications); err!=nil{
			s.logger.Error("delete applications with changed statuses failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}

		if err := s.processAudience(ctx, &audience, current_applications); err != nil {
			s.logger.Error("process audience failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}
	}
	return nil
}



func (s *Service) processAudience(ctx context.Context, audience *domain.Audience, application_ids []int64) error {
	requests, err := s.mysqlRepo.GetNewApplicationsByAudience(ctx, audience, application_ids)
	if err != nil {
		return fmt.Errorf("get requests: %w", err)
	}

	if err := s.audienceRepo.UpdateApplication(ctx, audience.ID, requests); err != nil {
		return fmt.Errorf("update requests: %w", err)
	}

	// Prepare message for RabbitMQ
	msg := domain.AudienceMessage{
		AudienceID:   audience.ID,
		Applications: audience.Applications,
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
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	s.logger.Info("audience processed and message published",
		zap.String("audience_id", string(audience.ID)),
		zap.Int("request_count", len(audience.Applications)))

	return nil
}

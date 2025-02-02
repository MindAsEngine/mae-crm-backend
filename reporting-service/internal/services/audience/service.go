package audience

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	//"golang.org/x/text/message"

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
	logger *zap.Logger) *Service {
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
		os.Getenv("RABBITMQ_EXCHANGE"), // name
		"direct",                       // type
		true,                           // durable
		false,                          // auto-deleted
		false,                          // internal
		false,                          // no-wait
		nil,                            // arguments
	)
	if err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	_, err = s.amqpChan.QueueDeclare(
		os.Getenv("RABBITMQ_QUEUE"), // name
		true,                        // durable
		false,                       // delete when unused
		false,                       // exclusive
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	return s.amqpChan.QueueBind(
		os.Getenv("RABBITMQ_QUEUE"),       // queue name
		os.Getenv("RABBITMQ_ROUTING_KEY"), // routing key
		os.Getenv("RABBITMQ_EXCHANGE"),    // exchange
		false,
		nil,
	)
}

func (s *Service) GetFilters(ctx context.Context) (domain.ApplicationFilterResponce, error) {
	filter := domain.ApplicationFilterResponce{}

	filter, err := s.mysqlRepo.GetFilters(ctx)

	if err != nil {
		return domain.ApplicationFilterResponce{}, fmt.Errorf("get filters: %w", err)
	}

	filter.AudienceNames, err = s.audienceRepo.ListAudiencenames(ctx)
	if err != nil {
		return domain.ApplicationFilterResponce{}, fmt.Errorf("get filters: %w", err)
	}
	return filter, nil
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

func (s *Service) AudienceList(ctx context.Context) ([]domain.AudienceResponse, error) {
	audiences, err := s.audienceRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("get audiences: %w", err)
	}

	var response []domain.AudienceResponse
	for _, a := range audiences {
		response = append(response, domain.AudienceResponse{
			ID:                 a.ID,
			Name:               a.Name,
			Integrations:       a.Integrations,
			Applications_count: len(a.Application_ids),
			CreatedAt:          a.CreatedAt,
			UpdatedAt:          a.UpdatedAt,
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

func (s *Service) ExportAudience(ctx context.Context, id int64) (string, string, error) {
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

func (s *Service) ListApplications(ctx context.Context, pagination *domain.PaginationRequest, filter *domain.ApplicationFilterRequest) (*domain.PaginationResponse, error) {
	// audienceApplicationIDs, err := s.audienceRepo.GetApplicationIdsByAudienceName(ctx, filter.AudienceName)

	// if audienceApplicationIDs !=nil && err == nil {
	// 	filter.AudienceIDs = audienceApplicationIDs
	// }
	//s.logger.Info("audienceApplicationIDs", zap.Any("audienceApplicationIDs", audienceApplicationIDs))

	if filter.AudienceName != "" {
		audienceId, err := s.audienceRepo.GetByName(ctx, filter.AudienceName)
		if err != nil {
			return nil, fmt.Errorf("get audience id: %w", err)
		}
		audience_filter, err := s.audienceRepo.GetFilterByAudienceId(ctx, audienceId.ID)

		if err != nil {
			return nil, fmt.Errorf("get filter by audience id: %w", err)
		}

		filter.AudienceIDs = append(filter.AudienceIDs, string(audienceId.ID))
		response, err := s.mysqlRepo.ListApplicationsWithFilters(ctx, pagination, filter, audience_filter)
		s.logger.Info("list applications", zap.Any("response", filter))
		if err != nil {
			return nil, fmt.Errorf("get applications: %w", err)
		}
		return response, nil

	} else {
		response, err := s.mysqlRepo.ListApplicationsWithFilters(ctx, pagination, filter, &domain.AudienceCreationFilter{})
		s.logger.Info("list applications", zap.Any("response", filter))
		if err != nil {
			return nil, fmt.Errorf("get applications: %w", err)
		}
		return response, nil
	}
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
		//s.logger.Info("filter", zap.Any("filter", filter))

		audience.Filter = domain.AudienceCreationFilter{
			StartDate:            filter.StartDate,
			EndDate:              filter.EndDate,
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
		//s.logger.Info("current applications", zap.Any("current_applications", current_applications))

		//Получаем заявки, которые изменили статус
		changed_applications, err := s.mysqlRepo.GetChangedApplicationIds(ctx, &audience.Filter, current_applications)
		if err != nil {
			s.logger.Error("get changed applications failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}
		//s.logger.Info("changed applications", zap.Any("changed_applications", changed_applications))

		//Вот это в "удаляемые"
		//Удаляем заявки с измененными статусами
		if err := s.audienceRepo.DeleteApplications(ctx, audience.ID, changed_applications); err != nil {
			s.logger.Error("delete applications with changed statuses failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}
		//s.logger.Info("applications deleted", zap.Any("applications", changed_applications))

		//Получаем заявки которые не изменили статус
		current_applications, err = s.audienceRepo.GetApplicationIdsByAdienceId(ctx, audience.ID)
		if err != nil {
			s.logger.Error("get changed applications failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}
		//s.logger.Info("current applications", zap.Any("current_applications", current_applications))

		//Получаем обновленные заявки которые ещё не в аудитории
		requests, err := s.mysqlRepo.GetNewApplicationsByAudience(ctx, &audience, current_applications)
		if err != nil {
			s.logger.Error("get requests: ", zap.Error(err))
			continue
		}

		//А эти в "новые"
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
		// req_ids, err := s.audienceRepo.GetApplicationIdsByAdienceId(ctx, audience.ID)

		// if err != nil {
		// 	s.logger.Error("get application ids by audience id: ", zap.Error(err))
		// 	continue
		// }

		// if len(req_ids) == 0 {
		// 	s.logger.Info("no requests found", zap.Any("audience_id", audience.ID))
		// 	continue
		// }

		// requests, err = s.mysqlRepo.ListApplicationsByIds(ctx, req_ids)
		// if err != nil {
		// 	s.logger.Error("get audience: ", zap.Error(err))
		// 	continue
		// }

		// audience.Application_ids = req_ids

		if integration_names, err := s.audienceRepo.GetIntegrationNamesByAudienceId(ctx, audience.ID); err != nil {
			s.logger.Error("get integration names by audience id: ", zap.Error(err))
			continue
		} else {
			audience.IntegrationNames = integration_names
		}

		new_ids := make([]int64, 0, len(requests))
		for _, application := range requests {
			new_ids = append(new_ids, application.ID)
		}

		if err := s.pushAudienceToRabbit(ctx, &audience, new_ids, changed_applications); err != nil {
			s.logger.Error("process audience failed",
				zap.String("audience_id", string(audience.ID)),
				zap.Error(err))
			continue
		}
		//}
	}
	return nil
}

func (s *Service) ExportApplications(ctx context.Context, filter domain.ApplicationFilterRequest) (string, string, error) {
	return s.exporter.ExportApplications(ctx, &filter)
}

func (s *Service) pushAudienceToRabbit(ctx context.Context, audience *domain.Audience, new_ids []int64, delete_ids []int64) error {
	new_ids_chunks := splitIntoChunks(new_ids, 500)
	delete_ids_chunks := splitIntoChunks(delete_ids, 500)

	messages := make([]domain.AudienceMessage, max(len(new_ids_chunks), len(delete_ids_chunks)))

	if len(new_ids) > 0 {
		for i, chunk := range new_ids_chunks {
			messages[i] = domain.AudienceMessage{
				New_application_ids: chunk,
			}
		}
	}

	if len(delete_ids) > 0 {
		for i, chunk := range delete_ids_chunks {
			messages[i] = domain.AudienceMessage{
				Delete_application_ids: chunk,
			}
		}
	}

	for i, message := range messages {
		message.AudienceName = audience.Name
		message.AudienceID = audience.ID
		message.Integrations = audience.Integrations
		message.TotalChunks = len(messages)
		message.CurrentChunk = i + 1

		s.logger.Info("publishing audience update message", zap.Any("message", message))

		body, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("marshal message: %w", err)
		}

		err = s.amqpChan.PublishWithContext(
			ctx,
			os.Getenv("RABBITMQ_EXCHANGE"),    // exchange
			os.Getenv("RABBITMQ_ROUTING_KEY"), // routing key
			false,                             // mandatory
			false,                             // immediate
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
			zap.Int("request_count", len(audience.Applications)))
	}
	return nil
}

func (s *Service) GetRegions(ctx context.Context, filter *domain.RegionFilter) (*domain.RegionsResponse, error) {
	s.logger.Info("getting regions data") //,
	//zap.String("search", filter.Search),
	//zap.Time("start_date", *filter.StartDate),
	//zap.Time("end_date", *filter.EndDate),
	//zap.String("sort", filter.Sort))

	response, err := s.mysqlRepo.GetRegionsData(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get regions data: %w", err)
	}

	// Validate response
	if len(response.Data) == 0 {
		s.logger.Warn("no regions data found")
		return response, nil
	}

	s.logger.Info("got regions data",
		zap.Int("total_projects", len(response.Data)),
		zap.Int("total_regions", len(response.Headers)-2),
		zap.Any("footer", response.Footer),
	)

	return response, nil
}

func (s *Service) ExportRegions(ctx context.Context, filter *domain.RegionFilter) (string, string, error) {
	s.logger.Info("exporting regions data")

	response, err := s.mysqlRepo.GetRegionsData(ctx, filter)
	if err != nil {
		return "", "", fmt.Errorf("get regions data: %w", err)
	}

	filePath, fileName, err := s.exporter.ExportRegionsData(response)
	if err != nil {
		return "", "", fmt.Errorf("export regions data: %w", err)
	}

	return filePath, fileName, nil
}

func (s *Service) GetCallCenterReport(ctx context.Context, filter *domain.CallCenterReportFilter) (*domain.CallCenterReport, error) {
	s.logger.Info("getting call center report")

	report, err := s.mysqlRepo.GetCallCenterReportData(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get call center report: %w", err)
	}

	// Process anomalies
	//report.Anomalies = s.detectAnomalies(report.Data)

	return report, nil
}

func (s *Service) ExportCallCenterReport(ctx context.Context, filter *domain.CallCenterReportFilter) (string, string, error) {
	s.logger.Info("exporting call center report")

	report, err := s.mysqlRepo.GetCallCenterReportData(ctx, filter)
	if err != nil {
		return "", "", fmt.Errorf("get call center report: %w", err)
	}

	filePath, fileName, err := s.exporter.ExportCallCenterReport(report)
	if err != nil {
		return "", "", fmt.Errorf("export to excel: %w", err)
	}

	return filePath, fileName, nil
}

func (s *Service) GetSpeedReport(ctx context.Context, filter *domain.StatusDurationFilter) (*domain.StatusDurationResponse, error) {
	s.logger.Info("getting speed report")

	report, err := s.mysqlRepo.GetStatusDurationReport(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get speed report: %w", err)
	}

	return report, nil
}

// func (s *Service) ExportStatusDurationReport(ctx context.Context, filter *domain.StatusDurationFilter) (string, string, error) {
// 	s.logger.Info("exporting speed report")

// 	report, err := s.mysqlRepo.GetStatusDurationReport(ctx, filter)
// 	if err != nil {
// 		return "", "", fmt.Errorf("get speed report: %w", err)
// 	}

// 	filePath, fileName, err := s.exporter.ExportStatusDurationReport(report)
// 	if err != nil {
// 		return "", "", fmt.Errorf("export to excel: %w", err)
// 	}

// 	return filePath, fileName, nil
// }

func (s *Service) detectAnomalies(data []domain.ManagerMetrics) []string {
	var anomalies []string

	// Calculate averages
	var avgTargetConv, avgVisitConv, avgVisitSuccess float64
	for _, m := range data {
		avgTargetConv += m.TargetConversion
		avgVisitConv += m.VisitConversion
		avgVisitSuccess += m.VisitSuccess
	}
	count := float64(len(data))
	avgTargetConv /= count
	avgVisitConv /= count
	avgVisitSuccess /= count

	// Check for anomalies
	for _, m := range data {
		if m.TargetConversion < avgTargetConv*0.5 {
			anomalies = append(anomalies, fmt.Sprintf("Низкая конверсия в целевые у %s: %.1f%%", m.ManagerName, m.TargetConversion))
		}
		if m.VisitConversion < avgVisitConv*0.5 {
			anomalies = append(anomalies, fmt.Sprintf("Низкая конверсия в визиты у %s: %.1f%%", m.ManagerName, m.VisitConversion))
		}
		if m.VisitSuccess < avgVisitSuccess*0.5 {
			anomalies = append(anomalies, fmt.Sprintf("Низкая успешность визитов у %s: %.1f%%", m.ManagerName, m.VisitSuccess))
		}
	}

	return anomalies
}

func splitIntoChunks(ids []int64, chunkSize int) [][]int64 {
	var chunks [][]int64
	for i := 0; i < len(ids); i += chunkSize {
		end := i + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		chunks = append(chunks, ids[i:end])
	}
	return chunks
}

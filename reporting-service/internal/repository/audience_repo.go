package repository

import (
	"context"
	//"database/sql"
	"encoding/csv"
	// "errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"log"
	"reporting-service/internal/domain"

	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
	"github.com/streadway/amqp"
)

type AudienceRepo struct {
	DB *sqlx.DB
}

type RedisCache struct {
	client *redis.Client
}

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

	queue, err := ch.QueueDeclare(
		"UploadMessages", // имя очереди
		true,             // сохранять сообщения
		false,            // автоудаление
		false,            // эксклюзивность
		false,            // ожидание
		nil,              // дополнительные аргументы
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	log.Printf("Queue declared: %s", queue.Name)

	return &RabbitMQ{Channel: ch}
}

func InitPostgres() *sqlx.DB {
	connStr := "host=db port=5432 user=postgres password=example dbname=audiences sslmode=disable"
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	return db
}

func InitRedis() *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // Укажите пароль, если он требуется
		DB:       0,  // Используйте базу по умолчанию
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return &RedisCache{client: client}
}

func (r *AudienceRepo) CreateAudience(ctx context.Context, audience *domain.Audience) error {
	query := `
            INSERT INTO audiences (
                id, name, creation_date_from, creation_date_to, 
                statuses, rejection_reasons, non_target_reasons, 
                request_ids, last_updated
            ) VALUES (
                :id, :name, :creation_date_from, :creation_date_to,
                :statuses, :rejection_reasons, :non_target_reasons,
                :request_ids, :last_updated
            )`

	audience.ID = uuid.New()
	audience.CreatedAt = time.Now()
	audience.UpdatedAt = time.Now()

	_, err := r.DB.ExecContext(ctx, query, audience)
	return err
}

func (r *AudienceRepo) FindRequestsForAudience(ctx context.Context, filter domain.AudienceFilter) ([]domain.Request, error) {
	query := `
		SELECT id, created_at, status, rejection_reason, 
		       non_target_reason, responsible_user_id, client_data
		FROM requests 
		WHERE 1=1`

	args := []interface{}{}
	var conditions []string

	if filter.CreationDateFrom != nil {
		conditions = append(conditions, "created_at >= $"+fmt.Sprint(len(args)+1))
		args = append(args, filter.CreationDateFrom)
	}

	if filter.CreationDateTo != nil {
		conditions = append(conditions, "created_at <= $"+fmt.Sprint(len(args)+1))
		args = append(args, filter.CreationDateTo)
	}

	if len(filter.Statuses) > 0 {
		conditions = append(conditions, "status = ANY($"+fmt.Sprint(len(args)+1)+")")
		args = append(args, pq.Array(filter.Statuses))
	}

	if len(filter.RejectionReasons) > 0 {
		conditions = append(conditions, "rejection_reason = ANY($"+fmt.Sprint(len(args)+1)+")")
		args = append(args, pq.Array(filter.RejectionReasons))
	}

	if len(filter.NonTargetReasons) > 0 {
		conditions = append(conditions, "non_target_reason = ANY($"+fmt.Sprint(len(args)+1)+")")
		args = append(args, pq.Array(filter.NonTargetReasons))
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	var requests []domain.Request
	err := r.DB.SelectContext(ctx, &requests, query, args...)
	return requests, err
}

func (r *AudienceRepo) UpdateAudienceRequests(ctx context.Context, audienceID uuid.UUID) error {
    // Находим существующую аудиторию
    var audience domain.Audience
    query := `SELECT id, creation_date_from, creation_date_to, 
                     statuses, rejection_reasons, non_target_reasons, 
                     request_ids 
              FROM audiences WHERE id = $1`
    
    err := r.DB.GetContext(ctx, &audience, query, audienceID)
    if err != nil {
        return fmt.Errorf("failed to find audience: %w", err)
    }

    // Формируем фильтр для поиска новых заявок
    filter := domain.AudienceFilter{
        CreationDateFrom:   audience.CreationDateFrom,
        CreationDateTo:     audience.CreationDateTo,
        Statuses:           audience.Statuses,
        RejectionReasons:   audience.RejectionReasons,
        NonTargetReasons:   audience.NonTargetReasons,
    }

    // Находим новые заявки
    newRequests, err := r.FindRequestsForAudience(ctx, filter)
    if err != nil {
        return fmt.Errorf("failed to find new requests: %w", err)
    }

    // Создаем множество существующих идентификаторов заявок
    existingRequestIDsMap := make(map[uuid.UUID]bool)
    for _, id := range audience.RequestIDs {
        existingRequestIDsMap[id] = true
    }

    // Добавляем только новые заявки
    var updatedRequestIDs []uuid.UUID = audience.RequestIDs
    for _, req := range newRequests {
        if !existingRequestIDsMap[req.ID] {
            updatedRequestIDs = append(updatedRequestIDs, req.ID)
            existingRequestIDsMap[req.ID] = true
        }
    }

    // Обновляем аудиторию с новыми заявками
    updateQuery := `UPDATE audiences 
                    SET request_ids = $1, 
                        last_updated = NOW() 
                    WHERE id = $2`
    
    _, err = r.DB.ExecContext(ctx, updateQuery, pq.Array(updatedRequestIDs), audienceID)
    if err != nil {
        return fmt.Errorf("failed to update audience requests: %w", err)
    }

    return nil
}

func (r *AudienceRepo) ExportAudienceToCSV(ctx context.Context, audienceID uuid.UUID) (string, error) {
    // Находим заявки для аудитории
    query := `SELECT r.* 
              FROM requests r
              JOIN audiences a ON r.id = ANY(a.request_ids)
              WHERE a.id = $1`
    
    var requests []domain.Request
    err := r.DB.SelectContext(ctx, &requests, query, audienceID)
    if err != nil {
        return "", fmt.Errorf("failed to fetch audience requests: %w", err)
    }

    // Создаем временный файл для CSV
    file, err := os.CreateTemp("", "audience_export_*.csv")
    if err != nil {
        return "", fmt.Errorf("failed to create temp file: %w", err)
    }
    defer file.Close()

    // Создаем CSVWriter
    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Записываем заголовки
    headers := []string{
        "ID", "CreatedAt", "Status", "RejectionReason", 
        "NonTargetReason", "ResponsibleUserID", "ClientData",
    }
    if err := writer.Write(headers); err != nil {
        return "", fmt.Errorf("failed to write headers: %w", err)
    }

    // Записываем данные о заявках
    for _, req := range requests {
        row := []string{
            req.ID.String(), 
            req.CreatedAt.Format(time.RFC3339),
            string(req.Status),
            string(req.RejectionReason),
            string(req.NonTargetReason),
            req.ResponsibleUserID.String(),
            string(req.ClientData),
        }
        if err := writer.Write(row); err != nil {
            return "", fmt.Errorf("failed to write request row: %w", err)
        }
    }

    return file.Name(), nil
}


func (r *AudienceRepo) GetAudiences() ([]domain.Audience, error) {
	rows, err := r.DB.Query("SELECT id, name, description, created_at, updated_at FROM audiences")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	audiences := []domain.Audience{}
	for rows.Next() {
		var audience domain.Audience
		if err := rows.Scan(&audience.ID, &audience.Name, &audience.CreatedAt, &audience.UpdatedAt); err != nil {
			return nil, err
		}
		audiences = append(audiences, audience)
	}
	return audiences, nil
}

func (r *AudienceRepo) GetIntegrations() ([]domain.Integration, error) {
	rows, err := r.DB.Query("SELECT id, name, description, created_at, updated_at FROM integrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	integrations := []domain.Integration{}
	for rows.Next() {
		var integration domain.Integration
		if err := rows.Scan(&integration.ID, &integration.AudienceID, &integration.Platform, &integration.CreatedAt); err != nil {
			return nil, err
		}
		integrations = append(integrations, integration)
	}
	return integrations, nil
}

func (r *AudienceRepo) CreateIntegration(integration *domain.Integration) error {
	query := "INSERT INTO integrations (audience_id, platform, status, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id"
	return r.DB.QueryRow(query, integration.AudienceID, integration.Platform, "pending").Scan(&integration.ID)
}

func (r *AudienceRepo) DeleteIntegration(id int) error {
	query := "DELETE FROM integrations WHERE id = $1"
	_, err := r.DB.Exec(query, id)
	return err
}

func (r *AudienceRepo) GetIntegrationByID(id int) (*domain.Integration, error) {
	query := "SELECT id, audience_id, platform, status, created_at FROM integrations WHERE id = $1"
	var integration domain.Integration
	err := r.DB.QueryRow(query, id).Scan(&integration.ID, &integration.AudienceID, &integration.Platform, &integration.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

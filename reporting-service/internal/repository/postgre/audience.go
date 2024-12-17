package postgre

import (
	"context"
	"fmt"
	"reporting-service/internal/domain"

	//"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type PostgresAudienceRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewPostgresAudienceRepository(db *sqlx.DB) *PostgresAudienceRepository {
	return &PostgresAudienceRepository{
		db:     db,
		logger: zap.L().With(zap.String("repository", "postgres_audience")),
	}
}

func (r *PostgresAudienceRepository) Create(ctx context.Context, audience *domain.Audience) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert audience
	query := `
        INSERT INTO audiences (name)
        VALUES ($1)
        RETURNING id`

	err = tx.QueryRowxContext(ctx, query,
		audience.Name,
	).Scan(&audience.ID)
	if err != nil {
		return fmt.Errorf("insert audience: %w", err)
	}

	// Insert requests
	for _, req := range audience.Applications {

		query := `
            INSERT INTO audience_requests (
                audience_id,
                request_id,
            ) VALUES ($1, $2)`

		_, err = tx.ExecContext(ctx, query,
			audience.ID,
			req.ID,
		)
		if err != nil {
			return fmt.Errorf("insert request: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresAudienceRepository) GetByID(ctx context.Context, id int64) (*domain.Audience, error) {
	audience := &domain.Audience{}

	query := `
        SELECT 
            a.id,
            a.name,
            a.created_at,
            a.updated_at
        FROM audiences a
        WHERE a.id = $1
        `

	err := r.db.GetContext(ctx, audience, query, id)
	if err != nil {
		return nil, fmt.Errorf("select audience: %w", err)
	}

	var integrations []domain.Integration
	integrationsQuery := `
        SELECT 
            i.id,
            i.cabinet_name,
            i.created_at,
            i.updated_at
        FROM integrations i
        WHERE i.audience_id = $1`

	if err := r.db.SelectContext(ctx, &integrations, integrationsQuery, audience.ID); err != nil {
		return nil, fmt.Errorf("select integrations: %w", err)
	}

	audience.Integrations = integrations
	return audience, nil
}


func (r *PostgresAudienceRepository) GetFilterByAudienceId(ctx context.Context, audience_id int64) (*domain.AudienceFilter, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	filter := &domain.AudienceFilter{}

	query := `
	SELECT *
	FROM audience_filters
	WHERE audience_id = $1`
	
	if err := r.db.SelectContext(ctx, &filter, query, audience_id); err != nil {
		return nil, fmt.Errorf("select filter: %w", err)
	}
	
	return filter, err
}


func (r *PostgresAudienceRepository) CreateIntegration(ctx context.Context, integration *domain.Integration, audience_id int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM integrations
		WHERE audience_id = $1 AND cabinet_name = $2`,
		audience_id, integration.CabinetName,).Scan(&integration.ID)

	if err == nil {
		return fmt.Errorf("integration already exists")
	}
	query := `
        INSERT INTO integrations (audience_id, cabinet_name)
        VALUES ($1, $2)
        RETURNING id`

	err = tx.QueryRowxContext(ctx, query,
		audience_id,
		integration.CabinetName,
	).Scan(&integration.ID)
	if err != nil {
		return fmt.Errorf("insert integration: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE audiences 
        SET updated_at = NOW() 
        WHERE id = $1`,
		audience_id)
	if err != nil {
		return fmt.Errorf("update audience timestamp: %w", err)
	}
	return tx.Commit()
}

func (r *PostgresAudienceRepository) List(ctx context.Context) ([]domain.Audience, error) {
	var audiences []domain.Audience

	// Get base audiences
	audiencesQuery := `
        SELECT 
            a.id,
            a.name,
            a.created_at,
            a.updated_at
        FROM audiences a
		`

	if err := r.db.SelectContext(ctx, &audiences, audiencesQuery); err != nil {
		return nil, fmt.Errorf("select audiences: %w", err)
	}

	// Get applications for each audience
	for i := range audiences {
		var applications []domain.Application
		applicationsQuery := `
			SELECT 
				ar.id
			FROM audience_requests ar
			WHERE ar.audience_id = $1`

		if err := r.db.SelectContext(ctx, &applications, applicationsQuery, audiences[i].ID); err != nil {
			return nil, fmt.Errorf("select applications: %w", err)
		}

		audiences[i].Applications = applications
	}

	// Get integrations for each audience
	for i := range audiences {
		var integrations []domain.Integration
		integrationsQuery := `
            SELECT 
                i.id,
                i.cabinet_name,
                i.created_at,
                i.updated_at
            FROM integrations i
            WHERE i.audience_id = $1`

		if err := r.db.SelectContext(ctx, &integrations, integrationsQuery, audiences[i].ID); err != nil {
			return nil, fmt.Errorf("select integrations: %w", err)
		}

		audiences[i].Integrations = integrations
	}

	return audiences, nil
}

func (r *PostgresAudienceRepository) UpdateApplication(ctx context.Context, audienceID int64, requests []domain.Application) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Batch insert new requests
	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO audience_requests (
            audience_id,
            request_id
        ) VALUES ($1, $2)`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert new requests
	for _, req := range requests {
		_, err = stmt.ExecContext(ctx,
			audienceID,
			req.ID,
		)
		if err != nil {
			return fmt.Errorf("insert request %d: %w", req.ID, err)
		}
	}

	// Update audience updated_at timestamp
	_, err = tx.ExecContext(ctx, `
        UPDATE audiences 
        SET updated_at = NOW() 
        WHERE id = $1`,
		audienceID)
	if err != nil {
		return fmt.Errorf("update audience timestamp: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresAudienceRepository) Delete(ctx context.Context, id int64) error {
	query := `
        DELETE FROM audiences 
        WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("execute delete: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("audience not found")
	}

	return nil
}

func (r *PostgresAudienceRepository) RemoveAllIntegrations(ctx context.Context, id int64) error {
	query := `
        DELETE FROM integrations 
        WHERE audience_id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("execute delete integrations: %w", err)
	}

	return nil
}

func (r *PostgresAudienceRepository) GetApplicationIdsByAdienceId(ctx context.Context, audienceID int64) ([]domain.Application, error) {
	var ids []int64

	query := `
		SELECT 
			request_id
		FROM audience_requests
		WHERE audience_id = $1`

	if err := r.db.SelectContext(ctx, &ids, query, audienceID); err != nil {
		return nil, fmt.Errorf("select ids: %w", err)
	}

	var applications []domain.Application
	for i := range ids {
		 applications = append(applications, domain.Application{ID: ids[i]})
	}
	return applications, nil
}
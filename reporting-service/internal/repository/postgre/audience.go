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
	db *sqlx.DB
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
        INSERT INTO audiences (id, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id`

	err = tx.QueryRowxContext(ctx, query,
		audience.ID,
		audience.Name,
		audience.CreatedAt,
		audience.UpdatedAt,
	).Scan(&audience.ID)

	if err != nil {
		return fmt.Errorf("insert audience: %w", err)
	}

	// Insert requests
	for _, req := range audience.Requests {
		query := `
            INSERT INTO audience_requests (
                audience_id, request_id, status, reason, created_at, updated_at
            ) VALUES ($1, $2, $3, $4, $5, $6)`

		_, err = tx.ExecContext(ctx, query,
			audience.ID,
			req.ID,
			req.Status,
			req.NonTargetReason,
			req.CreatedAt,
			req.UpdatedAt,
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
        SELECT id, name, created_at, updated_at
        FROM audiences
        WHERE id = $1`

	err := r.db.GetContext(ctx, audience, query, id)
	if err != nil {
		return nil, fmt.Errorf("select audience: %w", err)
	}

	// Get requests
	requestsQuery := `
        SELECT request_id, status, reason, created_at, updated_at
        FROM audience_requests
        WHERE audience_id = $1`

	err = r.db.SelectContext(ctx, &audience.Requests, requestsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("select requests: %w", err)
	}

	// Get integrations
	integrationsQuery := `
        SELECT cabinet_id, integration_id
        FROM audience_integrations
        WHERE audience_id = $1`

	err = r.db.SelectContext(ctx, &audience.Integrations, integrationsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("select integrations: %w", err)
	}

	return audience, nil
}

func (r *PostgresAudienceRepository) List(ctx context.Context) ([]domain.Audience, error) {
	var audiences []domain.Audience

	query := `
        SELECT id, name, created_at, updated_at
        FROM audiences
        ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &audiences, query)
	if err != nil {
		return nil, fmt.Errorf("select audiences: %w", err)
	}

	// For each audience, get requests and integrations
	for i := range audiences {
		requestsQuery := `
            SELECT request_id, status, reason, created_at, updated_at
            FROM audience_requests
            WHERE audience_id = $1`

		err = r.db.SelectContext(ctx, &audiences[i].Requests, requestsQuery, audiences[i].ID)
		if err != nil {
			return nil, fmt.Errorf("select requests: %w", err)
		}

		integrationsQuery := `
            SELECT cabinet_id, integration_id
            FROM audience_integrations
            WHERE audience_id = $1`

		err = r.db.SelectContext(ctx, &audiences[i].Integrations, integrationsQuery, audiences[i].ID)
		if err != nil {
			return nil, fmt.Errorf("select integrations: %w", err)
		}
	}

	return audiences, nil
}

func (r *PostgresAudienceRepository) UpdateRequests(ctx context.Context, audienceID int64, requests []domain.Request) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Delete existing requests for this audience
    _, err = tx.ExecContext(ctx, `
        DELETE FROM audience_requests 
        WHERE audience_id = $1`,
        audienceID)
    if err != nil {
        return fmt.Errorf("delete existing requests: %w", err)
    }

    // Batch insert new requests
    stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO audience_requests (
            audience_id,
            request_id,
            status,
            reason,
            created_at,
            updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6)`)
    if err != nil {
        return fmt.Errorf("prepare statement: %w", err)
    }
    defer stmt.Close()

    // Insert new requests
    for _, req := range requests {
        _, err = stmt.ExecContext(ctx,
            audienceID,
            req.ID,
			req.CreatedAt,
			req.UpdatedAt,
			req.Status,
			req.RejectionReason,
			req.NonTargetReason,
			req.ResponsibleUserID,
        )
        if err != nil {
            return fmt.Errorf("insert request %d: %w", req.ID, err)
        }
    }

    // Update audience last_updated timestamp
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
        DELETE FROM audience_integrations 
        WHERE audience_id = $1`

    _, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("execute delete integrations: %w", err)
    }

    return nil
}
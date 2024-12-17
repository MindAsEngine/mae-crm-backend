package mysql

import (
	"context"
	"fmt"
	"reporting-service/internal/domain"
	//"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type MySQLAudienceRepository struct {
    db *sqlx.DB
	logger *zap.Logger
}

type ValidationError struct {
	Field string
	Error string
}

func NewMySQLAudienceRepository(db *sqlx.DB) *MySQLAudienceRepository {
    return &MySQLAudienceRepository{
        db:     db,
        logger: zap.L().With(zap.String("repository", "mysql_audience")),
    }
}


func validateAudienceFilter(filter domain.AudienceFilter) []ValidationError {
	var errors []ValidationError

	// Date range validation
	if filter.CreationDateFrom != nil && filter.CreationDateTo != nil {
		if filter.CreationDateFrom.After(*filter.CreationDateTo) {
			errors = append(errors, ValidationError{
				Field: "date_range",
				Error: "start date must be before end date",
			})
		}
	}

	// At least one filter must be specified
	if (filter.CreationDateFrom == nil || filter.CreationDateTo == nil) &&
		len(filter.StatusIDs) == 0 &&
		len(filter.RejectionReasonIDs) == 0 &&
		len(filter.NonTargetReasonIDs) == 0 {
		errors = append(errors, ValidationError{
			Field: "filter",
			Error: "at least one filter must be specified",
		})
	}

	return errors
}

func (r *MySQLAudienceRepository) GetApplicationsByFilter(ctx context.Context, filter domain.AudienceFilter) ([]domain.Application, error) {
	// Validate filter
	errs := validateAudienceFilter(filter)
	errs_string := ""
	for _, err := range errs {
		errs_string += fmt.Sprintf("%s: %s\n", err.Field, err.Error)
	}

	if errs != nil {
		return nil, fmt.Errorf("invalid filters: %v", errs_string)
	}

	// Build query
	query := `
        SELECT 
            eb.id,
            eb.date_added,
            eb.updated_at,
            eb.status_name,
            eb.status_reason_id,
			eb.manager_id,
			eb.contacts_id,
            ebrs.type,
			ebrs.status_reason_id,
			ebrs.name,
        FROM estate_buys as eb LEFT JOIN estate_statuses_reasons as ebrs
        ON ebrs.status_reason_id=eb.status_reason_id
		WHERE 1=1
		`

	args := map[string]interface{}{}

	// Add date filters
	if filter.CreationDateFrom != nil && filter.CreationDateTo != nil{
		query += " AND eb.date_added >= :creation_date_from"
		args["creation_date_from"] = filter.CreationDateFrom
		query += " AND eb.date_added <= :creation_date_to"
		args["creation_date_to"] = filter.CreationDateTo
	}

	// Add status filter
	if len(filter.StatusIDs) > 0 {
		query += " AND eb.status_id IN (:status_ids)"
		args["statuse_ids"] = filter.StatusIDs
	}

	// Add reason filters
	if len(filter.RejectionReasonIDs) > 0 || len(filter.NonTargetReasonIDs) > 0 {
		reasons := append(filter.RejectionReasonIDs, filter.NonTargetReasonIDs...)
		query += " AND ebrs.status_reason_id IN (:reason_ids)"
		args["reason_ids"] = reasons
	}

	// Execute query
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to bind named params: %w", err)
	}

	query, params, err = sqlx.In(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to expand IN clause: %w", err)
	}

	query = r.db.Rebind(query)

	var results []domain.Application
	err = r.db.SelectContext(ctx, &results, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return results, nil
}
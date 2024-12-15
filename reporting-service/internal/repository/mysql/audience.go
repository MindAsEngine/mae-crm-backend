package mysql

import (
	"context"
	"fmt"
	"reporting-service/internal/domain"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type MySQLAudienceRepository struct {
    db *sqlx.DB
	logger *zap.Logger
}


type RequestDetails struct {
	ID          int64     `db:"id"`
	CreatedAt   time.Time `db:"date_added"`
	LastUpdated time.Time `db:"date_modified"`
	Status      string    `db:"status_name"`
	Reason      *string   `db:"reason"`
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

		// Check if date range is not more than 1 year
		if filter.CreationDateTo.Sub(*filter.CreationDateFrom) > time.Hour*24*365 {
			errors = append(errors, ValidationError{
				Field: "date_range",
				Error: "date range cannot exceed 1 year",
			})
		}
	}

	// // Status validation
	// if len(filter.Statuses) > 10 {
	// 	errors = append(errors, ValidationError{
	// 		Field: "statuses",
	// 		Error: "cannot request more than 10 statuses",
	// 	})
	// }

	// allowedStatuses := map[string]bool{
	// 	"new":        true,
	// 	"in_work":    true,
	// 	"rejected":   true,
	// 	"completed":  true,
	// 	"non_target": true,
	// }

	// for _, status := range filter.Statuses {
	// 	if !allowedStatuses[status] {
	// 		errors = append(errors, ValidationError{
	// 			Field: "statuses",
	// 			Error: fmt.Sprintf("invalid status value: %s", status),
	// 		})
	// 	}
	// }

	// // Rejection reasons validation
	// if len(filter.RejectionReasons) > 20 {
	// 	errors = append(errors, ValidationError{
	// 		Field: "rejection_reasons",
	// 		Error: "cannot request more than 20 rejection reasons",
	// 	})
	// }

	// for _, reason := range filter.RejectionReasons {
	// 	if len(reason) > 100 {
	// 		errors = append(errors, ValidationError{
	// 			Field: "rejection_reasons",
	// 			Error: fmt.Sprintf("rejection reason too long: %s", reason),
	// 		})
	// 	}
	// }

	// // Non-target reasons validation
	// if len(filter.NonTargetReasons) > 20 {
	// 	errors = append(errors, ValidationError{
	// 		Field: "non_target_reasons",
	// 		Error: "cannot request more than 20 non-target reasons",
	// 	})
	// }

	// for _, reason := range filter.NonTargetReasons {
	// 	if len(reason) > 100 {
	// 		errors = append(errors, ValidationError{
	// 			Field: "non_target_reasons",
	// 			Error: fmt.Sprintf("non-target reason too long: %s", reason),
	// 		})
	// 	}
	// }

	// At least one filter must be specified
	if filter.CreationDateFrom == nil &&
		filter.CreationDateTo == nil &&
		len(filter.Statuses) == 0 &&
		len(filter.RejectionReasons) == 0 &&
		len(filter.NonTargetReasons) == 0 {
		errors = append(errors, ValidationError{
			Field: "filter",
			Error: "at least one filter must be specified",
		})
	}

	return errors
}

func (r *MySQLAudienceRepository) GetAudienceByFilter(ctx context.Context, filter domain.AudienceFilter) ([]domain.Request, error) {
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
            eb.date_modified,
            eb.status_name,
            eba.attr_value as reason
        FROM estate_buys eb
        LEFT JOIN estate_buys_attributes eba ON eb.id = eba.entity_id 
            AND eba.entity = 'estate_buy'
        WHERE 1=1`

	args := map[string]interface{}{}

	// Add date filters
	if filter.CreationDateFrom != nil {
		query += " AND eb.date_added >= :creation_date_from"
		args["creation_date_from"] = filter.CreationDateFrom
	}
	if filter.CreationDateTo != nil {
		query += " AND eb.date_added <= :creation_date_to"
		args["creation_date_to"] = filter.CreationDateTo
	}

	// Add status filter
	if len(filter.Statuses) > 0 {
		query += " AND eb.status_name IN (:statuses)"
		args["statuses"] = filter.Statuses
	}

	// Add reason filters
	if len(filter.RejectionReasons) > 0 || len(filter.NonTargetReasons) > 0 {
		reasons := append(filter.RejectionReasons, filter.NonTargetReasons...)
		query += " AND eba.attr_value IN (:reasons)"
		args["reasons"] = reasons
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

	var results []domain.Request
	err = r.db.SelectContext(ctx, &results, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return results, nil
}
package mysql

import (
	"context"
	"fmt"
	"math"
	"reporting-service/internal/domain"
	"strings"

	//"reporting-service/internal/repository"

	//"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type MySQLAudienceRepository struct {
	db     *sqlx.DB
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
		if filter.CreationDateTo.Sub(*filter.CreationDateFrom).Hours() > 8784 {
			errors = append(errors, ValidationError{
				Field: "date_range",
				Error: "date range must be less than 366 days (8784 hours)",
			})

		}
	}

	// At least one filter must be specified
	if filter.CreationDateFrom == nil || filter.CreationDateTo == nil {
		errors = append(errors, ValidationError{
			Field: "filter",
			Error: "date range fields required",
		})
	}
	return errors
}

func (r *MySQLAudienceRepository) GetApplicationsByAudienceFilter(ctx context.Context, filter domain.AudienceFilter) ([]domain.Application, error) {
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
			eb.manager_id,
			eb.contacts_id,
			eb.status,
			COALESCE(ebrs.name, '') as name,
			COALESCE(ebrs.status_reason_id, 0) as status_reason_id
        FROM estate_buys eb LEFT JOIN estate_statuses_reasons ebrs
        ON ebrs.status_reason_id=eb.status_reason_id
		WHERE 1=1
		`

	args := map[string]interface{}{}

	// Add date filters
	if filter.CreationDateFrom != nil && filter.CreationDateTo != nil {
		query += " AND eb.date_added >= :creation_date_from"
		args["creation_date_from"] = filter.CreationDateFrom
		query += " AND eb.date_added <= :creation_date_to"
		args["creation_date_to"] = filter.CreationDateTo
	}

	// Add status filter
	if len(filter.StatusNames) > 0 {
		query += " AND eb.status_name IN (:status_names)"
		args["status_names"] = filter.StatusNames
	}

	// Add reason filters
	if len(filter.RegectionReasonNames) > 0 || len(filter.NonTargetReasonNames) > 0 {
		reasons := append(filter.RegectionReasonNames, filter.NonTargetReasonNames...)
		query += " AND ebrs.name IN (:reason_names)"
		args["reason_names"] = reasons
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

func (r *MySQLAudienceRepository) GetNewApplicationsByAudience(ctx context.Context, audience *domain.Audience, apllication_ids []int64) ([]domain.Application, error) {
	// Build query
	query := `
        SELECT 
            eb.id,
            eb.date_added,
            eb.updated_at,
            eb.status_name,
			eb.manager_id,
			eb.contacts_id,
			eb.status,
			COALESCE(ebrs.name, '') as name,
			COALESCE(ebrs.status_reason_id, 0) as status_reason_id
        FROM estate_buys eb LEFT JOIN estate_statuses_reasons ebrs
        ON ebrs.status_reason_id=eb.status_reason_id
		WHERE 1=1
		`

	args := map[string]interface{}{}

	query += " AND eb.id NOT IN (:apllication_ids)"
	args["apllication_ids"] = apllication_ids

	// Add date filters
	if audience.Filter.CreationDateFrom != nil && audience.Filter.CreationDateTo != nil {
		query += " AND eb.date_added >= :creation_date_from"
		args["creation_date_from"] = audience.Filter.CreationDateFrom
		query += " AND eb.date_added <= :creation_date_to"
		args["creation_date_to"] = audience.Filter.CreationDateTo
	}

	// Add status filter
	if len(audience.Filter.StatusNames) > 0 {
		query += " AND eb.status_name IN (:status_names)"
		args["status_names"] = audience.Filter.StatusNames
	}

	// Add reason filters
	if len(audience.Filter.RegectionReasonNames) > 0 || len(audience.Filter.NonTargetReasonNames) > 0 {
		reasons := append(audience.Filter.RegectionReasonNames, audience.Filter.NonTargetReasonNames...)
		query += " AND ebrs.name IN (:reason_names)"
		args["reason_names"] = reasons
	}

	// Get the latest creation date of all applications in the audience
	if len(audience.Applications) == 0 {
		return nil, fmt.Errorf("audience has no applications")
	}

	// Нахождение заявок которые появились после позднее последней заявки аудитории.(Не используется)

	// latestDate := audience.Applications[0].ID

	// for  i, app := range audience.Applications {
	//     if app.CreatedAt.After(latestDate) {
	//         latestDate = app.ID
	//     }
	// 	if 	!repository.SliceConatinsString(audience.Filter.StatusNames, app.StatusName){
	// 		// query :=`
	// 		// DELETE FROM audience_requests
	// 		// WHERE audience_id = :audience_id AND request_id = :request_id
	// 		// `
	// 	}
	// }

	// query += " AND eb.date_added >= :last_creation_date"
	// args["last_creation_date"] = latestDate

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

func (r *MySQLAudienceRepository) GetChangedApplicationIds(ctx context.Context, filter *domain.AudienceFilter, application_ids []int64) ([]int64, error) {
	// Build query
	args := map[string]interface{}{}

	query := `
		SELECT 
			eb.id
		FROM estate_buys eb LEFT JOIN estate_statuses_reasons ebrs
		ON ebrs.status_reason_id=eb.status_reason_id
		WHERE eb.id IN (:apllication_ids) AND NOT eb.status_name IN (:status_names)
		`
	args["apllication_ids"] = application_ids
	args["status_names"] = filter.StatusNames
	r.logger.Info("args", zap.Any("statuses:", filter.StatusNames))

	if application_ids == nil {
		return nil, fmt.Errorf("application_ids is nil")
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

	var results []int64
	err = r.db.SelectContext(ctx, &results, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return results, err
}

func (r *MySQLAudienceRepository) ListApplicationsByIds(ctx context.Context, application_ids []int64) ([]domain.Application, error) {

	r.logger.Info("application_id", zap.Any("application_ids", application_ids))
	query := `
		SELECT 
            eb.id,
            eb.date_added,
            eb.updated_at,
            eb.status_name,
			eb.manager_id,
			eb.contacts_id,
			eb.status,
			COALESCE(ebrs.name, '') as name,
			COALESCE(ebrs.status_reason_id, 0) as status_reason_id
        FROM estate_buys eb LEFT JOIN estate_statuses_reasons ebrs
        ON ebrs.status_reason_id=eb.status_reason_id
		WHERE id IN (:application_id)
		`
	args := map[string]interface{}{}
	args["application_id"] = application_ids

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

func (r *MySQLAudienceRepository) ListApplicationsWithFilters(ctx context.Context, pagination *domain.PaginationRequest, filter *domain.ApplicationFilter) (*domain.PaginationResponse, error) {
    // Build base query for counting
    countQuery := `
        SELECT COUNT(*) 
        FROM estate_buys eb
        LEFT JOIN estate_houses h ON h.id = eb.house_id
        WHERE eb.company_id = 528
    `

    // Build where conditions and args map
    whereConditions := []string{}
    args := map[string]interface{}{}

    if filter.Status != "" {
        whereConditions = append(whereConditions, "eb.status_name = :status")
        args["status"] = filter.Status
    }

    if filter.PropertyType != "" {
        whereConditions = append(whereConditions, "eb.category = :property_type")
        args["property_type"] = filter.PropertyType
    }

    if filter.ProjectName != "" {
        whereConditions = append(whereConditions, "h.complex_name = :project_name")
        args["project_name"] = "%" + filter.ProjectName + "%"
    }

    if filter.DaysInStatus > 0 {
        whereConditions = append(whereConditions, `
            DATEDIFF(NOW(), COALESCE(
                (SELECT MAX(log_date) 
                FROM estate_buys_statuses_log 
                WHERE estate_buy_id = eb.id 
                AND status_to = eb.status),
                eb.date_added
            )) >= :days_in_status
        `)
        args["days_in_status"] = filter.DaysInStatus
    }

    if len(whereConditions) > 0 {
        countQuery += " AND " + strings.Join(whereConditions, " AND ")
    }

    // Get total count
    var totalItems int64
    countQuery, countArgs, err := sqlx.Named(countQuery, args)
    if err != nil {
        return nil, fmt.Errorf("prepare count query: %w", err)
    }

    countQuery = r.db.Rebind(countQuery)
    if err := r.db.GetContext(ctx, &totalItems, countQuery, countArgs...); err != nil {
        return nil, fmt.Errorf("count applications: %w", err)
    }

    // Prepare main query
    query := `
        SELECT 
            eb.id AS id,
            eb.date_added AS date_added,
            COALESCE(edc.contacts_buy_name, 'Не указано') AS client_name,
            eb.status_name AS status_name,
            COALESCE(edc.contacts_buy_phones, 'Не указано') AS phone,
            COALESCE(u.users_name, 'Не назначен') AS manager_name,
            eb.category AS property_type,
            DATEDIFF(NOW(), COALESCE(
                (SELECT MAX(log_date) 
                FROM estate_buys_statuses_log 
                WHERE estate_buy_id = eb.id 
                AND status_to = eb.status),
                eb.date_added
            )) AS days_in_status,
            COALESCE(h.complex_name, 'Не указан') AS project_name
        FROM estate_buys eb
        LEFT JOIN estate_deals_contacts edc ON edc.id = eb.contacts_id
        LEFT JOIN users u ON u.id = eb.manager_id
        LEFT JOIN estate_houses h ON h.id = eb.house_id
        WHERE eb.company_id = 528
    `

    if len(whereConditions) > 0 {
        query += " AND " + strings.Join(whereConditions, " AND ")
    }

    // Add pagination
    if pagination.PageSize <= 0 {
        pagination.PageSize = 10
    }
    if pagination.Page <= 0 {
        pagination.Page = 1
    }
    
    offset := (pagination.Page - 1) * pagination.PageSize
    
    query += " ORDER BY eb.date_added DESC LIMIT :limit OFFSET :offset"
    args["limit"] = pagination.PageSize
    args["offset"] = offset

    // Execute main query
    query, queryArgs, err := sqlx.Named(query, args)
    if err != nil {
        return nil, fmt.Errorf("prepare main query: %w", err)
    }

    query = r.db.Rebind(query)
    var items []domain.Application
    if err := r.db.SelectContext(ctx, &items, query, queryArgs...); err != nil {
        return nil, fmt.Errorf("select applications: %w", err)
    }

    totalPages := int(math.Ceil(float64(totalItems) / float64(pagination.PageSize)))

    return &domain.PaginationResponse{
        Items:      items,
        TotalItems: totalItems,
        TotalPages: totalPages,
        Page:       pagination.Page,
        PageSize:   pagination.PageSize,
    }, nil
}
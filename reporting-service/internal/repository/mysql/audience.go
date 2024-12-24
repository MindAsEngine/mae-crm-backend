package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strconv"

	//"time"

	"reporting-service/internal/domain"
	"strings"

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

func validateAudienceFilter(filter domain.AudienceCreationFilter) []ValidationError {
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

func (r *MySQLAudienceRepository) GetFilters(ctx context.Context) (domain.ApplicationFilterResponce, error) {
	var filter domain.ApplicationFilterResponce

	query := `SELECT Distinct status_name FROM macro_bi_cmp_528.estate_buys`
	if err := r.db.SelectContext(ctx, &filter.Statuses, query); err != nil {
		return domain.ApplicationFilterResponce{}, fmt.Errorf("select filters: %w", err)
	}
	query = `SELECT Distinct complex_name FROM macro_bi_cmp_528.estate_houses`
	if err := r.db.SelectContext(ctx, &filter.ProjectNames, query); err != nil {
		return domain.ApplicationFilterResponce{}, fmt.Errorf("select filters: %w", err)
	}
	query = `SELECT Distinct category FROM macro_bi_cmp_528.estate_buys WHERE category != ''`
	if err := r.db.SelectContext(ctx, &filter.PropertyTypes, query); err != nil {
		return domain.ApplicationFilterResponce{}, fmt.Errorf("select filters: %w", err)
	}
	return filter, nil
}

func (r *MySQLAudienceRepository) GetApplicationsByAudienceFilter(ctx context.Context, filter domain.AudienceCreationFilter) ([]domain.Application, error) {
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
			COALESCE(eb.manager_id, -1) as manager_id,
			eb.contacts_id,
			eb.status,
			COALESCE(ebrs.name, '') as name,
			COALESCE(ebrs.status_reason_id, -1) as status_reason_id
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
			COALESCE(eb.manager_id, -1) as manager_id,
			eb.contacts_id,
			eb.status,
			COALESCE(ebrs.name, '') as name,
			COALESCE(ebrs.status_reason_id, -1) as status_reason_id
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

func (r *MySQLAudienceRepository) GetChangedApplicationIds(ctx context.Context, filter *domain.AudienceCreationFilter, application_ids []int64) ([]int64, error) {
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
			COALESCE(eb.manager_id, -1) as manager_id,
			eb.contacts_id,
			eb.status,
			COALESCE(ebrs.name, '') as name,
			COALESCE(ebrs.status_reason_id, -1) as status_reason_id
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

func (r *MySQLAudienceRepository) ListApplicationsWithFilters(ctx context.Context, pagination *domain.PaginationRequest, filter *domain.ApplicationFilterRequest) (*domain.PaginationResponse, error) {
	// Build base query for counting
	countQuery := `
        SELECT COUNT(*) 
        FROM estate_buys eb
        LEFT JOIN estate_houses h ON h.id = eb.house_id
        WHERE eb.company_id = 528
    `
	// Build where conditions and args map
	args := map[string]interface{}{}
	// Get total count
	var totalItems int64

	// Prepare main query
	query := `
        SELECT 
            eb.id AS id,
            eb.date_added AS date_added,
            COALESCE(edc.contacts_buy_name, 'Не указано') AS client_name,
            eb.status_name AS status_name,
            COALESCE(edc.contacts_buy_phones, 'Не указано') AS phone,
            COALESCE(u.users_name, 'Не назначен') AS manager_name,
            COALESCE(eb.category, "Не указано") AS property_type,
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
	whereConditions := []string{}
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
		args["project_name"] = filter.ProjectName
	}

	if filter.StatusDuration > 0 {
		whereConditions = append(whereConditions, `
		DATEDIFF(NOW(), COALESCE(
			(SELECT MAX(log_date) 
			FROM estate_buys_statuses_log 
			WHERE estate_buy_id = eb.id 
			AND status_to = eb.status),
			eb.date_added
		)) >= :days_in_status
	`)
		args["days_in_status"] = filter.StatusDuration
	}

	if filter.CreatedAtFrom != nil &&
		filter.CreatedAtTo != nil &&
		!filter.CreatedAtFrom.IsZero() &&
		!filter.CreatedAtTo.IsZero() {
		whereConditions = append(whereConditions, "eb.date_added >= :created_at_from")
		args["created_at_from"] = filter.CreatedAtFrom
		whereConditions = append(whereConditions, "eb.date_added <= :created_at_to")
		args["created_at_to"] = filter.CreatedAtTo
	}

	if len(whereConditions) > 0 {
		countQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

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

	// Define allowed sort fields mapping
	sortableFields := map[string]string{
		"id":           "eb.id",
		"created_date": "eb.date_added",
		"client_name":  "edc.contacts_buy_name",
		"status":       "eb.status_name",
		"phone":        "edc.contacts_buy_phones",
		"manager":      "u.users_name",
		"property":     "eb.category",
	}

	// Build ORDER BY clause
	orderClause := " ORDER BY eb.date_added DESC" // Default sorting
		
	
	if filter.OrderField != "" {
		dbField, exists := sortableFields[filter.OrderField]
		if !exists {
			r.logger.Warn("invalid sort field requested, using default",
				zap.String("field", filter.OrderField))
		} else {
			direction := "ASC"
			if strings.ToUpper(filter.OrderDirection) == "DESC" {
				direction = "DESC"
			}
			orderClause = fmt.Sprintf(" ORDER BY %s %s", dbField, direction)
		}
	}

	countQuery, countArgs, err := sqlx.Named(countQuery, args)
	if err != nil {
		return nil, fmt.Errorf("prepare count query: %w", err)
	}

	countQuery = r.db.Rebind(countQuery)
	if err := r.db.GetContext(ctx, &totalItems, countQuery, countArgs...); err != nil {
		return nil, fmt.Errorf("count applications: %w", err)
	}

	// Apply pagination after sorting
	fullQuery := query + orderClause + " LIMIT :limit OFFSET :offset"
	args["limit"] = pagination.PageSize
	args["offset"] = offset

	// Debug log
	r.logger.Debug("executing query",
		zap.String("query", fullQuery),
		zap.Any("args", args))

	// Execute query with sorting and pagination
	query, queryArgs, err := sqlx.Named(fullQuery, args)
	if err != nil {
		return nil, fmt.Errorf("prepare query: %w", err)
	}

	query = r.db.Rebind(query)
	var items []domain.Application
	if err := r.db.SelectContext(ctx, &items, query, queryArgs...); err != nil {
		return nil, fmt.Errorf("select applications: %w", err)
	}

	headers := []domain.Header{
		{
			Name:         "id",
			IsID:         true,
			Title:        "ID",
			IsVisible:    true,
			IsAdditional: false,
			Format:       "number",
		},
		{
			Name:         "created_at",
			IsID:         false,
			Title:        "Дата",
			IsVisible:    true,
			IsAdditional: false,
			Format:       "date",
		},
		{
			Name:         "name",
			IsID:         false,
			Title:        "ФИО",
			IsVisible:    true,
			IsAdditional: false,
			Format:       "string",
		},
		{
			Name:         "status_name",
			IsID:         false,
			Title:        "Этап",
			IsVisible:    true,
			IsAdditional: false,
			Format:       "enum",
		},
		{
			Name:         "phone",
			IsID:         false,
			Title:        "Номер телефона",
			IsVisible:    true,
			IsAdditional: false,
			Format:       "string",
		},
		{
			Name:         "manager_name",
			IsID:         false,
			Title:        "Посредник",
			IsVisible:    true,
			IsAdditional: false,
			Format:       "string",
		},
		{
			Name:         "property_type",
			IsID:         false,
			Title:        "Тип недвижимости",
			IsVisible:    true,
			IsAdditional: false,
			Format:       "enum",
		},
	}

	//TODO Later: implement method for auto headers generation
	// t := reflect.TypeOf(items[0])

	// for i := 0; i < t.NumField(); i++ {
	// 	headers = append(headers, domain.Header{
	// 		Name:         t.Field(i).Name,
	// 		IsID:         t.Field(i).Name == "ID",
	// 		Title:        t.Field(i).Name,
	// 		IsVisible:    true,
	// 		IsAdditional: false,
	// 		Format:       t.Field(i).Type.String(),
	// 	})
	// }

	totalPages := int(math.Ceil(float64(totalItems) / float64(pagination.PageSize)))

	return &domain.PaginationResponse{
		Headers:    headers,
		Items:      items,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
	}, nil
}

func (r *MySQLAudienceRepository) ExportApplicationsWithFilters(ctx context.Context, filter *domain.ApplicationFilterRequest) ([]domain.Application, error) {
	baseQuery := `
        SELECT 
            eb.id AS id,
            eb.date_added AS date_added,
            COALESCE(edc.contacts_buy_name, 'Не указано') AS client_name,
            COALESCE(eb.status_name, "Не указано") AS status_name,
            COALESCE(edc.contacts_buy_phones, 'Не указано') AS phone,
            COALESCE(u.users_name, 'Не назначен') AS manager_name,
            COALESCE(eb.category, "Не указано") AS property_type,
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
        WHERE eb.company_id = 528`

	whereConditions := []string{}
	var args []interface{}

	if filter.Status != "" {
		whereConditions = append(whereConditions, "eb.status_name = ?")
		args = append(args, filter.Status)
	}

	if filter.PropertyType != "" {
		whereConditions = append(whereConditions, "eb.category = ?")
		args = append(args, filter.PropertyType)
	}

	if filter.ProjectName != "" {
		whereConditions = append(whereConditions, "h.complex_name LIKE ?")
		args = append(args, "%"+filter.ProjectName+"%")
	}

	if filter.StatusDuration > 0 {
		whereConditions = append(whereConditions, `
            DATEDIFF(NOW(), COALESCE(
                (SELECT MAX(log_date) 
                FROM estate_buys_statuses_log 
                WHERE estate_buy_id = eb.id 
                AND status_to = eb.status),
                eb.date_added
            )) >= ?
        `)
		args = append(args, filter.StatusDuration)
	}

	if filter.CreatedAtFrom != nil && filter.CreatedAtTo != nil {
		whereConditions = append(whereConditions, "eb.date_added >= ?")
		args = append(args, filter.CreatedAtFrom)
		whereConditions = append(whereConditions, "eb.date_added <= ?")
		args = append(args, filter.CreatedAtTo)

	}

	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	orderClause := " ORDER BY eb.date_added DESC"
	if filter.OrderField != "" {
		sortableFields := map[string]string{
			"id":           "eb.id",
			"created_date": "eb.date_added",
			"client_name":  "edc.contacts_buy_name",
			"status":       "eb.status_name",
			"phone":        "edc.contacts_buy_phones",
			"manager":      "u.users_name",
			"property":     "eb.category",
		}

		if dbField, exists := sortableFields[filter.OrderField]; exists {
			direction := "ASC"
			if strings.ToUpper(filter.OrderDirection) == "DESC" {
				direction = "DESC"
			}
			orderClause = fmt.Sprintf(" ORDER BY %s %s", dbField, direction)
		} else {
			r.logger.Warn("invalid sort field requested, using default",
				zap.String("field", filter.OrderField))
		}
	}

	fullQuery := baseQuery + orderClause

	r.logger.Debug("executing export query",
		zap.String("query", fullQuery),
		zap.Any("args", args))

	var applications []domain.Application
	if err := r.db.SelectContext(ctx, &applications, fullQuery, args...); err != nil {
		return nil, fmt.Errorf("select applications for export: %w", err)
	}

	r.logger.Info("applications exported successfully",
		zap.Int("count", len(applications)))

	return applications, nil
}

func (r *MySQLAudienceRepository) GetRegionsData(ctx context.Context, filter *domain.RegionFilter) (*domain.RegionsResponse, error) {
    // Step 1: Set session variables
    setSessionQuery := `
        SET SESSION group_concat_max_len = 10000;
    `
    if _, err := r.db.ExecContext(ctx, setSessionQuery); err != nil {
        return nil, fmt.Errorf("set session variables: %w", err)
    }

    // Step 2: Initialize variables
    if _, err := r.db.ExecContext(ctx, `SET @cities = NULL`); err != nil {
        return nil, fmt.Errorf("initialize @cities variable: %w", err)
    }
    if _, err := r.db.ExecContext(ctx, `SET @rn = 0`); err != nil {
        return nil, fmt.Errorf("initialize @rn variable: %w", err)
    }

    // Step 3: Generate cities SQL
    citiesQuery := `
        SELECT 
    GROUP_CONCAT(DISTINCT CONCAT(
        "SUM(CASE WHEN city_name = '", city_name, "' THEN total_requests ELSE 0 END) AS '", city_name, "'"
    )) INTO @cities
FROM (
    SELECT DISTINCT 
        REGEXP_SUBSTR(ec.passport_address, 'г\\.\\s*([^,\\s]+)') AS city_name
    FROM macro_bi_cmp_528.estate_deals_contacts ec
    LEFT JOIN macro_bi_cmp_528.estate_buys eb ON eb.contacts_id = ec.id
    WHERE (ec.passport_address IS NOT NULL) AND ec.passport_address != ''AND (eb.status_name = 'Сделка проведена' OR eb.status_name = 'Сделка в работе') AND eb.date_added >= '2024-01-01'
) city_list;
    `
    if _, err := r.db.ExecContext(ctx, citiesQuery); err != nil {
        return nil, fmt.Errorf("generate cities SQL: %w", err)
    }

    // var citiesSQL string
    // getCitiesSQLQuery := `SELECT @cities`
    // if err := r.db.QueryRowContext(ctx, getCitiesSQLQuery).Scan(&citiesSQL); err != nil {
    //     return nil, fmt.Errorf("get cities SQL: %w", err)
    // }
    // if _, err := r.db.ExecContext(ctx, `SET @sql = NULL;`); err != nil {
    //     return nil, fmt.Errorf("initialize @rn variable: %w", err)
    // }
    // Step 4: Build main query
    mainQuery := fmt.Sprintf(`
SET @sql = CONCAT(
    "SELECT 
        project_name, ",
        @cities, "
     FROM (
         SELECT 
             h.complex_name AS project_name, 
             REGEXP_SUBSTR(ec.passport_address, 'г\\.\\s*([^,\\s]+)') AS city_name,
             COUNT(eb.id) AS total_requests
         FROM macro_bi_cmp_528.estate_buys eb
         LEFT JOIN macro_bi_cmp_528.estate_sells es
			 ON es.estate_buy_id = eb.id
         LEFT JOIN macro_bi_cmp_528.estate_houses h 
             ON h.id = eb.house_id
         LEFT JOIN macro_bi_cmp_528.estate_deals_contacts ec 
             ON ec.id = eb.contacts_id
         WHERE 
             eb.house_id != 0 
             AND ec.passport_address != ''
             AND es.estate_sell_status_name = 'Сделка проведена' OR es.estate_sell_status_name = 'Сделка в работе'
         GROUP BY h.complex_name, REGEXP_SUBSTR(ec.passport_address, 'г\\.\\s*([^,\\s]+)')
     ) AS data
     GROUP BY project_name
     ORDER BY project_name"
);
    `)

    if _, err := r.db.ExecContext(ctx, mainQuery); err != nil {
        return nil, fmt.Errorf("build main query: %w", err)
    }

    // Step 5: Execute the prepared statement
    if _, err := r.db.ExecContext(ctx, "PREPARE stmt FROM @sql;"); err != nil {
        return nil, fmt.Errorf("prepare statement: %w", err)
    }

    // Step 6: Query the results
    rows, err := r.db.QueryContext(ctx, "EXECUTE stmt")
    if err != nil {
        return nil, fmt.Errorf("execute query: %w", err)
    }
    defer rows.Close()

    // if _, err := r.db.ExecContext(ctx, "DEALLOCATE PREPARE stmt;"); err != nil {
    //     return nil, fmt.Errorf("deallocate statement: %w", err)
    // }

    // Get column names for mapping
    columns, err := rows.Columns()
    if err != nil {
        return nil, fmt.Errorf("get columns: %w", err)
    }

    // Prepare headers
    headers := []domain.Header{
        {Name: "id", IsID: true, Title: "№", IsVisible: false, IsAdditional: false, Format: "number"},
        {Name: "name_of_projects", Title: "Наименование проектов", IsVisible: true, IsAdditional: false, Format: "string"},
    }

    for i := 2; i < len(columns); i++ {
        headers = append(headers, domain.Header{
            Name:         columns[i],
            Title:        columns[i],
            IsVisible:    true,
            IsAdditional: true,
            Format:       "number",
        })
    }

    // Process rows
    var data []map[string]interface{}
    footer := map[string]int{
        "id":             0,
        "name_of_projects": -1,
    }

	for rows.Next() {
        values := make([]interface{}, len(columns))
        valuePointers := make([]interface{}, len(columns))
        for i := range values {
            valuePointers[i] = new(sql.RawBytes)
        }

        if err := rows.Scan(valuePointers...); err != nil {
            return nil, fmt.Errorf("scan row: %w", err)
        }

        rowData := make(map[string]interface{})
        for i, col := range columns {
            rawValue := *(valuePointers[i].(*sql.RawBytes))
            if len(rawValue) == 0 {
                rowData[col] = 0
                if i > 1 { // Sum for footer (skip id and name)
                    footer[col] += 0
                }
                continue
            }

            valueStr := string(rawValue)
            fmt.Printf("Processing column: %s, value: %s\n", col, valueStr) // Debugging statement

            if num, err := strconv.Atoi(valueStr); err == nil {
                rowData[col] = num
                if i > 1 { // Sum for footer (skip id and name)
                    footer[col] += num
                }
            } else {
                rowData[col] = valueStr
            }
        }
        data = append(data, rowData)
    }

	footerData := make(map[string]interface{})
    for k, v := range footer {
        footerData[k] = v
    }
	footerData["name_of_projects"] = "Общее"
    return &domain.RegionsResponse{
        Headers: headers,
        Data:    data,
        Footer:  footerData,
    }, nil
}

func (r *MySQLAudienceRepository) GetCallCenterReportData(ctx context.Context, filter *domain.CallCenterReportFilter) (*domain.CallCenterReport, error) {
	query := `
        SELECT
            u.users_name,
            COUNT(*) as total_inquiries,
            COUNT(CASE 
                WHEN eb.status_name IN ('Подбор', 'Отказ') 
                THEN 1 
            END) as target_inquiries,
            COUNT(CASE 
                WHEN eb.status_name = 'Подбор' AND eb.custom_status_name = 'Визит состоялся'
                THEN 1 
            END) as completed_visits,
            COUNT(CASE 
                WHEN eb.status_name = 'Подбор' AND eb.custom_status_name = 'Назначенная встреча'
                THEN 1 
            END) as appointed_visits 
            -- COUNT(CASE 
            --     WHEN eb.status_name = 'Бронь'
            --     THEN 1 
            -- END) as brons,
            -- COUNT(CASE 
            --     WHEN eb.status_name IN ('Сделка проведена', 'Сделка в работе')
            --     THEN 1 
            -- END) as ddus
        FROM estate_buys eb 
        RIGHT JOIN users u ON eb.manager_id = u.id
        LEFT JOIN estate_buys_statuses_log ebsl ON eb.id = ebsl.estate_buy_id
        WHERE eb.company_id = 528 AND
        u.departments_id = 1903
    `

	orderClause := " GROUP BY u.users_name ORDER BY u.users_name DESC" // Default sorting

	var args []interface{}

	whereConditions := []string{}

	if filter.StartDate != nil &&
		filter.EndDate != nil &&
		!filter.StartDate.IsZero() &&
		!filter.EndDate.IsZero() {
			whereConditions = append(whereConditions, "eb.date_added >= ?")
			args = append(args, filter.StartDate)
			whereConditions = append(whereConditions, "eb.date_added <= ?")
			args = append(args, filter.EndDate)
	}

	if len(whereConditions) > 0 {
		query += " AND " + strings.Join(whereConditions, " AND ")
	}
	
	query += orderClause
	var metrics []domain.ManagerMetrics
	
	if err := r.db.SelectContext(ctx, &metrics, query, args...); err != nil {
		return nil, fmt.Errorf("get sales metrics: %w", err)
	}

	// Calculate conversions and totals
	totals := domain.ManagerMetrics{ManagerName: "Итого"}
	for i := range metrics {
		m := &metrics[i]

		// Calculate basic conversions
		if m.TotalInquiries > 0 {
			m.TargetConversion = float64(m.TargetInquiries) / float64(m.TotalInquiries) * 100
		}
		if m.TargetInquiries > 0 {
			m.VisitConversion = float64(m.AppointedVisits) / float64(m.TargetInquiries) * 100
			m.LeadToVisit = float64(m.CompletedVisits) / float64(m.TargetInquiries) * 100
		}
		if m.AppointedVisits > 0 {
			m.VisitSuccess = float64(m.CompletedVisits) / float64(m.AppointedVisits) * 100
		}

		//if m.CompletedVisits > 0 {
		//    m.VisitToBooking = float64(m.Bookings) / float64(m.CompletedVisits) * 100
		//}
		//if m.Bookings > 0 {
		//    m.BookingToContract = float64(m.Contracts) / float64(m.Bookings) * 100
		//}
		//if m.TargetInquiries > 0 {
		//    m.LeadToContract = float64(m.Contracts) / float64(m.TargetInquiries) * 100
		//}

		// Update totals
		totals.TotalInquiries += m.TotalInquiries
		totals.TargetInquiries += m.TargetInquiries
		totals.AppointedVisits += m.AppointedVisits
		totals.CompletedVisits += m.CompletedVisits
		//totals.Bookings += m.Bookings
		//totals.Contracts += m.Contracts
	}

	// Calculate total conversions
	if totals.TotalInquiries > 0 {
		totals.TargetConversion = float64(totals.TargetInquiries) / float64(totals.TotalInquiries) * 100
	}
	if totals.TargetInquiries > 0 {
		totals.VisitConversion = float64(totals.AppointedVisits) / float64(totals.TargetInquiries) * 100
		totals.LeadToVisit = float64(totals.CompletedVisits) / float64(totals.TargetInquiries) * 100
	}
	if totals.AppointedVisits > 0 {
		totals.VisitSuccess = float64(totals.CompletedVisits) / float64(totals.AppointedVisits) * 100
	}

	// Create headers
	headers := []domain.Header{
		{Name: "manager_name", Title: "ФИО менеджера", IsVisible: true, IsAdditional: true, Format: "string"},
		{Name: "total_inquiries", Title: "Всего обращений", IsVisible: true, IsAdditional: true, Format: "number"},
		{Name: "target_inquiries", Title: "Целевые", IsVisible: true, IsAdditional: true, Format: "number"},
		{Name: "target_conversion", Title: "Конверсия в целевые", IsVisible: true, IsAdditional: true, Format: "percent"},
		{Name: "appointed_visits", Title: "Назначено визитов", IsVisible: true, IsAdditional: true, Format: "number"},
		{Name: "visit_conversion", Title: "Конверсия в визиты", IsVisible: true, IsAdditional: true, Format: "percent"},
		{Name: "completed_visits", Title: "Визиты состоялись", IsVisible: true, IsAdditional: true, Format: "number"},
		{Name: "visit_success", Title: "Конверсия визитов", IsVisible: true, IsAdditional: true, Format: "percent"},
		{Name: "lead_to_visit", Title: "Конверсия лид->визит", IsVisible: true, IsAdditional: true, Format: "percent"},

		//{Name: "bookings", Title: "Бронирования", IsVisible: true, IsAdditional: true, Format: "number"},
		//{Name: "visit_to_booking", Title: "Конверсия визит->бронь", IsVisible: true, IsAdditional: true, Format: "percent"},
		//{Name: "contracts", Title: "ДДУ", IsVisible: true, IsAdditional: true, Format: "number"},
		//{Name: "booking_to_contract", Title: "Конверсия бронь->ДДУ", IsVisible: true, IsAdditional: true, Format: "percent"},
		//{Name: "lead_to_contract", Title: "Конверсия лид->ДДУ", IsVisible: true, IsAdditional: true, Format: "percent"},
	}

	// if showOptional {
	//     optionalHeaders := []domain.Header{
	//         {Name: "bookings", Title: "Бронирования", IsVisible: true, Format: "number"},
	//         {Name: "visit_to_booking", Title: "Конверсия визит->бронь", IsVisible: true, Format: "percent"},
	//         {Name: "contracts", Title: "ДДУ", IsVisible: true, Format: "number"},
	//         {Name: "booking_to_contract", Title: "Конверсия бронь->ДДУ", IsVisible: true, Format: "percent"},
	//         {Name: "lead_to_contract", Title: "Конверсия лид->ДДУ", IsVisible: true, Format: "percent"},
	//     }
	//     headers = append(headers, optionalHeaders...)
	// }

	return &domain.CallCenterReport{
		Headers: headers,
		Data:    metrics,
		Totals:  totals,
	}, nil
}

func sanitizeColumnName(name string) string {
	// Replace non-alphanumeric chars with underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	safe := reg.ReplaceAllString(name, "_")
	return fmt.Sprintf("region_%s", safe)
}

package mysql

import (
	"context"
	//"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"math"
	"reporting-service/internal/domain"
	//"strconv"
	"strings"
	"time"
)

type MySQLAudienceRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

type ValidationError struct {
	Field string
	Error string
}

func NewMySQLAudienceRepository(db *sqlx.DB, logger *zap.Logger) *MySQLAudienceRepository {
	return &MySQLAudienceRepository{
		db:     db,
		logger: logger,
	}
}

func validateAudienceFilter(filter domain.AudienceCreationFilter) []ValidationError {
	var errors []ValidationError

	// Date range validation
	if filter.StartDate != nil && filter.EndDate != nil {
		if filter.StartDate.After(*filter.EndDate) {
			errors = append(errors, ValidationError{
				Field: "date_range",
				Error: "start date must be before end date",
			})
		}
		if filter.EndDate.Sub(*filter.StartDate).Hours() > 8784 {
			errors = append(errors, ValidationError{
				Field: "date_range",
				Error: "date range must be less than 366 days (8784 hours)",
			})

		}
	}

	if filter.StartDate == nil {
		errors = append(errors, ValidationError{
			Field: "filter",
			Error: "start date field required",
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
	query = `
	SELECT DISTINCT 
        coalesce(TRIM(LOWER(REGEXP_REPLACE(REGEXP_SUBSTR(
                    passport_address,
                    '((г\\.|город )\\s*([^,\\s\\.]+))|(([^,\\s\\.]+)\\s(shah|shax|Ш(и|а)\\SРИ|ша\\Sар|город,|ш\\.))'
                ),
                '(г\\.|город\\s|\\sshah|\\sshax|\\sШ(и|а)\\SРИ|\\sша\\Sар|\\sгород,|\\sш\\.)', ''
            ))), "Не указано")
    	    AS city_name
    	FROM macro_bi_cmp_528.estate_deals_contacts ec
    	WHERE ec.passport_address IS NOT NULL AND
		TRIM(REGEXP_SUBSTR(ec.passport_address COLLATE utf8_general_ci, 'г\\.\\s*([^,\\s]+)')) != ''`
	if err := r.db.SelectContext(ctx, &filter.RegionNames, query); err != nil {
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

	if filter.EndDate != nil && !filter.EndDate.IsZero() {
		query += " AND eb.date_added <= :creation_date_to"
		args["creation_date_to"] = filter.EndDate
	}

	if filter.StartDate != nil && !filter.StartDate.IsZero() {
		query += " AND eb.date_added >= :creation_date_from"
		args["creation_date_from"] = filter.StartDate
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
	if audience.Filter.StartDate != nil && !audience.Filter.StartDate.IsZero() {
		query += " AND eb.date_added >= :creation_date_from"
		args["creation_date_from"] = audience.Filter.StartDate

	}

	if audience.Filter.EndDate != nil && !audience.Filter.EndDate.IsZero() {
		query += " AND eb.date_added <= :creation_date_to"
		args["creation_date_to"] = audience.Filter.EndDate
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

func (r *MySQLAudienceRepository) ListApplicationsWithFilters(ctx context.Context, pagination *domain.PaginationRequest, filter *domain.ApplicationFilterRequest, audience_filter *domain.AudienceCreationFilter) (*domain.PaginationResponse, error) {
	// Build base query for counting
	countQuery := `
        SELECT COUNT(*) 
        FROM estate_buys eb
        LEFT JOIN estate_houses h ON h.id = eb.house_id
		LEFT JOIN estate_deals_contacts edc ON edc.id = eb.contacts_id
        LEFT JOIN users u ON u.id = eb.manager_id
		LEFT JOIN estate_sells es ON es.id = eb.estate_sell_id
		LEFT JOIN estate_deals ed ON eb.deal_id = ed.id
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
			coalesce(TRIM(LOWER(REGEXP_REPLACE(REGEXP_SUBSTR(
                    passport_address,
                    '((г\\.|город )\\s*([^,\\s\\.]+))|(([^,\\s\\.]+)\\s(shah|shax|Ш(и|а)\\SРИ|ша\\Sар|город,|ш\\.))'
                ),
                '(г\\.|город\\s|\\sshah|\\sshax|\\sШ(и|а)\\SРИ|\\sша\\Sар|\\sгород,|\\sш\\.)', ''
            ))), "Не указано")
    	    AS region,
            DATEDIFF(NOW(), COALESCE(
                (SELECT MAX(log_date) 
                FROM estate_buys_statuses_log 
                WHERE estate_buy_id = eb.id 
                AND status_to = eb.status),
                eb.date_added
            )) AS days_in_status,
            COALESCE(h.complex_name, 'Не указано') AS project_name
        FROM estate_buys eb
        LEFT JOIN estate_deals_contacts edc ON edc.id = eb.contacts_id
        LEFT JOIN users u ON u.id = eb.manager_id
        LEFT JOIN estate_houses h ON h.id = eb.house_id
		LEFT JOIN estate_sells es ON es.id = eb.estate_sell_id
		LEFT JOIN estate_deals ed ON eb.deal_id = ed.id
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

	if filter.EndDate != nil && !filter.EndDate.IsZero() {
		whereConditions = append(whereConditions, "eb.date_added <= :created_at_to")
		args["created_at_to"] = filter.EndDate
	}

	if filter.StartDate != nil && !filter.StartDate.IsZero() {
		whereConditions = append(whereConditions, "eb.date_added >= :created_at_from")
		args["created_at_from"] = filter.StartDate
	}

	if filter.RegionName != "" {
		whereConditions = append(whereConditions,
			`coalesce(TRIM(LOWER(REGEXP_REPLACE(REGEXP_SUBSTR(
                    edc.passport_address,
                    '((г\\.|город )\\s*([^,\\s\\.]+))|(([^,\\s\\.]+)\\s(shah|shax|Ш(и|а)\\SРИ|ша\\Sар|город,|ш\\.))'
                ),
                '(г\\.|город\\s|\\sshah|\\sshax|\\sШ(и|а)\\SРИ|\\sша\\Sар|\\sгород,|\\sш\\.)', ''
            ))), "Не указано") = :region_name`)
		args["region_name"] = filter.RegionName
	}

	// OMAGAD this is crap code
	// ids:=strings.Join(filter.AudienceIDs, ", ")
	// r.logger.Info("audience_ids", zap.Any("audience_ids", ids))
	// whereConditions = append(whereConditions, " eb.id IN (:audience_ids)")
	// args["audience_ids"] = filter.AudienceIDs
	//whereConditions = append(whereConditions, " INSTR((:audience_ids), eb.id ) > 0")
	//args["audience_ids"] = ids

	//AUDIENCE filters
	if audience_filter.StatusNames != nil {
		whereConditions = append(whereConditions, "eb.status_name IN (:status_names)")
		args["status_names"] = audience_filter.StatusNames
	}
	if audience_filter.StartDate != nil && !audience_filter.StartDate.IsZero() {
		whereConditions = append(whereConditions, "eb.date_added >= :creation_date_from")
		args["creation_date_from"] = audience_filter.StartDate
	}
	if audience_filter.EndDate != nil && !audience_filter.StartDate.IsZero() {
		whereConditions = append(whereConditions, "eb.date_added <= :creation_date_to")
		args["creation_date_to"] = audience_filter.EndDate
	}
	if audience_filter.StatusNames != nil {
		whereConditions = append(whereConditions, "eb.status_name IN (:status_names)")
		args["status_names"] = audience_filter.StatusNames
	}
	if audience_filter.RegectionReasonNames != nil {
		whereConditions = append(whereConditions, "ebrs.name IN (:reason_names)")
		args["reason_names"] = audience_filter.RegectionReasonNames
	}
	if audience_filter.NonTargetReasonNames != nil {
		whereConditions = append(whereConditions, "ebrs.name IN (:reason_names)")
		args["reason_names"] = audience_filter.NonTargetReasonNames
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
			direction := ""
			if strings.ToUpper(filter.OrderDirection) == "DESC" {
				direction = "DESC"
			} else if strings.ToUpper(filter.OrderDirection) == "ASC" {
				direction = "ASC"
			} else {
				direction = "DESC"
			}
			orderClause = fmt.Sprintf(" ORDER BY %s %s", dbField, direction)
		}
	}

	countQuery, params, err := sqlx.Named(countQuery, args)
	if err != nil {
		return nil, fmt.Errorf("failed to bind named params: %w", err)
	}

	countQuery, params, err = sqlx.In(countQuery, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to expand IN clause: %w", err)
	}

	countQuery = r.db.Rebind(countQuery)

	if err := r.db.GetContext(ctx, &totalItems, countQuery, params...); err != nil {
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

	query, queryArgs, err = sqlx.In(query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to expand IN clause: %w", err)
	}

	query = r.db.Rebind(query)

	var items []domain.Application
	if err := r.db.SelectContext(ctx, &items, query, queryArgs...); err != nil {
		return nil, fmt.Errorf("select applications: %w", err)
	}

	// appls := []domain.Application{}

	// var raw_items []domain.Application
	// if err := r.db.SelectContext(ctx, &raw_items, query, queryArgs...); err != nil {
	// 	return nil, fmt.Errorf("select applications: %w", err)
	// }

	// var items []domain.Application
	// for _, item := range raw_items {
	// 	for _, application := range appls {
	// 		if item.ID == application.ID {
	// 			items = append(items, item)
	// 			break
	// 		}
	// 	}
	// }

	// if filter.AudienceIDs != nil {
	// 	appls, err := r.GetApplicationsByAudienceFilter(ctx, *audience_filter)

	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to get applications by audience filter: %w", err)
	// 	}
	// }

	r.logger.Info("audience_name", zap.Any("audience_name", filter.AudienceName))
	r.logger.Info("audience_ids_count", zap.Any("audience_ids_count", len(filter.AudienceIDs)))
	r.logger.Info("query", zap.Any("query", query))

	headers := []domain.Header{
		{
			Name:          "id",
			IsID:          true,
			IsAsideHeader: true,
			Title:         "ID",
			IsVisible:     true,
			IsAdditional:  false,
			Format:        "number",
			IsSortable:    true,
		},
		{
			Name:          "name",
			IsID:          false,
			IsAsideHeader: false,
			Title:         "ФИО",
			IsVisible:     true,
			IsAdditional:  false,
			Format:        "string",
			IsSortable:    true,
		},
		{
			Name:         "created_at",
			IsID:         false,
			Title:        "Дата",
			IsVisible:    true,
			IsAdditional: false,
			Format:       "date",
			IsSortable:   true,
		},
		{
			Name:          "status_name",
			IsID:          false,
			IsAsideHeader: false,
			Title:         "Этап",
			IsVisible:     true,
			IsAdditional:  false,
			Format:        "enum",
			IsSortable:    true,
		},
		{
			Name:          "phone",
			IsID:          false,
			IsAsideHeader: false,
			Title:         "Номер телефона",
			IsVisible:     true,
			IsAdditional:  false,
			Format:        "string",
			IsSortable:    true,
		},
		{
			Name:          "manager_name",
			IsID:          false,
			IsAsideHeader: false,
			Title:         "Посредник",
			IsVisible:     true,
			IsAdditional:  false,
			Format:        "string",
			IsSortable:    true,
		},
		{
			Name:          "property_type",
			IsID:          false,
			IsAsideHeader: false,
			Title:         "Тип недвижимости",
			IsVisible:     true,
			IsAdditional:  false,
			Format:        "enum",
			IsSortable:    true,
		},
		{
			Name:          "project_name",
			IsID:          false,
			IsAsideHeader: false,
			Title:         "Проект",
			IsVisible:     true,
			IsAdditional:  false,
			Format:        "string",
		},
		{
			Name:          "region",
			IsID:          false,
			IsAsideHeader: false,
			Title:         "Регион",
			IsVisible:     true,
			IsAdditional:  false,
			Format:        "string",
		},
		{
			Name:          "reason_name",
			IsID:          false,
			IsAsideHeader: false,
			Title:         "Причина статуса",
			IsVisible:     true,
			IsAdditional:  false,
			Format:        "string",
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
            COALESCE(h.complex_name, 'Не указано') AS project_name
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

	if filter.EndDate != nil && !filter.EndDate.IsZero() {
		whereConditions = append(whereConditions, "eb.date_added <= ?")
		args = append(args, filter.EndDate)
	}

	if filter.StartDate != nil && !filter.StartDate.IsZero() {
		whereConditions = append(whereConditions, "eb.date_added >= ?")
		args = append(args, filter.StartDate)
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
	try_date_query := `SELECT * FROM macro_bi_cmp_528.estate_buys eb WHERE 1=1`
	if filter.StartDate != nil && !filter.StartDate.IsZero() {
		try_date_query = try_date_query + ` AND eb.date_added >= '` + filter.StartDate.Format("2006-01-02") + `'`
	}
	if filter.EndDate != nil && !filter.EndDate.IsZero() {
		try_date_query = try_date_query + ` AND eb.date_added <= '` + filter.EndDate.Format("2006-01-02") + `'`
	}
	try_date_query = try_date_query + ` LIMIT 1`
	if response, err := r.db.QueryContext(ctx, try_date_query); err != nil || !response.Next() {
		r.logger.Error("No appliccations for this dates: ", zap.Error(err))
		return nil, fmt.Errorf("no applications for this dates: %w", err)
	}

	columns_query := `
		SELECT DISTINCT 
    	    	coalesce(TRIM(LOWER(REGEXP_REPLACE(REGEXP_SUBSTR(
                    passport_address,
                    '((г\\.|город )\\s*([^,\\s\\.]+))|(([^,\\s\\.]+)\\s(shah|shax|Ш(и|а)\\SРИ|ша\\Sар|город,|ш\\.))'
                ),
                '(г\\.|город\\s|\\sshah|\\sshax|\\sШ(и|а)\\SРИ|\\sша\\Sар|\\sгород,|\\sш\\.)', ''
            ))), "Не указано")
    	    AS city_name
    	FROM macro_bi_cmp_528.estate_deals_contacts ec
		LEFT JOIN macro_bi_cmp_528.estate_buys eb ON eb.contacts_id = ec.id
		ORDER BY city_name`

	rows_query := `SELECT distinct h.complex_name from macro_bi_cmp_528.estate_houses h`

	data_query := `
	SELECT 
    	coalesce(h.complex_name, 'Не указано') AS project,
    	coalesce(TRIM(LOWER(REGEXP_REPLACE(REGEXP_SUBSTR(
                    passport_address,
                    '((г\\.|город )\\s*([^,\\s\\.]+))|(([^,\\s\\.]+)\\s(shah|shax|Ш(и|а)\\SРИ|ша\\Sар|город,|ш\\.))'
                ),
                '(г\\.|город\\s|\\sshah|\\sshax|\\sШ(и|а)\\SРИ|\\sша\\Sар|\\sгород,|\\sш\\.)', ''
            ))), "Не указано") AS region,
    	COUNT(eb.id) AS application_count
	FROM estate_buys eb
	LEFT JOIN estate_houses h ON h.id = eb.house_id
	LEFT JOIN estate_deals_contacts edc ON edc.id = eb.contacts_id
	LEFT JOIN users u ON u.id = eb.manager_id
	LEFT JOIN estate_sells es ON es.id = eb.estate_sell_id
	LEFT JOIN estate_deals ed ON eb.deal_id = ed.id
	WHERE eb.company_id = 528 
		`

	data_group_query := `
		GROUP BY 
    		coalesce(TRIM(LOWER(REGEXP_REPLACE(REGEXP_SUBSTR(
                    passport_address,
                    '((г\\.|город )\\s*([^,\\s\\.]+))|(([^,\\s\\.]+)\\s(shah|shax|Ш(и|а)\\SРИ|ша\\Sар|город,|ш\\.))'
                ),
                '(г\\.|город\\s|\\sshah|\\sshax|\\sШ(и|а)\\SРИ|\\sша\\Sар|\\sгород,|\\sш\\.)', ''
            ))), "Не указано"),
    		coalesce(h.complex_name, "Не указано")
		ORDER BY 
     		coalesce(region,"Не указано"),
     		coalesce(h.complex_name, "Не указано");`

	// totals_query := `
	// 	SELECT
	// 		COALESCE(lower(TRIM(
	// 	        REGEXP_REPLACE(
	// 	            REGEXP_SUBSTR(
	// 	                passport_address,
	// 	                '((г\\.|город )\\s*([^,\\s\\.]+))|(([^,\\s\\.]+)\\s(shah|shax|Ш(и|а)\\SРИ|ша\\Sар|город,|ш\\.))'
	// 	            ),
	// 	            '(г\\.|город\\s|\\sshah|\\sshax|\\sШ(и|а)\\SРИ|\\sша\\Sар|\\sгород,|\\sш\\.)', ''
	// 	        )
	// 	    )), "Не указан")  AS region,
	// 		COUNT(eb.id) AS application_count
	// 	FROM
	// 		estate_buys eb
	// 	LEFT JOIN
	// 		estate_deals_contacts edc ON edc.id = eb.contacts_id
	// 	Where eb.company_id = 528 `

	// totals_group_query := `GROUP BY COALESCE(lower(TRIM(
	//         REGEXP_REPLACE(
	//             REGEXP_SUBSTR(
	//                 passport_address,
	//                 '((г\\.|город )\\s*([^,\\s\\.]+))|(([^,\\s\\.]+)\\s(shah|shax|Ш(и|а)\\SРИ|ша\\Sар|город,|ш\\.))'
	//             ),
	//             '(г\\.|город\\s|\\sshah|\\sshax|\\sШ(и|а)\\SРИ|\\sша\\Sар|\\sгород,|\\sш\\.)', ''
	//         )
	//     )), 'Не указан')`

	//COALESCE(h.complex_name, 'Не указан')
	if filter.Status == "" {
		data_query = data_query + ` AND (es.estate_sell_status_name = 'Сделка проведена' OR es.estate_sell_status_name = 'Сделка в работе' OR es.estate_sell_status_name = 'Бронь')
`
		//totals_query = totals_query + ` AND es.estate_sell_status_name = 'Сделка проведена' OR es.estate_sell_status_name = 'Сделка в работе' OR eb.status_name = 'Бронь'`
		//columns_query = columns_query + ` AND es.estate_sell_status_name = 'Сделка проведена' OR es.estate_sell_status_name = 'Сделка в работе' OR eb.status_name = 'Бронь'`
	} else {
		data_query = data_query + ` AND eb.status_name = '` + filter.Status + `'`
		//totals_query = totals_query + ` AND eb.status_name = '` + filter.Status + `'`
		//columns_query = columns_query + ` AND eb.status_name = '` + filter.Status + `'`
	}

	if filter.Project != "" {
		data_query = data_query + ` AND h.complex_name = '` + filter.Project + `'`
		//totals_query = totals_query + ` AND h.complex_name = '` + filter.Project + `'`
		//columns_query = columns_query + ` AND h.complex_name = '` + filter.Project + `'`
	}

	if filter.StartDate != nil && !filter.StartDate.IsZero() {
		data_query = data_query + ` AND eb.date_added >= '` + filter.StartDate.Format("2006-01-02") + `'`
		//totals_query = totals_query + ` AND eb.date_added >= '` + filter.StartDate.Format("2006-01-02") + `'`
		//columns_query = columns_query + ` AND eb.date_added >= '` + filter.StartDate.Format("2006-01-02") + `'`

	}

	if filter.EndDate != nil && !filter.EndDate.IsZero() {
		data_query = data_query + ` AND eb.date_added <= '` + filter.EndDate.Format("2006-01-02") + `'`
		//totals_query = totals_query + ` AND eb.date_added <= '` + filter.EndDate.Format("2006-01-02") + `'`
		// columns_query = columns_query + ` AND eb.date_added <= '` + filter.EndDate.Format("2006-01-02") + `'`

	}

	// r.logger.Info("date_added in IF", zap.Any("start",filter.StartDate.Format("2006-01-02")))
	// r.logger.Info("date_added in IF", zap.Any("end",filter.EndDate.Format("2006-01-02")))

	data_query = data_query + data_group_query
	//totals_query = totals_query + totals_group_query

	var cities []string
	if err := r.db.SelectContext(ctx, &cities, columns_query); err != nil {
		return nil, fmt.Errorf("select columns: %w", err)
	}

	var projects []string
	if err := r.db.SelectContext(ctx, &projects, rows_query); err != nil {
		return nil, fmt.Errorf("select rows: %w", err)
	}

	//var totals []domain.Total_row
	//if err := r.db.SelectContext(ctx, &totals, totals_query); err != nil {
	//	return nil, fmt.Errorf("select totaals: %w", err)
	//}

	//TODO: REWRITE HEADER CREATION BASED ON DATA RECIEVED
	// Prepare headers
	headers := []domain.Header{
		//{Name: "id", IsID: true, Title: "№", IsVisible: false, IsAsideHeader: false, IsAdditional: false, Format: "number"},
		{Name: "name_of_projects",
			Title:         "Наименование проектов",
			IsAsideHeader: true,
			IsVisible:     true,
			IsAdditional:  false,
			Format:        "string"},
	}

	footer := map[string]int{
		"id":               0,
		"name_of_projects": -1,
	}

	var data []domain.Data_row
	if err := r.db.SelectContext(ctx, &data, data_query); err != nil {
		return nil, fmt.Errorf("select data: %w", err)
	}

	var data_map = make(map[string]interface{})
	// Convert rawDataRow to DataRow with a composite key.
	for _, data_row := range data {
		data_map[data_row.Project+" "+data_row.Region] = data_row.ApplicationCount
	}

	var data_responce []map[string]interface{}

	for _, city := range cities {
		for _, project := range projects {
			if data_map[project+" "+city] != nil {
				headers = append(headers, domain.Header{
					Name:          city,
					Title:         city,
					IsAsideHeader: false,
					IsVisible:     true,
					IsAdditional:  true,
					Format:        "number",
				})
				break
			}
		}
	}

	for _, project := range projects {
		var data_bit = make(map[string]interface{})
		data_bit["name_of_projects"] = project
		for _, city := range cities {
			if data_map[project+" "+city] != nil {
				data_bit[city] = data_map[project+" "+city]
				footer[city] += data_map[project+" "+city].(int)
			} else {
				data_bit[city] = 0
			}
		}
		data_responce = append(data_responce, data_bit)
	}

	footerData := make(map[string]interface{})
	for k, v := range footer {
		footerData[k] = v
	}

	footerData["name_of_projects"] = "Общее"
	return &domain.RegionsResponse{
		Headers: headers,
		Data:    data_responce,
		Footer:  footerData,
	}, nil
}

func (r *MySQLAudienceRepository) GetCallCenterReportData(ctx context.Context, filter *domain.CallCenterReportFilter) (*domain.CallCenterReport, error) {

	StartDateCondition := ""
	EndDateCondition := ""

	if filter.StartDate != nil && !filter.StartDate.IsZero() {
		StartDateCondition = " AND ebsl.log_date >= "+filter.StartDate.Format("2006-01-02")
	}

	if filter.EndDate != nil && !filter.EndDate.IsZero() {
		EndDateCondition =  " AND ebsl.log_date <= "+filter.EndDate.Format("2006-01-02")
	}


	query := `
	SELECT 
	    t1.users_id,
        u.users_name,
	    COALESCE(total_requests, 0) AS total_requests,
	    COALESCE(target_requests, 0) AS target_requests,
	    COALESCE(appointed_visits, 0) AS appointed_visits,
        COALESCE(successful_visits, 0) AS successful_visits
	FROM (
	    -- Общее количество заявок
	    SELECT 
	        ebsl.users_id,
	        COUNT(DISTINCT ebsl.estate_buy_id) AS total_requests
	    FROM estate_buys_statuses_log ebsl
		WHERE 1=1 `+StartDateCondition+EndDateCondition+`
	    GROUP BY ebsl.users_id
	) t1
	LEFT JOIN (
	    -- Количество целевых заявок
	    SELECT 
	        ebsl.users_id,
	        COUNT(*) AS target_requests
	    FROM estate_buys_statuses_log ebsl
	    WHERE status_custom_to_name != 'Нецелевой' `+StartDateCondition+EndDateCondition+`
	    GROUP BY ebsl.users_id
	) t2 ON t1.users_id = t2.users_id
	LEFT JOIN (
	    -- Количество назначенных визитов
	    SELECT 
	        ebsl.users_id,
	        COUNT(*) AS appointed_visits
	    FROM estate_buys_statuses_log ebsl
	    WHERE (status_from_name IN ('Проверка', 'Подбор', 'Неразобранное')) 
	          AND status_custom_to_name = 'Назначенная встреча' `+StartDateCondition+EndDateCondition+`
	    GROUP BY ebsl.users_id
	) t3 ON t1.users_id = t3.users_id 
    LEFT JOIN (
		WITH LastCallCenterManager AS (
		-- Находим последнего менеджера из колл-центра перед изменением статуса на "Состоялся визит"
		SELECT 
			ebsl.estate_buy_id,
			MAX(ebsl.id) AS last_cc_log_id
		FROM estate_buys_statuses_log ebsl
		JOIN users u ON u.id = ebsl.users_id
		WHERE u.departments_id = 1903 `+StartDateCondition+EndDateCondition+`
		GROUP BY ebsl.estate_buy_id
	),
	FinalVisits AS (
		-- Фиксируем заявки, у которых статус изменился на "Состоялся визит"
		SELECT 
			estate_buy_id,
			users_id AS final_manager
		FROM estate_buys_statuses_log ebsl
		WHERE status_custom_to_name = 'Визит состоялся' `+StartDateCondition+EndDateCondition+`
	)
	-- Теперь считаем заявки для каждого менеджера
	SELECT 
		users_id,
		COUNT(DISTINCT lv.estate_buy_id) AS successful_visits
	FROM LastCallCenterManager lccm
	JOIN estate_buys_statuses_log ebsl ON ebsl.id = lccm.last_cc_log_id  -- Последняя запись от колл-центра
	JOIN FinalVisits lv ON lv.estate_buy_id = ebsl.estate_buy_id  -- Заявка должна быть завершена другим менеджером
	JOIN users u ON u.id = ebsl.users_id  -- Берём ID последнего менеджера колл-центра
	WHERE 1=1 `+StartDateCondition+EndDateCondition+`
	GROUP BY u.id
	ORDER BY successful_visits DESC
    ) t4 ON t1.users_id = t4.users_id 
    LEFT JOIN
		users u on t1.users_id = u.id
		WHERE u.departments_id=1903
		ORDER BY users_id;
	`

	var metrics []domain.ManagerMetrics

	if err := r.db.SelectContext(ctx, &metrics, query); err != nil {
		return nil, fmt.Errorf("get sales metrics: %w", err)
	}

	// Calculate conversions and footer
	footer := domain.ManagerMetrics{ManagerName: "Итого"}
	for i := range metrics {
		m := &metrics[i]

		// Calculate basic conversions
		if m.TotalInquiries > 0 {
			m.TargetConversion = float64(m.TargetInquiries) / float64(m.TotalInquiries+m.TargetInquiries) // * 100
		}
		if m.TargetInquiries > 0 {
			m.VisitConversion = float64(m.AppointedVisits) / float64(m.TargetInquiries+m.AppointedVisits) // * 100
			m.LeadToVisit = float64(m.CompletedVisits) / float64(m.TargetInquiries+m.CompletedVisits)     // * 100
		}
		if m.AppointedVisits > 0 {
			m.VisitSuccess = float64(m.CompletedVisits) / float64(m.AppointedVisits+m.CompletedVisits) // * 100
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
		footer.TotalInquiries += m.TotalInquiries
		footer.TargetInquiries += m.TargetInquiries
		footer.AppointedVisits += m.AppointedVisits
		footer.CompletedVisits += m.CompletedVisits
		//totals.Bookings += m.Bookings
		//totals.Contracts += m.Contracts
	}

	// Calculate total conversions
	if footer.TotalInquiries > 0 {
		footer.TargetConversion = float64(footer.TargetInquiries) / float64(footer.TotalInquiries+footer.TargetInquiries) // * 100
	}
	if footer.TargetInquiries > 0 {
		footer.VisitConversion = float64(footer.AppointedVisits) / float64(footer.TargetInquiries+footer.AppointedVisits) // * 100
		footer.LeadToVisit = float64(footer.CompletedVisits) / float64(footer.TargetInquiries+footer.CompletedVisits)     // * 100
	}
	if footer.AppointedVisits > 0 {
		footer.VisitSuccess = float64(footer.CompletedVisits) / float64(footer.AppointedVisits+footer.CompletedVisits) // * 100
	}

	// Create headers
	headers := []domain.Header{
		{Name: "manager_name", IsAsideHeader: true, Title: "ФИО менеджера", IsVisible: true, IsAdditional: false, Format: "string"},
		{Name: "total_inquiries", IsAsideHeader: false, Title: "Всего обращений", IsVisible: true, IsAdditional: true, Format: "number"},
		{Name: "target_inquiries", IsAsideHeader: false, Title: "Целевые", IsVisible: true, IsAdditional: false, Format: "number"},
		{Name: "target_conversion", IsAsideHeader: false, Title: "Конверсия в целевые", IsVisible: true, IsAdditional: false, Format: "percent"},
		{Name: "appointed_visits", IsAsideHeader: false, Title: "Назначено визитов", IsVisible: true, IsAdditional: false, Format: "number"},
		{Name: "visit_conversion", IsAsideHeader: false, Title: "Конверсия в визиты", IsVisible: true, IsAdditional: false, Format: "percent"},
		{Name: "completed_visits", IsAsideHeader: false, Title: "Визиты состоялись", IsVisible: true, IsAdditional: false, Format: "number"},
		{Name: "visit_success", IsAsideHeader: false, Title: "Конверсия визитов", IsVisible: true, IsAdditional: false, Format: "percent"},
		{Name: "lead_to_visit", IsAsideHeader: false, Title: "Конверсия лид->визит", IsVisible: true, IsAdditional: false, Format: "percent"},

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
		Footer:  footer,
	}, nil
}

func (r *MySQLAudienceRepository) GetStatusDurationReport(ctx context.Context, filter *domain.StatusDurationFilter) (*domain.StatusDurationResponse, error) {
	query := `
        WITH StatusDurations AS (
            SELECT 
                sl.status_to_name as status_name,
                sl.estate_buy_id,
                TIMESTAMPDIFF(DAY, 
                    sl.log_date,
                    COALESCE(
                        LEAD(sl.log_date) OVER (PARTITION BY sl.estate_buy_id ORDER BY sl.log_date),
                        NOW()
                    )
                ) as days_in_status
            FROM estate_buys_statuses_log sl
            WHERE sl.company_id = 528
            AND sl.log_date BETWEEN ? AND ?
        )
        SELECT 
            status_name,
            AVG(days_in_status) as avg_days,
            COUNT(DISTINCT estate_buy_id) as total_requests,
            SUM(CASE WHEN days_in_status > ? THEN 1 ELSE 0 END) as over_threshold
        FROM StatusDurations
        GROUP BY status_name
        ORDER BY avg_days DESC
    `
	// Подготовка аргументов для SQL-запроса
	args := []interface{}{}
	if filter.EndDate != nil && !filter.EndDate.IsZero() {
		args = append(args, filter.EndDate.Format("2006-01-02"))
	} else {
		args = append(args, time.Now().AddDate(-1, 0, 0).Format("2006-01-02"))
	}

	if filter.StartDate != nil && !filter.StartDate.IsZero() {
		args = append(args, filter.StartDate.Format("2006-01-02"))
	} else {
		args = append(args, time.Now().Format("2006-01-02"))
	}

	if filter.ThresholdDays <= 0 {
		filter.ThresholdDays = 1
	}
	args = append(args, filter.ThresholdDays)

	// Выполняем запрос к базе данных
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query status durations: %w", err)
	}
	defer rows.Close()

	// Переменные для хранения результатов
	statusMap := make(map[string]domain.StatusDuration)
	var totalRequests int
	avgDaysMap := make(map[string]float64)

	for rows.Next() {
		var status domain.StatusDuration
		if err := rows.Scan(&status.StatusName, &status.AverageDays, &status.TotalRequests, &status.OverThreshold); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		statusMap[status.StatusName] = status
		totalRequests += status.TotalRequests
		avgDaysMap[status.StatusName] = status.AverageDays
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Формируем заголовки
	headers := []domain.Header{
		{Name: "id", Title: "№", IsVisible: false, IsID: true, Format: "number"},
		{Name: "status_name", Title: "Наименование статуса", IsVisible: true, IsID: false, Format: "string"},
	}
	for statusName := range statusMap {
		headers = append(headers, domain.Header{
			Name:         statusName,
			Title:        statusName,
			IsVisible:    true,
			IsAdditional: true,
			Format:       "number",
		})
	}

	// Формируем данные
	data := []map[string]interface{}{
		{
			"id":          1,
			"status_name": "Количество заявок",
		},
	}
	for statusName, status := range statusMap {
		data[0][statusName] = status.TotalRequests
	}

	// Формируем футер
	footer := map[string]interface{}{
		"id":          0,
		"status_name": "Среднее время",
	}
	for statusName, avgDays := range avgDaysMap {
		footer[statusName] = fmt.Sprintf("%.0f дней", avgDays)
	}

	return &domain.StatusDurationResponse{
		Headers: headers,
		Data:    data,
		Footer:  footer,
	}, nil
}

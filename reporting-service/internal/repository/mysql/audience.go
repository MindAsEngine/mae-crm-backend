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
	query = `SELECT Distinct geo_complex_name FROM macro_bi_cmp_528.geo_city_complex`
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

	totalPages := int(math.Ceil(float64(totalItems) / float64(pagination.PageSize)))

	return &domain.PaginationResponse{
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
    // Get unique regions from correct column
    regionsQuery := `
        SELECT DISTINCT passport_bithplace 
        FROM estate_deals_contacts 
        WHERE passport_bithplace IS NOT NULL AND passport_bithplace != ''
        ORDER BY passport_bithplace
    `
    
    var regions []string
    if err := r.db.SelectContext(ctx, &regions, regionsQuery); err != nil {
        return nil, fmt.Errorf("get regions: %w", err)
    }

    // Build dynamic query
    baseQuery := `
        SELECT 
            gcc.id,
            gcc.geo_complex_name as name,
    `
    
    // Add dynamic CASE statements for each region
    caseClauses := []string{}
    for i := 0; i < len(regions); i++ {
        caseClauses = append(caseClauses, fmt.Sprintf(`
            COUNT(CASE WHEN passport_bithplace = ? THEN 1 END) as region%d`,
            i+1))
    }

    query := baseQuery + strings.Join(caseClauses, ",") + `
        FROM geo_city_complex gcc
        LEFT JOIN estate_houses h ON h.complex_id = gcc.geo_complex_id
		LEFT JOIN estate_deals_contacts edc ON edc.id = h.contacts_id
        WHERE gcc.company_id = 528
    `

    args := []interface{}{}
    // Add region args
    for _, region := range regions {
        args = append(args, region)
    }

    // Add filters
    if filter.Search != "" {
        query += " AND gcc.geo_complex_name LIKE ?"
        args = append(args, "%"+filter.Search+"%")
    }

    if filter.StartDate != nil {
        query += " AND h.updated_at >= ?"
        args = append(args, filter.StartDate)
    }

    if filter.EndDate != nil {
        query += " AND h.updated_at <= ?"
        args = append(args, filter.EndDate)
    }

    query += " GROUP BY gcc.id, gcc.geo_complex_name"

    // Add sorting
    if filter.Sort != "" {
        parts := strings.Split(filter.Sort, "_")
        if len(parts) == 2 {
            sortableFields := map[string]string{
                "name":    "gcc.geo_complex_name",
            }
            
            field := parts[0]
            direction := strings.ToUpper(parts[1])
            
            if dbField, exists := sortableFields[field]; exists && (direction == "ASC" || direction == "DESC") {
                query += fmt.Sprintf(" ORDER BY %s %s", dbField, direction)
            }
        }
    } else {
        query += " ORDER BY gcc.sort_order ASC, gcc.geo_complex_name ASC"
    }

    // Debug log
    r.logger.Debug("executing query", 
        zap.String("query", query),
        zap.Any("args", args))

	print(query)

    // Execute query
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("execute query: %w", err)
    }
    defer rows.Close()

    // Process results
    var data []domain.RegionData
    footer := domain.RegionData{
        NameOfProject: "Общее",
        RegionCounts:  make(map[int]int),
    }

    for rows.Next() {
        var item domain.RegionData
        item.RegionCounts = make(map[int]int)
        
        scanArgs := []interface{}{&item.ID, &item.NameOfProject}
        for i := range regions {
            var count int
            scanArgs = append(scanArgs, &count)
            item.RegionCounts[i+1] = count
        }
        
        if err := rows.Scan(scanArgs...); err != nil {
            return nil, fmt.Errorf("scan row: %w", err)
        }
        
        // Update footer totals
        for region, count := range item.RegionCounts {
            footer.RegionCounts[region] += count
        }
        
        data = append(data, item)
    }

    // Create headers
    headers := []domain.Header{
        {Name: "id", IsID: true, Title: "№", IsVisible: false, IsAdditional: false, Format: "number"},
        {Name: "name_of_projects", Title: "Наименование проектов", IsVisible: true, IsAdditional: false, Format: "string"},
    }

    for i, region := range regions {
        headers = append(headers, domain.Header{
            Name:         fmt.Sprintf("region%d", i+1),
            Title:        region,
            IsVisible:    true,
            IsAdditional: true,
            Format:       "number",
        })
    }

    return &domain.RegionsResponse{
        Headers: headers,
        Data:    data,
        Footer:  footer,
    }, nil
}
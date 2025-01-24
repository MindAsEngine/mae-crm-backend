package domain

import "time"

type PaginationResponse struct {
	Headers    []Header    `json:"header"`
	Items      interface{} `json:"items"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
}

//application responces

type ApplicationFilterResponce struct {
	//OrderField     string 	  `json:"order_field"`
	//OrderDirection string 	  `json:"order_direction"`
	Statuses      []string `json:"status_names" form:"status_names"`
	//StatusDuration int64  	  `json:"status_duration,omitempty" form:"status_duration"`
	ProjectNames  []string `json:"project_names" form:"project_names"`
	PropertyTypes []string `json:"property_types" form:"property_types"`
	AudienceNames []string `json:"audience_names" form:"audience_names"`
	RegionNames   []string `json:"regions" form:"regions"`
	//CreatedAtFrom  *time.Time `json:"created_at_from" form:"created_at_from"`
	//CreatedAtTo    *time.Time `json:"created_at_to" form:"created_at_to"`
	//DeadlinePassed bool       `json:"deadline_passed" form:"deadline_passed"`
}

//audience reports
type AudienceResponse struct {
	ID                 int64         `json:"id"`
	Name               string        `json:"name"`
	Integrations       []Integration `json:"integrations"`
	//Application_ids    []int64       `json:"application_ids"`
	Applications_count int       `json:"application_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type IntegrationsCreateResponse struct {
	Integrations []Integration `json:"integrations"`
}

//region report
type Header struct {
	Name          string `json:"name"`
	IsAsideHeader bool   `json:"is_aside_header"`
	IsID          bool   `json:"is_id"`
	Title         string `json:"title"`
	IsVisible     bool   `json:"is_visible"`
	IsAdditional  bool   `json:"is_additional"`
	IsSortable    bool   `json:"is_sortable"`
	Format        string `json:"format"`
}

type Data_row struct {
	Project          string `db:"project"`
	Region           string `db:"region"`
	ApplicationCount int    `db:"application_count"`
}

type RegionsResponse struct {
	Headers []Header                 `json:"headers"`
	Data    []map[string]interface{} `json:"data"`
	Footer  map[string]interface{}   `json:"footer"`
}

type StatusDurationResponse struct {
    Headers []Header                 `json:"headers"`
    Data    []map[string]interface{} `json:"data"`
	Footer  map[string]interface{}   `json:"footer"`
}
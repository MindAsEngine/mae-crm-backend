package domain

import (
	//"github.com/google/uuid"
	"time"
)

type Application struct {
	ID             int64     `json:"id" db:"id"`
	CreatedAt      time.Time `json:"created_at" db:"date_added"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	StatusName     string    `json:"status_name" db:"status_name"`
	ManagerID      int64     `json:"manager_id" db:"manager_id"`
	ManagerName    string    `json:"manager_name" db:"manager_name"`
	ClientID       int64     `json:"client_id" db:"contacts_id"`
	StatusID       int64     `json:"status_id" db:"status"`
	ReasonName     string    `json:"reason_name" db:"name"`
	ReasonId       int64     `json:"status_reason_id,omitempty" db:"status_reason_id"`
	Name           string    `json:"name" db:"client_name"`
	Phone          string    `json:"phone" db:"phone"`
	BirthPlace     string    `json:"birth_place" db:"birth_place"`
	PropertyType   string    `json:"property_type" db:"property_type"`
	ProjectName    string    `json:"project_name" db:"project_name"`
	StatusDuration int64     `json:"days_in_status" db:"days_in_status"`
}

type ApplicationFilter struct {
	OrderField     string `json:"order_field"`
	OrderDirection string `json:"order_direction"`
	Status         string `json:"status_name" form:"status_name"`
	StatusDuration int64  `json:"status_duration,omitempty" db:"status_duration"`
	ProjectName    string `json:"project_name" form:"project_name"`
	DaysInStatus   int    `json:"days_in_status" form:"days_in_status"`
	PropertyType   string `json:"property_type" form:"property_type"`
}

type Audience struct {
	ID               int64          `json:"id" db:"id"`
	Name             string         `json:"name" db:"name"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
	Application_ids  []int64        `json:"request_ids" db:"request_ids"`
	Applications     []Application  `json:"requests" db:"requests"`
	Integrations     []Integration  `json:"integrations" db:"integrations"`
	IntegrationNames []string       `json:"integration_names" db:"integration_names"`
	Filter           AudienceFilter `json:"filter" db:"filter"`
}

type Integration struct {
	ID          int64  `json:"id" db:"id"`
	AudienceID  int64  `json:"audience_id" db:"audience_id"`
	CabinetName string `json:"cabinet_name" db:"cabinet_name"` // Например: Google Ads, Facebook
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

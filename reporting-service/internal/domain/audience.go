package domain

import (
	//"github.com/google/uuid"
	"time"
)

type Application struct {
	ID              int64      `json:"id" db:"id"`
	CreatedAt       time.Time  `json:"created_at" db:"date_added"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	StatusName      string     `json:"status_name" db:"status_name"`
	ManagerID       int64      `json:"manager_id" db:"manager_id"`
	ClientID        int64      `json:"client_id" db:"contacts_id"`
	StatusID        int64      `json:"status_id" db:"status"`
	ReasonName      string     `json:"reason_name" db:"name"`
	ReasonId        int64      `json:"status_reason_id,omitempty" db:"status_reason_id"`
	ClientData      ClientData `json:"client_data" db:"client_data"`
}

type ClientData struct {
	FIO        string `json:"fio" db:"fio"`
	Phone      string `json:"phone" db:"phone"`
	BirthPlace string `json:"birth_place" db:"birth_place"`
}

type ApplicationFilter struct {
	OrderField     string `json:"order_field"`
	OrderDirection string `json:"order_direction"`
	StatusID       int64  `json:"status_id,omitempty" db:"status_id"`
	StatusDuration int64  `json:"status_duration,omitempty" db:"status_duration"`
	ProjectID      int64  `json:"project_id,omitempty" db:"project_id"`
	EstateTypeID   int64  `json:"estate_type_id,omitempty" db:"estate_type_id"`
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

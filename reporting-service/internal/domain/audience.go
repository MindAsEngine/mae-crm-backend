package domain

import (
	//"github.com/google/uuid"
	"time"
)


type Application struct {
	ID                 int64      `json:"id" db:"id"`
	CreatedAt          time.Time  `json:"created_at" db:"date_added"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	Status             string     `json:"status" db:"status_name"`
	RejectionReason    string     `json:"rejection_reason,omitempty" db:"rejection_reason"`
	NonTargetReason    string     `json:"non_target_reason,omitempty" db:"non_target_reason"`
	ResponsibleUserID  int64      `json:"responsible_user_id" db:"manager_id"`
	// ClientData         json.RawMessage  `json:"client_data" db:"client_data"`
}

type Audience struct {
	ID           int64          `json:"id" db:"id"`
	Name         string         `json:"name" db:"name"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
	Requests     []Application  `json:"request_ids" db:"request_ids"`
	Integrations []Integration  `json:"integrations" db:"integrations"`
	Filter       AudienceFilter `json:"filter" db:"filter"`
}


type Integration struct {
	ID          int64  `json:"id" db:"id"`
	AudienceID  int64  `json:"audience_id" db:"audience_id"`
	CabinetName string `json:"cabinet_name" db:"cabinet_name"` // Например: Google Ads, Facebook
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}
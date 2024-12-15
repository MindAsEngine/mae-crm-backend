package domain

import (
	//"github.com/google/uuid"
	"time"
)

type Audience struct {
	ID               int64     `json:"id" db:"id"`
	Name             string        `json:"name" db:"name"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time     `json:"last_updated" db:"last_updated"`
	Requests         []Request     `json:"request_ids" db:"request_ids"`
	Integrations     []Integration `json:"integrations" db:"integrations"`
	Filter 		 	 AudienceFilter `json:"filter" db:"filter"`
}

type AudienceFilter struct {
	ID    			 *int64 `json:"id" db:"id"`
	IntegrationID    *int64 `json:"integration_id" db:"request_ids"`
	CreationDateFrom *time.Time `json:"created_at" db:"created_at"`
	CreationDateTo   *time.Time `json:"last_updated" db:"last_updated"`
	Statuses         []string   `json:"statuses" db:"statuses"`
	RejectionReasons []string   `json:"rejection_reasons" db:"rejection_reasons"`
	NonTargetReasons []string   `json:"non_target_reasons" db:"non_target_reasons"`
}

type Integration struct {
	ID         int    `json:"id"`
	AudienceID int    `json:"audience_id"`
	Platform   string `json:"platform"` // Например: Google Ads, Facebook
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type AudienceMessage struct {
    AudienceID    int64 `json:"audience_id"`
    UpdatedAt     time.Time `json:"updated_at"`
    RequestCount  int       `json:"request_count"`
    LastRequestID int64     `json:"last_request_id"`
    Status        string    `json:"status"`
    Filter        struct {
        CreationDateFrom *time.Time `json:"creation_date_from,omitempty"`
        CreationDateTo   *time.Time `json:"creation_date_to,omitempty"`
        Statuses        []string   `json:"statuses,omitempty"`
        RejectionReasons []string   `json:"rejection_reasons,omitempty"`
        NonTargetReasons []string   `json:"non_target_reasons,omitempty"`
    } `json:"filter"`
}

type AudienceCreateRequest struct {
    Name   string         `json:"name" validate:"required"`
    Filter AudienceFilter `json:"filter" validate:"required"`
}


type AudienceResponse struct {
    ID        int64      `json:"id"`
    Name      string         `json:"name"`
    Filter    AudienceFilter `json:"filter"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
}

type ErrorResponse struct {
    Error       string `json:"error"`
    Description string `json:"description,omitempty"`
}
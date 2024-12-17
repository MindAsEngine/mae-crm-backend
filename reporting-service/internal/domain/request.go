package domain

import (
	"time"
	//"encoding/json"
	//"github.com/google/uuid"
)


type AudienceFilter struct {
	CreationDateFrom *time.Time `json:"creation_date_from"`
	CreationDateTo   *time.Time `json:"creation_date_to"`
	Statuses         []string   `json:"statuses"`
	RejectionReasons []string   `json:"rejection_reasons"`
	NonTargetReasons []string   `json:"non_target_reasons"`
}


type AudienceMessage struct {
	AudienceID    int64          `json:"audience_id"`
	UpdatedAt     time.Time      `json:"updated_at"`
	RequestCount  int            `json:"request_count"`
	LastRequestID int64          `json:"last_request_id"`
	Status        string         `json:"status"`
	Filter        AudienceFilter `json:"filter"`
}

type AudienceCreateRequest struct {
	Name   string         `json:"name" validate:"required"`
	Filter AudienceFilter `json:"filter" validate:"required"`
}

type IntegrationsCreateRequest struct {
    CabinetName string     `json:"cabinet_name"`
    AudienceIds []int64    `json:"audience_ids"`
}


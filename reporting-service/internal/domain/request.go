package domain

import (
	"time"
	//"encoding/json"
	//"github.com/google/uuid"
)

type AudienceMessage struct {
	AudienceID    int64          `json:"audience_id"`
	UpdatedAt     time.Time      `json:"updated_at"`
	RequestCount  int            `json:"request_count"`
	LastRequestID int64          `json:"last_request_id"`
	Status        string         `json:"status"`
	Filter        AudienceFilter `json:"filter"`
}

type AudienceFilter struct {
	CreationDateFrom   *time.Time `json:"creation_date_from"`
	CreationDateTo     *time.Time `json:"creation_date_to"`
	StatusNames		   []string   `json:"status_names"`
	StatusIDs          []int64    `json:"status_ids"`
	ReasonNames        []string   `json:"reason_names"`
	RejectionReasonIDs []int64    `json:"rejection_reason_ids"`
	NonTargetReasonIDs []int64    `json:"non_target_reason_ids"`
}


type AudienceCreateRequest struct {
	Name   string         `json:"name" validate:"required"`
	Filter AudienceFilter `json:"filter" validate:"required"`
}

type IntegrationsCreateRequest struct {
    CabinetName string     `json:"cabinet_name"`
    AudienceIds []int64    `json:"audience_ids"`
}


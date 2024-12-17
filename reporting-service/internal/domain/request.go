package domain

import (
	"time"
	//"encoding/json"
	//"github.com/google/uuid"
)

type AudienceMessage struct {
	AudienceID    int64          `json:"audience_id"`
	Applications  []Application  `json:"applications"`
	Filter        AudienceFilter `json:"filter"`
}

type AudienceFilter struct {
	CreationDateFrom     *time.Time `json:"creation_date_from"`
	CreationDateTo       *time.Time `json:"creation_date_to"`
	StatusNames          []string   `json:"statuses"`
	StatusIDs            []int64    `json:"status_ids"`
	RegectionReasonNames []string   `json:"rejection_reasons"`
	NonTargetReasonNames []string   `json:"non_target_reasons"`
	RejectionReasonIDs   []int64    `json:"rejection_reason_ids"`
	NonTargetReasonIDs   []int64    `json:"non_target_reason_ids"`
}

type AudienceCreateRequest struct {
	Name   string         `json:"name" validate:"required"`
	Filter AudienceFilter `json:"filter" validate:"required"`
}

type IntegrationsCreateRequest struct {
	CabinetName string  `json:"cabinet_name"`
	AudienceIds []int64 `json:"audience_ids"`
}

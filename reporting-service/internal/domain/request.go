package domain

import (
	"time"
	//"encoding/json"
	//"github.com/google/uuid"
)

type AudienceMessage struct {
	CurrentChunk    int           `json:"current_chunk"`
	TotalChunks     int           `json:"total_chunks"`
	AudienceName    string        `json:"audience_name"`
	AudienceID      int64         `json:"audience_id"`
	Integrations    []Integration `json:"integrations"`
	Application_ids []int64       `json:"application_ids"`
}

type AudienceCreationFilter struct {
	ID                   int64      `json:"id" db:"id"`
	AudienceId           int64      `json:"audience_id" db:"audience_id"`
	CreationDateFrom     *time.Time `json:"creation_date_from" db:"creation_date_from"`
	CreationDateTo       *time.Time `json:"creation_date_to" db:"creation_date_to"`
	StatusNames          []string   `json:"statuses" db:"status_names"`
	StatusIDs            []int64    `json:"status_ids" db:"status_ids"`
	RegectionReasonNames []string   `json:"rejection_reasons" db:"rejection_reasons"`
	NonTargetReasonNames []string   `json:"non_target_reasons" db:"non_target_reasons"`
	ReasonIDs            []int64    `json:"reason_ids" db:"reason_ids"`
}

type AudienceFilter struct {}

type RegionFilter struct {
	Search    string    `json:"search"`
	StartDate *time.Time `json:"start_date" validate:"required"`
	EndDate   *time.Time `json:"end_date" validate:"required"`
	Sort      string    `json:"sort"` //
}

type AudienceCreateRequest struct {
	Name   string         `json:"name" validate:"required"`
	Filter AudienceCreationFilter `json:"filter" validate:"required"`
}

type IntegrationsCreateRequest struct {
	CabinetName string  `json:"cabinet_name"`
	AudienceIds []int64 `json:"audience_ids"`
}

type PaginationRequest struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

package domain

import (
	"time"
	//"encoding/json"
	//"github.com/google/uuid"
)

type AudienceMessage struct {
	AudienceID        int64    `json:"audience_id"`
	Integration_names []string `json:"integration_names"`
	Application_ids   []int64  `json:"application_ids"`
}

type AudienceFilter struct {
	ID 				 	 int64      `json:"id" db:"id"`
	AudienceId		     int64      `json:"audience_id" db:"audience_id"`
	CreationDateFrom     *time.Time `json:"creation_date_from" db:"creation_date_from"`
	CreationDateTo       *time.Time `json:"creation_date_to" db:"creation_date_to"`
	StatusNames          []string   `json:"statuses" db:"status_names"`
	StatusIDs            []int64    `json:"status_ids" db:"status_ids"`
	RegectionReasonNames []string   `json:"rejection_reasons" db:"rejection_reasons"`
	NonTargetReasonNames []string   `json:"non_target_reasons" db:"non_target_reasons"`
	ReasonIDs            []int64    `json:"reason_ids" db:"reason_ids"`
}

type AudienceCreateRequest struct {
	Name   string         `json:"name" validate:"required"`
	Filter AudienceFilter `json:"filter" validate:"required"`
}

type IntegrationsCreateRequest struct {
	CabinetName string  `json:"cabinet_name"`
	AudienceIds []int64 `json:"audience_ids"`
}

package domain

import (
	"time"
	"github.com/google/uuid"
)

type Audience struct {
	ID                 uuid.UUID     `json:"id" db:"id"`
	Name               string        `json:"name" db:"name"`
	CreatedAt          time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time     `json:"last_updated" db:"last_updated"`
	CreationDateFrom   *time.Time     `json:"creation_date_from" db:"creation_date_from"`
	CreationDateTo     *time.Time     `json:"creation_date_to" db:"creation_date_to"`
	Statuses           []string      `json:"statuses" db:"statuses"`
	RejectionReasons   []string      `json:"rejection_reasons" db:"rejection_reasons"`
	NonTargetReasons   []string      `json:"non_target_reasons" db:"non_target_reasons"`
	RequestIDs         []uuid.UUID   `json:"request_ids" db:"request_ids"`
}

type AudienceFilter struct {
	IntegrationID      *uuid.UUID `json:"integration_id" db:"request_ids"`
	CreationDateFrom   *time.Time `json:"created_at" db:"created_at"`
	CreationDateTo     *time.Time `json:"last_updated" db:"last_updated"`
	Statuses           []string   `json:"statuses" db:"statuses"`
	RejectionReasons   []string   `json:"rejection_reasons" db:"rejection_reasons"`
	NonTargetReasons   []string   `json:"non_target_reasons" db:"non_target_reasons"`
}

type Integration struct {
	ID         int    `json:"id"`
	AudienceID int    `json:"audience_id"`
	Platform   string `json:"platform"` // Например: Google Ads, Facebook
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type UploadMsg struct {
	ID            int    `json:"id"`
	IntegrationID int    `json:"integration_id"`
	Title         string `json:"title"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}
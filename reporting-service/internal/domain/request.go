package domain

import (
	"time"
	//"encoding/json"
	//"github.com/google/uuid"
)

type ApplicationFilterRequest struct {
	OrderField     string 	  `json:"order_field"`
	OrderDirection string 	  `json:"order_direction"`
	Status         string 	  `json:"status_name" form:"status_name"`
	StatusDuration int  	  `json:"status_duration,omitempty" form:"status_duration"`
	ProjectName    string 	  `json:"project_name" form:"project_name"`
	PropertyType   string 	  `json:"property_type" form:"property_type"`
	AudienceName   string 	  `json:"audience_name" form:"audience_name"`
	AudienceIDs    []string   `json:"audience_ids" form:"audience_ids"`
	RegionName 	   string 	  `json:"region" form:"region"`
	StartDate      *time.Time `json:"created_at_from" form:"created_at_from"`
	EndDate        *time.Time `json:"created_at_to" form:"created_at_to"`
	DeadlinePassed bool       `json:"deadline_passed" form:"deadline_passed"`
}

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
	StartDate    		 *time.Time `json:"creation_date_from" db:"creation_date_from"`
	EndDate      		 *time.Time `json:"creation_date_to" db:"creation_date_to"`
	StatusNames          []string   `json:"statuses" db:"status_names"`
	StatusIDs            []int64    `json:"status_ids" db:"status_ids"`
	RegectionReasonNames []string   `json:"rejection_reasons" db:"rejection_reasons"`
	NonTargetReasonNames []string   `json:"non_target_reasons" db:"non_target_reasons"`
	ReasonIDs            []int64    `json:"reason_ids" db:"reason_ids"`
}

type AudienceFilter struct {}

type RegionFilter struct {
	Project   string     `json:"project"`
	Search    string     `json:"search"`
	StartDate *time.Time `json:"start_date" `
	EndDate   *time.Time `json:"end_date"`
	Sort      string     `json:"sort"`
	Status    string     `json:"status"`
}

type AudienceCreateRequest struct {
	Name   string                 `json:"name" validate:"required"`
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

type StatusDurationFilter struct {
    StartDate     *time.Time `json:"start_date"`
    EndDate       *time.Time `json:"end_date"`
    ThresholdDays int        `json:"threshold_days"`
}
package domain

import "time"

type IntegrationsCreateResponse struct {
    Integrations []Integration `json:"integrations"`
}

type AudienceResponse struct {
	ID           int64         `json:"id"`
	Name         string        `json:"name"`
	Integrations []Integration `json:"integrations"`
	Applications []Application `json:"applications"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
}
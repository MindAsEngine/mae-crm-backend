package domain

import (
	"time"
	"encoding/json"
	"github.com/google/uuid"
)

type RequestStatus string

const (
	StatusNew          RequestStatus = "new"
	StatusInProgress   RequestStatus = "in_progress"
	StatusRejected     RequestStatus = "rejected"
	StatusNonTarget    RequestStatus = "non_target"
	StatusCompleted    RequestStatus = "completed"
)

type RejectionReason string
type NonTargetReason string

var (
	RejectionReasons = map[RejectionReason]string{
		"insufficient_docs":    "Недостаточно документов",
		"credit_history_issue": "Проблемы с кредитной историей",
	}

	NonTargetReasons = map[NonTargetReason]string{
		"low_budget":      "Низкий бюджет",
		"wrong_location":  "Неподходящая локация",
	}
)

type Request struct {
	ID                 uuid.UUID        `json:"id" db:"id"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	Status             RequestStatus    `json:"status" db:"status"`
	RejectionReason    RejectionReason  `json:"rejection_reason,omitempty" db:"rejection_reason"`
	NonTargetReason    NonTargetReason  `json:"non_target_reason,omitempty" db:"non_target_reason"`
	ResponsibleUserID  uuid.UUID        `json:"responsible_user_id" db:"responsible_user_id"`
	ClientData         json.RawMessage  `json:"client_data" db:"client_data"`
}
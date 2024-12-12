package domain

import "time"

type RegionReport struct {
    Region      string `json:"region"`
    Properties  int    `json:"properties"`
    TotalAmount float64 `json:"total_amount"`
}

type SpeedReport struct {
    LeadID    int    `json:"lead_id"`
    LeadTime  string `json:"lead_time"`
    Status    string `json:"status"`
    Employee  string `json:"employee"`
}


// Фильтр для отчета по регионам
type RegionReportFilter struct {
	StartDate time.Time
	EndDate   time.Time
	Statuses  []string
}

// Строка отчета по регионам
type RegionReportRow struct {
	Region       string `db:"region"`
	TotalClients int    `db:"total_clients"`
	Purchases    int    `db:"purchases"`
}

// Фильтр для отчета по скорости обработки
type SpeedReportFilter struct {
	StartDate time.Time
	EndDate   time.Time
}

// Строка отчета по скорости обработки
type SpeedReportRow struct {
	Status          string  `db:"status"`
	AvgDurationHours float64 `db:"avg_duration_hours"`
}

// Фильтр для отчета по колл-центру
type CallCenterReportFilter struct {
	StartDate time.Time
	EndDate   time.Time
}

// Строка отчета по колл-центру
type CallCenterReportRow struct {
	ManagerName      string `db:"manager_name"`
	TotalCalls       int    `db:"total_calls"`
	TargetCalls      int    `db:"target_calls"`
	ScheduledVisits  int    `db:"scheduled_visits"`
	CompletedVisits  int    `db:"completed_visits"`
	PaidBookings     int    `db:"paid_bookings"`
	Contracts        int    `db:"contracts"`
}
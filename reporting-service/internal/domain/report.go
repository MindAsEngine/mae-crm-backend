package domain

import "time"

type SpeedReport struct {
    LeadID    int    `json:"lead_id"`
    LeadTime  string `json:"lead_time"`
    Status    string `json:"status"`
    Employee  string `json:"employee"`
}

// Фильтр для отчета по колл-центру
type CallCenterReportFilter struct {
	StartDate *time.Time  `json:"start_date"`
	EndDate   *time.Time  `json:"end_date"`
}

// Строка отчета по колл-центру
type ManagerMetrics struct {
    ManagerID            int     `json:"manager_id" db:"users_id"`
    IsAnomaly            bool    `json:"is_anomaly,omitempty"`
    ManagerName          string  `json:"manager_name" db:"users_name"`
    TotalInquiries       int     `json:"total_inquiries" db:"total_requests"`
    TargetInquiries      int     `json:"target_inquiries" db:"target_requests"`
    TargetConversion     float64 `json:"target_conversion"`
    AppointedVisits      int     `json:"appointed_visits" db:"appointed_visits"`
    VisitConversion      float64 `json:"visit_conversion"`
    CompletedVisits      int     `json:"completed_visits" db:"successful_visits"`
    VisitSuccess         float64 `json:"visit_success"`
    LeadToVisit          float64 `json:"lead_to_visit"`
    // Optional metrics
    //Bookings            int     `json:"bookings,omitempty" db:"brons"`
    //VisitToBooking      float64 `json:"visit_to_booking,omitempty"`
    //Contracts           int     `json:"contracts,omitempty" db:"ddus"`
    //BookingToContract   float64 `json:"booking_to_contract,omitempty"`
    //LeadToContract      float64 `json:"lead_to_contract,omitempty"`
}

type CallCenterReport struct {
    Headers     []Header         `json:"headers"`
    Data        []ManagerMetrics `json:"data"`
    Footer      ManagerMetrics   `json:"footer"`
    //Anomalies   []string        `json:"anomalies,omitempty"`
}
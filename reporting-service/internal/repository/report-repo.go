package repository

import (
	"context"
	"fmt"
	//"time"
    "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"reporting-service/internal/domain"
)

type ReportsRepo struct {
	Db *sqlx.DB
}

func NewReportsRepository(Db *sqlx.DB) *ReportsRepo {
	return &ReportsRepo{Db: Db}
}

// FetchRegionsReport формирует отчет по регионам
func (r *ReportsRepo) FetchRegionsReport(ctx context.Context, filter domain.RegionReportFilter) ([]domain.RegionReportRow, error) {
	query := `
		SELECT 
			region, 
			COUNT(*) AS total_clients,
			COUNT(CASE WHEN status = 'Покупка' THEN 1 END) AS purchases
		FROM estate_deals_contacts
		WHERE (created_at BETWEEN $1 AND $2) 
		  AND (status = ANY($3))
		GROUP BY region
		ORDER BY total_clients DESC;
	`

	rows := []domain.RegionReportRow{}
	err := r.Db.SelectContext(ctx, &rows, query, filter.StartDate, filter.EndDate, pq.Array(filter.Statuses))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch regions report: %w", err)
	}
	return rows, nil
}

// FetchApplicationSpeedReport формирует отчет по скорости обработки заявок
func (r *ReportsRepo) FetchApplicationSpeedReport(ctx context.Context, filter domain.SpeedReportFilter) ([]domain.SpeedReportRow, error) {
	query := `
		SELECT 
			status,
			AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 3600) AS avg_duration_hours
		FROM applications
		WHERE created_at BETWEEN $1 AND $2
		GROUP BY status
		ORDER BY avg_duration_hours DESC;
	`

	rows := []domain.SpeedReportRow{}
	err := r.Db.SelectContext(ctx, &rows, query, filter.StartDate, filter.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch application speed report: %w", err)
	}
	return rows, nil
}

// FetchCallCenterReport формирует отчет для колл-центра
func (r *ReportsRepo) FetchCallCenterReport(ctx context.Context, filter domain.CallCenterReportFilter) ([]domain.CallCenterReportRow, error) {
	query := `
		SELECT 
			manager_name,
			COUNT(*) AS total_calls,
			COUNT(CASE WHEN is_target = TRUE THEN 1 END) AS target_calls,
			COUNT(CASE WHEN is_target = TRUE AND visit_scheduled = TRUE THEN 1 END) AS scheduled_visits,
			COUNT(CASE WHEN visit_scheduled = TRUE AND visit_happened = TRUE THEN 1 END) AS completed_visits,
			COUNT(CASE WHEN completed_visits = TRUE AND booking_paid = TRUE THEN 1 END) AS paid_bookings,
			COUNT(CASE WHEN booking_paid = TRUE AND contract_signed = TRUE THEN 1 END) AS contracts
		FROM call_logs
		WHERE created_at BETWEEN $1 AND $2
		GROUP BY manager_name;
	`

	rows := []domain.CallCenterReportRow{}
	err := r.Db.SelectContext(ctx, &rows, query, filter.StartDate, filter.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch call center report: %w", err)
	}
	return rows, nil
}


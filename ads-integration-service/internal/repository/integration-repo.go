package repository

import (
    "ads-integration-service/internal/domain"
    "database/sql"
    "log"

    _ "github.com/lib/pq"
)

type IntegrationRepo struct {
    DB *sql.DB
}

func InitPostgres() *sql.DB {
    connStr := "host=db port=5432 user=postgres password=example dbname=ads_integration sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatalf("Failed to connect to PostgreSQL: %v", err)
    }
    return db
}

func (r *IntegrationRepo) GetIntegrationByID(id int) (*domain.Integration, error) {
    query := "SELECT id, audience_id, cabinet_name, status, created_at FROM integrations WHERE id = $1"
    var integration domain.Integration
    err := r.DB.QueryRow(query, id).Scan(&integration.ID, &integration.AudienceID, &integration.Platform, &integration.Status, &integration.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &integration, nil
}
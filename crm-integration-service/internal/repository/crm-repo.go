package repository

import (
    "crm-integration-service/internal/domain"
    "database/sql"
    "log"

    _ "github.com/lib/pq"
)

type CRMRepo struct {
    DB *sql.DB
}

func InitPostgres() *sql.DB {
    connStr := "host=db port=5432 user=postgres password=example dbname=crm_integration sslmode=disable" //TODO: change for env
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatalf("Failed to connect to PostgreSQL: %v", err)
    }
    return db
}

func (r *CRMRepo) CreateTask(task *domain.Task) error {
    query := "INSERT INTO tasks (lead_id, employee_id, title, status, created_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING id"
    return r.DB.QueryRow(query, task.LeadID, task.EmployeeID, task.Title, task.Status).Scan(&task.ID)
}

func (r *CRMRepo) LogTaskResult(taskID int, result string) error {
    query := "INSERT INTO task_logs (task_id, result, created_at) VALUES ($1, $2, NOW())"
    _, err := r.DB.Exec(query, taskID, result)
    return err
}

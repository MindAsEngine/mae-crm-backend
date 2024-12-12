package services

import (
    "crm-integration-service/internal/domain"
    "crm-integration-service/internal/repository"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
)

type CRMService struct {
    Repo *repository.CRMRepo
}

func NewCRMService(repo *repository.CRMRepo) *CRMService {
    return &CRMService{Repo: repo}
}

func (s *CRMService) CreateTasks(c *gin.Context) {
    var req domain.TaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    for _, leadID := range req.LeadIDs {
        task := domain.Task{
            LeadID:     leadID,
            EmployeeID: req.EmployeeID,
            Title:      "CRM Task",
            Status:     "created",
        }

        if err := s.Repo.CreateTask(&task); err != nil {
            log.Printf("Failed to create task for lead %d: %v", leadID, err)
            continue
        }

        log.Printf("Task created: %+v", task)

        // Логируем результат
        if err := s.LogTaskResult(task.ID, "Task successfully created"); err != nil {
            log.Printf("Failed to log task %d: %v", task.ID, err)
        }
    }

    c.JSON(http.StatusOK, gin.H{"status": "tasks created"})
}

func (s *CRMService) LogTaskResult(taskID int, result string) error {
    return s.Repo.LogTaskResult(taskID, result)
}
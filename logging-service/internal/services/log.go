package services

import (
    //"encoding/json"
    repository "logging-service/internal/repo"
    "net/http"

    "github.com/gin-gonic/gin"
)

type LogService struct {
    Repo *repository.LogRepo
}

func NewLogService(repo *repository.LogRepo) *LogService {
    return &LogService{Repo: repo}
}

// Получение логов
func (s *LogService) GetLogs(c *gin.Context) {
    // Пример фильтрации по уровню логов и тексту
    level := c.Query("level")
    text := c.Query("text")

    query := `{
        "query": {
            "bool": {
                "must": [
                    {"match": {"level": "` + level + `"}},
                    {"match": {"message": "` + text + `"}}
                ]
            }
        }
    }`

    logs, err := s.Repo.SearchLogs(query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, logs)
}

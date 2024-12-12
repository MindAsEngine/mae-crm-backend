package services

import (
	"net/http"
	"reporting-service/internal/domain"
	"reporting-service/internal/repository"

	"github.com/gin-gonic/gin"
)

type ReportService struct {
    Repo *repository.ReportsRepo
}

func NewReportService(repo *repository.ReportsRepo) *ReportService {
    return &ReportService{Repo: repo}
}

// Генерация отчета по регионам
func (s *ReportService) GenerateRegionReport(c *gin.Context, filter domain.RegionReportFilter) {
    data, err := s.Repo.FetchRegionsReport(c, filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}

// Генерация отчета по скорости обработки заявок
func (s *ReportService) GenerateSpeedReport(c *gin.Context, filter domain.SpeedReportFilter) {
    data, err := s.Repo.FetchApplicationSpeedReport(c, filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}

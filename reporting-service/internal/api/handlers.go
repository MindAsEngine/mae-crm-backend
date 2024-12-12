package api

import (
	//"go/format"
	"net/http"
	"reporting-service/internal/domain"
	"reporting-service/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, report_service *services.ReportService, audience_service *services.AudienceService) {
	router.GET("/reports/regions", func(c *gin.Context) {
		// Извлечение параметров запроса
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")
		statuses := c.QueryArray("statuses")
		timeFormat := time.RFC3339
		// Валидация
		if startDate == "" || endDate == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
			return
		}
		ts, errs := time.Parse(timeFormat, startDate)
		te, erre := time.Parse(timeFormat, endDate)

		if erre != nil || errs != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "start_date or end_date parsing error" +
					"\nStartDateParsingErr" + errs.Error() +
					"\nEndDateParsingErr" + erre.Error()})
		}
		// Конвертация параметров
		reportFilter := domain.RegionReportFilter{
			StartDate: ts,
			EndDate:   te,
			Statuses:  statuses,
		}

		// Передача параметров в сервис
		report_service.GenerateRegionReport(c, reportFilter)
	})

	router.GET("/reports/speed", func(c *gin.Context) {
		// Извлечение параметров фильтра
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")
		timeFormat := time.RFC3339
		// Валидация
		if startDate == "" || endDate == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
			return
		}
		ts, errs := time.Parse(timeFormat, startDate)
		te, erre := time.Parse(timeFormat, endDate)

		if erre != nil || errs != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "start_date or end_date parsing error" +
					"\nStartDateParsingErr" + errs.Error() +
					"\nEndDateParsingErr" + erre.Error()})
		}
		// Формирование параметров
		reportParams := domain.SpeedReportFilter{
			StartDate: ts,
			EndDate:   te,
		}

		// Передача параметров в сервис
		report_service.GenerateSpeedReport(c, reportParams)
	})

	// Маршруты для AudienceService
	router.POST("/audiences", audience_service.CreateAudience)
	router.GET("/audiences", audience_service.GetAudiences)
	router.GET("/audiences/:id/export", audience_service.ExportAudience)
	router.POST("/integrations", audience_service.CreateIntegration)
	router.GET("/integrations", audience_service.GetIntegrations)
}

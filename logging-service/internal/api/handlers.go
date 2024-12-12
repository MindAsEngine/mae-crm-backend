package api

import (
    "logging-service/internal/services"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, service *services.LogService) {
    router.GET("/logs", service.GetLogs)
}

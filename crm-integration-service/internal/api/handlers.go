package api

import (
    "crm-integration-service/internal/services"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, service *services.CRMService) {
    router.POST("/tasks", service.CreateTasks)
}

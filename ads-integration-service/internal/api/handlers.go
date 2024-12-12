package api

import (
    "ads-integration-service/internal/services"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, service *services.IntegrationService) {
    //router.POST("/process/{integration}", service.CreateIntegration)
}

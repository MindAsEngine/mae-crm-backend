package main

import (
    "crm-integration-service/internal/api"
    "crm-integration-service/internal/repository"
    "crm-integration-service/internal/services"
    "crm-integration-service/internal/workers"
    "log"

    "github.com/gin-gonic/gin"
)

func main() {
    // Инициализация PostgreSQL
    db := repository.InitPostgres()

    // Инициализация сервиса
    repo := repository.CRMRepo{DB: db}
    service := services.NewCRMService(&repo)

    // Создание HTTP роутеров
    router := gin.Default()
    api.RegisterRoutes(router, service)

    // Запуск логгера для фоновой записи логов
    go workers.StartLogger(service)

    // Запуск сервера
    log.Println("CRM Integration Service is running on port 8082...")
    if err := router.Run(":8082"); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    }
}

package main

import (
    "ads-integration-service/internal/api"
    "ads-integration-service/internal/repository"
    "ads-integration-service/internal/services"
    "ads-integration-service/internal/workers"
    "log"

    "github.com/gin-gonic/gin"
)

func main() {
    // Инициализация PostgreSQL
    db := repository.InitPostgres()

    // Инициализация RabbitMQ
    mq := workers.InitRabbitMQ()

    // Создание сервиса
    repo := repository.IntegrationRepo{DB: db}
    service := services.NewIntegrationService(&repo, mq)

    // Создание HTTP роутеров
    router := gin.Default()
    api.RegisterRoutes(router, service)

    // Запуск RabbitMQ worker
    go workers.StartLoader(mq, service)

    // Запуск сервера
    log.Println("Ads Integration Service is running on internal port 8081...")
    if err := router.Run(":8081"); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    }
}

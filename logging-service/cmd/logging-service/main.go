package main

import (
    "log"
    "logging-service/internal/api"
    repository "logging-service/internal/repo"
    "logging-service/internal/services"

    "github.com/gin-gonic/gin"
)

func main() {
    // Инициализация Elasticsearch
    esClient := repository.InitElasticsearch()

    // Инициализация сервиса
    repo := repository.LogRepo{Client: esClient}
    service := services.NewLogService(&repo)

    // Создание HTTP роутеров
    router := gin.Default()
    api.RegisterRoutes(router, service)

    // Запуск сервера
    log.Println("Logging Service is running on port 8084...")
    if err := router.Run(":8084"); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    }
}

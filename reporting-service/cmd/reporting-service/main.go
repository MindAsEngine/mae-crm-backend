package main

import (
	"log"
	"reporting-service/internal/middleware"
	"reporting-service/internal/api"
	"reporting-service/internal/repository"
	"reporting-service/internal/services"
	"reporting-service/internal/scheduler"
	"github.com/gin-gonic/gin"
	"context"
)

func main() {
	// Инициализация PostgreSQL
	db := repository.InitPostgres()
	cache := repository.InitRedis()
	broker := repository.InitRabbitMQ()


	// Инициализация сервиса
	report_repo := repository.ReportsRepo{Db: db}
	audience_repo := repository.AudienceRepo{DB: db}
	report_service := services.NewReportService(&report_repo)
	audience_service := services.NewAudienceService(&audience_repo, cache, broker)
	audienceUpdateService, err := scheduler.NewAudienceUpdateService(&audience_repo, broker.Channel)

    if err != nil {
        log.Fatal(err)
    }

    // Запуск планировщика задач
    ctx := context.Background()
    go audienceUpdateService.StartDailyUpdateScheduler(ctx)

    // Запуск воркера для обработки задач
    go func() {
        err := audienceUpdateService.WorkerForUpdateTasks(ctx)
        if err != nil {
            log.Fatal(err)
        }
    }()
	
	// Создание HTTP роутеров
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	api.RegisterRoutes(router, report_service, audience_service)

	// Запуск планировщика
	//scheduler.StartUpdater(audience_service, &audience_repo)


	// mux := http.NewServeMux()
	//   mux.HandleFunc("/reports/regions", func(w http.ResponseWriter, r *http.Request) {
	//     w.Header().Set("Content-Type", "application/json")
	//     w.Write([]byte(`{"data": "Пример ответа"}`))
	//   })

	//   // Настройка CORS через библиотеку
	//   handler := cors.New(cors.Options{
	//     AllowedOrigins:   []string{"http://localhost:5173"},
	//     AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
	//     AllowedHeaders:   []string{"Content-Type", "Authorization"},
	//     AllowCredentials: true,
	//   }).Handler(mux)

	// Запуск сервера
	log.Println("Audience Service is running on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

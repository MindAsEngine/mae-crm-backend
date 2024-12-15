package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"reporting-service/internal/config"
	"reporting-service/internal/api"
	
	"reporting-service/internal/services/audience"

	mysqlRepo "reporting-service/internal/repository/mysql"
	postgreRepo "reporting-service/internal/repository/postgre"
)

func main() {
    // Initialize logger
    logger, err := zap.NewProduction()
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    defer logger.Sync()

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        logger.Fatal("Failed to load config", zap.Error(err))
    }

    // Connect to MySQL
    mysqlDB, err := sqlx.Connect("mysql", cfg.MySQL.DSN)
    if err != nil {
        logger.Fatal("Failed to connect to MySQL", zap.Error(err))
    }
    defer mysqlDB.Close()

    // Connect to PostgreSQL
    postgresDB, err := sqlx.Connect("postgres", cfg.Postgres.DSN)
    if err != nil {
        logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
    }
    defer postgresDB.Close()

    // Connect to RabbitMQ
    amqpConn, err := amqp.Dial(cfg.RabbitMQ.URL)
    if err != nil {
        logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
    }
    defer amqpConn.Close()

    amqpChan, err := amqpConn.Channel()
    if err != nil {
        logger.Fatal("Failed to create RabbitMQ channel", zap.Error(err))
    }
    defer amqpChan.Close()

    // Initialize repositories
    mysqlAudienceRepo := mysqlRepo.NewMySQLAudienceRepository(mysqlDB)
    postgresAudienceRepo := postgreRepo.NewPostgresAudienceRepository(postgresDB)

    // Initialize services
    audienceService := audience.NewService(audience.Config{
        UpdateTime: cfg.Service.UpdateTime,
        BatchSize:  cfg.Service.BatchSize,
        ExportPath: cfg.Service.ExportPath,
    }, mysqlAudienceRepo, postgresAudienceRepo, logger)

    // Initialize HTTP handler
    handler := api.NewHandler(audienceService, logger)

    // Setup router with CORS
    router := mux.NewRouter()
    handler.RegisterRoutes(router)

    corsHandler := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"},
        AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    }).Handler(router)

    // Create HTTP server
    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
        Handler: corsHandler,
    }

    // Start server in goroutine
    go func() {
        logger.Info("Starting server", zap.Int("port", cfg.Server.Port))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("Server failed", zap.Error(err))
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        logger.Fatal("Server forced to shutdown", zap.Error(err))
    }

    logger.Info("Server exited properly")
}
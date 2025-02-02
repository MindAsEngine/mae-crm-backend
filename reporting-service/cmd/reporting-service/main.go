package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"reporting-service/internal/api"
	config "reporting-service/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/cors"
	"go.uber.org/zap"

	"reporting-service/internal/services/audience"

	middleware "reporting-service/internal/middleware"

	mysqlRepo "reporting-service/internal/repository/mysql"
	postgreRepo "reporting-service/internal/repository/postgre"
)

func connectWithRetry(cfg *config.Config, logger *zap.Logger) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	for i := 0; i < cfg.MySQL.MaxRetries; i++ {
		db, err = sqlx.Connect("mysql", cfg.MySQL.DSN)
		if err == nil {
			return db, nil
		}

		logger.Warn("Failed to connect to MySQL, retrying...",
			zap.Int("attempt", i+1),
			zap.Error(err))

		time.Sleep(cfg.MySQL.RetryInterval)
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w",
		cfg.MySQL.MaxRetries, err)
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	} else {
		logger.Info("Logger initialized")
	}
	defer logger.Sync()


        // If YAML fails, create config from env vars
	cfg, err := Load();

    if err!=nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	} else {
		logger.Info("Config loaded", zap.Any("config", cfg))
	}

	// Connect to MySQL
	mysqlDB, err := connectWithRetry(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to MySQL", zap.Error(err))
	} else {
		logger.Info("Connected to MySQL")
	}
	defer mysqlDB.Close()

	// Connect to PostgreSQL
	postgresDB, err := sqlx.Connect("postgres", cfg.Postgres.DSN)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	} else {
		logger.Info("Connected to PostgreSQL")
	}
	defer postgresDB.Close()

	// Connect to RabbitMQ
	amqpConn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	} else {
		logger.Info("Connected to RabbitMQ")
	}
	defer amqpConn.Close()

	amqpChan, err := amqpConn.Channel()
	if err != nil {
		logger.Fatal("Failed to create RabbitMQ channel", zap.Error(err))
	} else {
		logger.Info("Created RabbitMQ channel")
	}
	defer amqpChan.Close()

	// Initialize repositories
	mysqlAudienceRepo := mysqlRepo.NewMySQLAudienceRepository(mysqlDB, logger)
	postgresAudienceRepo := postgreRepo.NewPostgresAudienceRepository(postgresDB, logger)

	// Initialize services
	audienceService := audience.NewService(audience.Config{
		UpdateTime: cfg.Service.UpdateTime,
		BatchSize:  cfg.Service.BatchSize,
		ExportPath: cfg.Service.ExportPath,
	}, mysqlAudienceRepo, postgresAudienceRepo, amqpChan, logger)

	// Initialize HTTP handler
	handler := api.NewHandler(audienceService, logger)

	// Setup router with CORS
	router := mux.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, //[]string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Headers",
			"Access-Control-Allow-Methods",
		},
		AllowCredentials: true,
		MaxAge:           43200, //12hours
		Debug:            true,  // Enable for debugging
	})

	router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") //http://frontend:3000
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusOK)
	})

	// Apply CORS middleware
	router.Use(corsMiddleware.Handler)

	authMiddleware := middleware.NewAuthMiddleware("http://auth-service:8081")

	// Protect all routes
	// router.Use(authMiddleware.Validate)

	// Create a subrouter for protected routes
	protected := router.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.Validate)

	// Register routes that need protection on the protected subrouter
	handler.RegisterProtectedRoutes(protected)

	// Register public routes on the main router
	handler.RegisterPublicRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	updateInterval := 24 * time.Hour
	startHour := 1

	if cfg.Service.TestMode {
		updateInterval = 2 * time.Minute // Test every n minutes
		startHour = time.Now().Hour()    // Start from current hour
	}

	//Start daily update scheduler
	go func() {
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), startHour, 0, 0, 0, now.Location())
		if now.After(nextRun) {
			nextRun = nextRun.Add(updateInterval)
		}

		timer := time.NewTimer(time.Until(nextRun))
		defer timer.Stop()

		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-timer.C:
				// First run
				if err := audienceService.ProcessAllAudiences(context.Background()); err != nil {
					logger.Error("Failed to process audiences", zap.Error(err))
				}
			case <-ticker.C:
				// Subsequent runs
				if err := audienceService.ProcessAllAudiences(context.Background()); err != nil {
					logger.Error("Failed to process audiences", zap.Error(err))
				}
			}
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


func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}

func Load() (*config.Config, error) {
    return &config.Config{
        Server: config.ServerConfig{
            Port: getEnvAsInt("SERVER_PORT", 8080),
        },
        MySQL: config.MySQLConfig{
            DSN: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
                getEnvOrDefault("MYSQL_USER", "user"),
                getEnvOrDefault("MYSQL_PASSWORD", "password"),
                getEnvOrDefault("MYSQL_HOST", "localhost"),
                getEnvOrDefault("MYSQL_PORT", "3306"),
                getEnvOrDefault("MYSQL_DATABASE", "macro_bi_cmp_528"),
            ),
            MaxRetries:      getEnvAsInt("MYSQL_MAX_RETRIES", 3),
            RetryInterval:   time.Duration(getEnvAsInt("MYSQL_RETRY_INTERVAL", 15)) * time.Second,
            MaxOpenConns:    getEnvAsInt("MYSQL_MAX_OPEN_CONNS", 10),
            MaxIdleConns:    getEnvAsInt("MYSQL_MAX_IDLE_CONNS", 5),
            ConnMaxLifetime: time.Duration(getEnvAsInt("MYSQL_CONN_MAX_LIFETIME", 3600)) * time.Second,
        },
        Postgres: config.PostgresConfig{
            DSN: fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
                getEnvOrDefault("POSTGRES_USER", "postgres"),
                getEnvOrDefault("POSTGRES_PASSWORD", "password"),
                getEnvOrDefault("POSTGRES_HOST", "localhost"),
                getEnvOrDefault("POSTGRES_PORT", "5432"),
                getEnvOrDefault("POSTGRES_DB", "reports_db"),
            ),
            MaxOpenConns:    getEnvAsInt("POSTGRES_MAX_OPEN_CONNS", 10),
            MaxIdleConns:    getEnvAsInt("POSTGRES_MAX_IDLE_CONNS", 5),
            ConnMaxLifetime: time.Duration(getEnvAsInt("POSTGRES_CONN_MAX_LIFETIME", 3600)) * time.Second,
        },
        RabbitMQ: config.RabbitMQConfig{
            URL: fmt.Sprintf("amqp://%s:%s@%s:%s/",
                getEnvOrDefault("RABBITMQ_USER", "guest"),
                getEnvOrDefault("RABBITMQ_PASSWORD", "guest"),
                getEnvOrDefault("RABBITMQ_HOST", "localhost"),
                getEnvOrDefault("RABBITMQ_PORT", "5672"),
            ),
            Exchange:    getEnvOrDefault("RABBITMQ_EXCHANGE", "audience"),
            Queue:       getEnvOrDefault("RABBITMQ_QUEUE", "audience.updates"),
            RoutingKey:  getEnvOrDefault("RABBITMQ_ROUTING_KEY", "audience.update"),
        },
        Service: config.ServiceConfig{
            UpdateTime: (time.Duration(getEnvAsInt("SERVICE_UPDATE_TIME", 24)) * time.Hour).String(),
            BatchSize:  getEnvAsInt("SERVICE_BATCH_SIZE", 1000),
            ExportPath: getEnvOrDefault("SERVICE_EXPORT_PATH", "./export"),
        },
    }, nil
}
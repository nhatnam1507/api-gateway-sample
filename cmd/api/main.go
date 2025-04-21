package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway-sample/internal/application/usecase"
	"api-gateway-sample/internal/infrastructure/auth"
	"api-gateway-sample/internal/infrastructure/cache"
	"api-gateway-sample/internal/infrastructure/client"
	"api-gateway-sample/internal/infrastructure/persistence"
	"api-gateway-sample/internal/infrastructure/ratelimit"
	"api-gateway-sample/internal/infrastructure/repository"
	"api-gateway-sample/internal/interfaces/api"
	"api-gateway-sample/pkg/config"
	"api-gateway-sample/pkg/logger"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.NewZapLogger(cfg.Logging.Level, cfg.Logging.Development)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	appLogger.Info("Starting API Gateway")

	// Initialize database
	db, err := persistence.NewDatabase(cfg.Database)
	if err != nil {
		appLogger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize cache
	cacheRepo := cache.NewRedisCache(redisClient)
	cacheService := cache.NewCacheService(cacheRepo)

	// Initialize repositories
	serviceRepo := repository.NewServiceRepositoryImpl(db, appLogger)

	// Initialize HTTP client
	httpClient := client.NewHTTPClient(30*time.Second, appLogger)

	// Initialize authentication service
	authService := auth.NewJWTAuth(
		[]byte(cfg.Auth.SecretKey),
		cfg.Auth.Issuer,
		cfg.Auth.Expiration,
		appLogger,
	)

	// Initialize rate limiting service
	rateLimitService := ratelimit.NewTokenBucketRateLimiter(redisClient, appLogger)

	// Initialize gateway service
	gatewayService := client.NewGatewayService(httpClient, appLogger)

	// Initialize use cases
	proxyUseCase := usecase.NewProxyUseCase(
		serviceRepo,
		gatewayService,
		authService,
		rateLimitService,
		cacheService,
		appLogger,
	)

	authUseCase := usecase.NewAuthUseCase(authService, appLogger)
	rateLimitUseCase := usecase.NewRateLimitUseCase(rateLimitService, appLogger)
	serviceManagementUseCase := usecase.NewServiceManagementUseCase(serviceRepo, appLogger)

	// Initialize handler
	handler := api.NewHandler(
		proxyUseCase,
		authUseCase,
		rateLimitUseCase,
		serviceManagementUseCase,
		appLogger,
	)

	// Initialize router
	router := api.NewRouter(
		handler,
		appLogger,
		authUseCase,
		rateLimitUseCase,
	)

	// Initialize server
	server := api.NewServer(
		router.Setup(),
		cfg.Server.Port,
		cfg.Server.ReadTimeout,
		cfg.Server.WriteTimeout,
		cfg.Server.ShutdownTimeout,
		appLogger,
	)

	// Start server
	appLogger.Info("Server initialized", "port", cfg.Server.Port)
	if err := server.Start(); err != nil {
		appLogger.Error("Server failed", "error", err)
		os.Exit(1)
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Server shutting down")
	if err := server.Stop(); err != nil {
		appLogger.Error("Server forced to shutdown", "error", err)
	}

	appLogger.Info("Server exiting")
}

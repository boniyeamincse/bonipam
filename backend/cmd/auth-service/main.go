package main

import (
	"boni-pam/internal/app"
	"boni-pam/internal/middleware"
	"boni-pam/internal/service"
	transporthttp "boni-pam/internal/transport/http"
	"boni-pam/pkg/config"
	"boni-pam/pkg/logger"
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load("auth-service", 8081)
	if err != nil {
		panic(err)
	}

	log, err := logger.New(os.Getenv("LOG_LEVEL"))
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), middleware.CorrelationID())

	router.GET("/health", transporthttp.HealthHandler(cfg.ServiceName, cfg.Env))

	v1 := router.Group("/api/v1")
	authService, err := service.NewAuthService(service.AuthTokenConfig{
		Issuer:           os.Getenv("JWT_ISSUER"),
		SigningKey:       os.Getenv("JWT_SIGNING_KEY"),
		RequireStrongKey: true,
	})
	if err != nil {
		log.Fatal("failed to initialize auth service", zap.Error(err))
	}

	authHandler := transporthttp.NewAuthHandler(authService)
	authHandler.RegisterRoutes(v1)

	userHandler := transporthttp.NewUserHandler(service.NewUserService())
	userHandler.RegisterRoutes(v1)

	roleHandler := transporthttp.NewRoleHandler(service.NewRoleService())
	roleHandler.RegisterRoutes(v1)

	intervalSeconds := envInt("IDP_SYNC_INTERVAL_SECONDS", 300)
	if intervalSeconds > 0 {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				synced, err := authService.SyncAllUsersFromIdP()
				if err != nil {
					log.Error("periodic idp user sync failed", zap.Error(err))
					continue
				}
				log.Info("periodic idp user sync completed", zap.Int("synced_users", synced))
			}
		}()
	}

	server := app.NewServer(cfg.HTTPPort, router, log)
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start auth service", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
	}
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

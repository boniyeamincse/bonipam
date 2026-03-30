package main

import (
	"boni-pam/internal/app"
	"boni-pam/internal/middleware"
	transporthttp "boni-pam/internal/transport/http"
	"boni-pam/internal/service"
	"boni-pam/pkg/config"
	"boni-pam/pkg/logger"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load("api-gateway", 8080)
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
	gatewayHost := os.Getenv("GATEWAY_HOST")
	gatewayHandler := transporthttp.NewGatewayHandler(service.NewGatewayService(gatewayHost))
	gatewayHandler.RegisterRoutes(v1)

	server := app.NewServer(cfg.HTTPPort, router, log)
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start api gateway", zap.Error(err))
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

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smarthome/handlers"
	"smarthome/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Downstream services are configured via environment, not hardcoded ports.
	deviceURL := getEnv("DEVICE_SERVICE_URL", "http://device_service:5002")
	telemetryURL := getEnv("TELEMETRY_SERVICE_URL", "http://telemetry_service:5001")

	deviceService := services.NewDeviceService(deviceURL)
	telemetryService := services.NewTelemetryService(telemetryURL)
	log.Printf("Device service URL:    %s", deviceURL)
	log.Printf("Telemetry service URL: %s", telemetryURL)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	apiRoutes := router.Group("/api/v1")
	sensorHandler := handlers.NewSensorHandler(deviceService, telemetryService)
	sensorHandler.RegisterRoutes(apiRoutes)

	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited properly")
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
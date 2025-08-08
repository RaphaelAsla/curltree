package main

import (
	"fmt"
	"log"
	"net/http"

	"curltree/internal/config"
	"curltree/internal/database"
	"curltree/internal/handlers"
	"curltree/pkg/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger, err := utils.NewLogger(&cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	db, err := database.NewSQLiteDB(cfg.GetDatabaseURL())
	if err != nil {
		logger.LogError(err, "Failed to initialize database")
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	handler := handlers.NewHandler(db)
	rateLimiter := handlers.NewRateLimiter(
		cfg.Server.RateLimit.RequestsPerMinute,
		cfg.Server.RateLimit.Burst,
		logger,
	)
	rateLimiter.StartCleanupTask()

	loggingMiddleware := handlers.NewLoggingMiddleware(logger)

	mux := http.NewServeMux()
	
	mux.HandleFunc("/api/profiles", loggingMiddleware.Middleware(handler.CreateProfile))
	mux.HandleFunc("/api/profiles/update", loggingMiddleware.Middleware(handler.UpdateProfile))
	mux.HandleFunc("/api/profiles/delete", loggingMiddleware.Middleware(handler.DeleteProfile))
	mux.HandleFunc("/", loggingMiddleware.Middleware(rateLimiter.Middleware(handler.GetProfile)))

	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	
	logger.Info("Starting HTTP server",
		"address", serverAddr,
		"database", cfg.Database.Type,
		"rate_limit", cfg.Server.RateLimit.RequestsPerMinute,
		"rate_burst", cfg.Server.RateLimit.Burst,
	)
	
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	
	if err := server.ListenAndServe(); err != nil {
		logger.LogError(err, "Server failed to start")
		log.Fatalf("Server failed to start: %v", err)
	}
}


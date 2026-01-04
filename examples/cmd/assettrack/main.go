// AssetTrack - Enterprise Asset Management
//
// A production-ready asset management application demonstrating:
// - Chi router with middleware
// - Structured logging (slog)
// - Graceful shutdown
// - Separate API and UI routers
// - Server timeouts
//
// Usage:
//
//	go run ./cmd/assettrack
//	# or
//	go build -o assettrack ./cmd/assettrack && ./assettrack
//
// Environment:
//
//	PORT       - HTTP port (default: 31271)
//	LOG_LEVEL  - Log level: debug, info, warn, error (default: info)
//	LOG_FORMAT - Log format: text, json (default: text)
package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/ha1tch/assettrack/internal/api"
	"github.com/ha1tch/assettrack/internal/middleware"
	"github.com/ha1tch/assettrack/internal/store"
	"github.com/ha1tch/assettrack/internal/ui"
)

// Config holds application configuration.
type Config struct {
	Port      string
	LogLevel  string
	LogFormat string
}

func main() {
	// Parse configuration
	cfg := Config{
		Port:      getEnv("PORT", "31271"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "text"),
	}

	flag.StringVar(&cfg.Port, "port", cfg.Port, "HTTP port")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level (debug, info, warn, error)")
	flag.StringVar(&cfg.LogFormat, "log-format", cfg.LogFormat, "Log format (text, json)")
	flag.Parse()

	// Setup logger
	logger := setupLogger(cfg)
	logger.Info("starting AssetTrack",
		slog.String("port", cfg.Port),
		slog.String("log_level", cfg.LogLevel),
	)

	// Initialize store
	dataStore := store.NewMemoryStore()
	logger.Info("initialized in-memory store")

	// Initialize handlers
	apiHandler := api.NewHandler(dataStore, logger)
	uiHandler := ui.NewHandler(dataStore, logger)

	// Build router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recoverer(logger))
	r.Use(middleware.SecureHeaders)
	r.Use(chimw.Compress(5))

	// API routes (JSON)
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.ContentType("application/json"))
		r.Use(middleware.CORS([]string{"*"})) // Configure for production
		r.Mount("/", apiHandler.Router())
	})

	// UI routes (HTML)
	r.Mount("/", uiHandler.Router())

	// Configure server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server listening", slog.String("addr", srv.Addr))
		serverErr <- srv.ListenAndServe()
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Error("server error", slog.Any("error", err))
	case sig := <-quit:
		logger.Info("shutting down", slog.String("signal", sig.String()))
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("server stopped")
}

func setupLogger(cfg Config) *slog.Logger {
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.LogFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

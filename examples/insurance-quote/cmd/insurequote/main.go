package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/ha1tch/insurance-quote/internal/store"
	"github.com/ha1tch/insurance-quote/internal/ui"
)

func main() {
	// Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Store
	s := store.New()

	// Handler
	h := ui.NewHandler(s, logger)

	// Routes
	http.HandleFunc("/", h.Dashboard)
	http.HandleFunc("/quote", h.QuoteWizard)
	http.HandleFunc("/quotes", h.MyQuotes)
	http.HandleFunc("/claims", h.Claims)
	http.HandleFunc("/compare", h.ComparePlans)
	http.HandleFunc("/settings", h.Settings)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("starting InsureQuote", "port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Error("server error", slog.Any("error", err))
		os.Exit(1)
	}
}

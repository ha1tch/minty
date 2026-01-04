// Package api provides REST API handlers for AssetTrack.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ha1tch/assettrack/internal/models"
	"github.com/ha1tch/assettrack/internal/store"
)

// Handler holds dependencies for API handlers.
type Handler struct {
	store  store.Store
	logger *slog.Logger
}

// NewHandler creates a new API handler.
func NewHandler(s store.Store, logger *slog.Logger) *Handler {
	return &Handler{store: s, logger: logger}
}

// Router returns the API router.
func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	// Assets
	r.Route("/assets", func(r chi.Router) {
		r.Get("/", h.ListAssets)
		r.Post("/", h.CreateAsset)
		r.Get("/stats", h.GetAssetStats)
		r.Get("/{id}", h.GetAsset)
		r.Put("/{id}", h.UpdateAsset)
		r.Delete("/{id}", h.DeleteAsset)
		r.Get("/{id}/maintenance", h.GetAssetMaintenance)
	})

	// Maintenance
	r.Route("/maintenance", func(r chi.Router) {
		r.Get("/", h.ListAllMaintenance)
		r.Post("/", h.CreateMaintenance)
	})

	// Health check
	r.Get("/health", h.HealthCheck)

	return r
}

// Response helpers

type apiResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
	Meta  *meta       `json:"meta,omitempty"`
}

type meta struct {
	Total int `json:"total,omitempty"`
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiResponse{Data: data})
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiResponse{Error: message})
}

func (h *Handler) respondList(w http.ResponseWriter, data interface{}, total int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiResponse{
		Data: data,
		Meta: &meta{Total: total},
	})
}

// Handlers

// ListAssets returns all assets, optionally filtered.
// GET /api/assets?status=active&category=Laptops&search=mac
func (h *Handler) ListAssets(w http.ResponseWriter, r *http.Request) {
	filter := models.AssetFilter{
		Status:     r.URL.Query().Get("status"),
		Category:   r.URL.Query().Get("category"),
		Department: r.URL.Query().Get("department"),
		Search:     r.URL.Query().Get("search"),
	}

	assets, err := h.store.ListAssets(filter)
	if err != nil {
		h.logger.Error("failed to list assets", slog.Any("error", err))
		h.respondError(w, http.StatusInternalServerError, "Failed to list assets")
		return
	}

	h.respondList(w, assets, len(assets))
}

// GetAsset returns a single asset.
// GET /api/assets/{id}
func (h *Handler) GetAsset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	asset, err := h.store.GetAsset(id)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Asset not found")
		return
	}

	h.respondJSON(w, http.StatusOK, asset)
}

// CreateAsset creates a new asset.
// POST /api/assets
func (h *Handler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	var asset models.Asset
	if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if asset.Name == "" {
		h.respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if err := h.store.CreateAsset(&asset); err != nil {
		h.logger.Error("failed to create asset", slog.Any("error", err))
		h.respondError(w, http.StatusInternalServerError, "Failed to create asset")
		return
	}

	h.respondJSON(w, http.StatusCreated, asset)
}

// UpdateAsset updates an existing asset.
// PUT /api/assets/{id}
func (h *Handler) UpdateAsset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var asset models.Asset
	if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	asset.ID = id
	if err := h.store.UpdateAsset(&asset); err != nil {
		h.respondError(w, http.StatusNotFound, "Asset not found")
		return
	}

	h.respondJSON(w, http.StatusOK, asset)
}

// DeleteAsset removes an asset.
// DELETE /api/assets/{id}
func (h *Handler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.store.DeleteAsset(id); err != nil {
		h.respondError(w, http.StatusNotFound, "Asset not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAssetStats returns aggregate statistics.
// GET /api/assets/stats
func (h *Handler) GetAssetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetAssetStats()
	if err != nil {
		h.logger.Error("failed to get stats", slog.Any("error", err))
		h.respondError(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// GetAssetMaintenance returns maintenance records for an asset.
// GET /api/assets/{id}/maintenance
func (h *Handler) GetAssetMaintenance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	records, err := h.store.ListMaintenance(id)
	if err != nil {
		h.logger.Error("failed to list maintenance", slog.Any("error", err))
		h.respondError(w, http.StatusInternalServerError, "Failed to list maintenance")
		return
	}

	h.respondList(w, records, len(records))
}

// ListAllMaintenance returns all maintenance records.
// GET /api/maintenance
func (h *Handler) ListAllMaintenance(w http.ResponseWriter, r *http.Request) {
	records, err := h.store.ListAllMaintenance()
	if err != nil {
		h.logger.Error("failed to list maintenance", slog.Any("error", err))
		h.respondError(w, http.StatusInternalServerError, "Failed to list maintenance")
		return
	}

	h.respondList(w, records, len(records))
}

// CreateMaintenance creates a new maintenance record.
// POST /api/maintenance
func (h *Handler) CreateMaintenance(w http.ResponseWriter, r *http.Request) {
	var record models.MaintenanceRecord
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if record.AssetID == "" {
		h.respondError(w, http.StatusBadRequest, "Asset ID is required")
		return
	}

	if err := h.store.CreateMaintenance(&record); err != nil {
		h.logger.Error("failed to create maintenance", slog.Any("error", err))
		h.respondError(w, http.StatusInternalServerError, "Failed to create maintenance record")
		return
	}

	h.respondJSON(w, http.StatusCreated, record)
}

// HealthCheck returns server health status.
// GET /api/health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

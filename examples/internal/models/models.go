// Package models defines the domain types for AssetTrack.
package models

import "time"

type Asset struct {
	ID           string    `json:"id"`
	Tag          string    `json:"tag"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Status       string    `json:"status"` // active, maintenance, retired
	Location     string    `json:"location"`
	Department   string    `json:"department"`
	AssignedTo   string    `json:"assigned_to"`
	PurchaseDate string    `json:"purchase_date"`
	PurchaseCost float64   `json:"purchase_cost"`
	CurrentValue float64   `json:"current_value"`
	Vendor       string    `json:"vendor"`
	SerialNumber string    `json:"serial_number"`
	Model        string    `json:"model"`
	Warranty     string    `json:"warranty"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type MaintenanceRecord struct {
	ID          string    `json:"id"`
	AssetID     string    `json:"asset_id"`
	Date        string    `json:"date"`
	Type        string    `json:"type"` // Scheduled, Repair, Upgrade
	Description string    `json:"description"`
	Cost        float64   `json:"cost"`
	Technician  string    `json:"technician"`
	Status      string    `json:"status"` // pending, completed
	CreatedAt   time.Time `json:"created_at"`
}

type AuditEntry struct {
	ID        string    `json:"id"`
	AssetID   string    `json:"asset_id"`
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"` // admin, user, viewer
	Avatar   string `json:"avatar"`
}

// AssetFilter defines filtering options for asset queries.
type AssetFilter struct {
	Status     string
	Category   string
	Department string
	Search     string
	Limit      int
	Offset     int
}

// AssetStats holds aggregate statistics.
type AssetStats struct {
	Total          int     `json:"total"`
	Active         int     `json:"active"`
	Maintenance    int     `json:"maintenance"`
	Retired        int     `json:"retired"`
	TotalValue     float64 `json:"total_value"`
	ByCategory     map[string]int `json:"by_category"`
	ByDepartment   map[string]int `json:"by_department"`
}

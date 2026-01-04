// Package store provides data storage for AssetTrack.
// Currently implements in-memory storage; can be swapped for database later.
package store

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ha1tch/assettrack/internal/models"
)

// Store defines the data access interface.
type Store interface {
	// Assets
	ListAssets(filter models.AssetFilter) ([]models.Asset, error)
	GetAsset(id string) (*models.Asset, error)
	CreateAsset(asset *models.Asset) error
	UpdateAsset(asset *models.Asset) error
	DeleteAsset(id string) error
	GetAssetStats() (*models.AssetStats, error)

	// Maintenance
	ListMaintenance(assetID string) ([]models.MaintenanceRecord, error)
	ListAllMaintenance() ([]models.MaintenanceRecord, error)
	CreateMaintenance(record *models.MaintenanceRecord) error

	// Audit
	ListAuditEntries(assetID string) ([]models.AuditEntry, error)
	CreateAuditEntry(entry *models.AuditEntry) error
}

// MemoryStore implements Store with in-memory storage.
type MemoryStore struct {
	mu          sync.RWMutex
	assets      map[string]models.Asset
	maintenance map[string][]models.MaintenanceRecord
	audit       map[string][]models.AuditEntry
	nextID      int
}

// NewMemoryStore creates a new in-memory store with sample data.
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		assets:      make(map[string]models.Asset),
		maintenance: make(map[string][]models.MaintenanceRecord),
		audit:       make(map[string][]models.AuditEntry),
		nextID:      100,
	}
	s.loadSampleData()
	return s
}

func (s *MemoryStore) loadSampleData() {
	assets := []models.Asset{
		{ID: "A001", Tag: "IT-LAP-001", Name: "MacBook Pro 16\"", Category: "Laptops", Status: "active", Location: "HQ Floor 3", Department: "Engineering", AssignedTo: "John Smith", PurchaseDate: "2024-01-15", PurchaseCost: 2499.00, CurrentValue: 2100.00, Vendor: "Apple Inc.", SerialNumber: "C02XG123HKGY", Model: "MacBook Pro 16 M3", Warranty: "2027-01-15", Notes: "Primary development machine"},
		{ID: "A002", Tag: "IT-LAP-002", Name: "ThinkPad X1 Carbon", Category: "Laptops", Status: "active", Location: "HQ Floor 2", Department: "Sales", AssignedTo: "Jane Doe", PurchaseDate: "2024-02-20", PurchaseCost: 1899.00, CurrentValue: 1650.00, Vendor: "Lenovo", SerialNumber: "PF3ABCD1", Model: "X1 Carbon Gen 11", Warranty: "2027-02-20", Notes: ""},
		{ID: "A003", Tag: "IT-MON-001", Name: "Dell U2723QE", Category: "Monitors", Status: "active", Location: "HQ Floor 3", Department: "Engineering", AssignedTo: "John Smith", PurchaseDate: "2024-01-15", PurchaseCost: 799.00, CurrentValue: 700.00, Vendor: "Dell", SerialNumber: "CN0M2K831234", Model: "UltraSharp 27 4K", Warranty: "2027-01-15", Notes: "4K USB-C Hub Monitor"},
		{ID: "A004", Tag: "IT-SRV-001", Name: "Dell PowerEdge R750", Category: "Servers", Status: "active", Location: "Data Center", Department: "IT Operations", AssignedTo: "Unassigned", PurchaseDate: "2023-06-01", PurchaseCost: 12500.00, CurrentValue: 10000.00, Vendor: "Dell", SerialNumber: "SVCTAG001", Model: "PowerEdge R750xs", Warranty: "2026-06-01", Notes: "Primary application server"},
		{ID: "A005", Tag: "IT-LAP-003", Name: "MacBook Air M2", Category: "Laptops", Status: "maintenance", Location: "IT Storage", Department: "Marketing", AssignedTo: "Bob Wilson", PurchaseDate: "2024-03-10", PurchaseCost: 1299.00, CurrentValue: 1150.00, Vendor: "Apple Inc.", SerialNumber: "C02YH456JKLY", Model: "MacBook Air 13 M2", Warranty: "2027-03-10", Notes: "Battery replacement scheduled"},
		{ID: "A006", Tag: "IT-NET-001", Name: "Cisco Catalyst 9300", Category: "Network", Status: "active", Location: "Server Room A", Department: "IT Operations", AssignedTo: "Unassigned", PurchaseDate: "2023-01-20", PurchaseCost: 4500.00, CurrentValue: 3800.00, Vendor: "Cisco", SerialNumber: "FCW2345K001", Model: "C9300-48P", Warranty: "2026-01-20", Notes: "48-port PoE+ switch"},
		{ID: "A007", Tag: "IT-PRN-001", Name: "HP LaserJet Pro", Category: "Printers", Status: "active", Location: "HQ Floor 2", Department: "Shared", AssignedTo: "Unassigned", PurchaseDate: "2024-04-05", PurchaseCost: 549.00, CurrentValue: 500.00, Vendor: "HP", SerialNumber: "VNB3R12345", Model: "M404dn", Warranty: "2025-04-05", Notes: "Department printer"},
		{ID: "A008", Tag: "IT-LAP-004", Name: "Dell XPS 15", Category: "Laptops", Status: "retired", Location: "IT Storage", Department: "Finance", AssignedTo: "Unassigned", PurchaseDate: "2021-05-15", PurchaseCost: 1799.00, CurrentValue: 0.00, Vendor: "Dell", SerialNumber: "5CG123ABC", Model: "XPS 15 9510", Warranty: "2024-05-15", Notes: "End of life - data wiped"},
		{ID: "A009", Tag: "IT-MON-002", Name: "LG 34WN80C-B", Category: "Monitors", Status: "active", Location: "HQ Floor 2", Department: "Sales", AssignedTo: "Jane Doe", PurchaseDate: "2024-02-20", PurchaseCost: 699.00, CurrentValue: 600.00, Vendor: "LG", SerialNumber: "LG34WN001", Model: "34WN80C-B", Warranty: "2027-02-20", Notes: "Ultrawide monitor"},
		{ID: "A010", Tag: "IT-LAP-005", Name: "HP EliteBook 840", Category: "Laptops", Status: "active", Location: "HQ Floor 1", Department: "HR", AssignedTo: "Alice Brown", PurchaseDate: "2024-05-01", PurchaseCost: 1499.00, CurrentValue: 1400.00, Vendor: "HP", SerialNumber: "5CG456DEF", Model: "EliteBook 840 G9", Warranty: "2027-05-01", Notes: ""},
	}

	for _, a := range assets {
		a.CreatedAt = time.Now()
		a.UpdatedAt = time.Now()
		s.assets[a.ID] = a
	}

	// Sample maintenance records
	s.maintenance["A001"] = []models.MaintenanceRecord{
		{ID: "M001", AssetID: "A001", Date: "2025-01-02", Type: "Scheduled", Description: "Annual checkup and cleaning", Cost: 75.00, Status: "completed"},
		{ID: "M002", AssetID: "A001", Date: "2024-09-15", Type: "Repair", Description: "Battery replacement", Cost: 199.00, Status: "completed"},
		{ID: "M003", AssetID: "A001", Date: "2024-06-20", Type: "Upgrade", Description: "RAM upgrade to 32GB", Cost: 350.00, Status: "completed"},
	}
	s.maintenance["A005"] = []models.MaintenanceRecord{
		{ID: "M004", AssetID: "A005", Date: "2025-01-03", Type: "Repair", Description: "Battery replacement", Cost: 189.00, Status: "pending"},
		{ID: "M005", AssetID: "A005", Date: "2024-08-10", Type: "Scheduled", Description: "Annual checkup", Cost: 50.00, Status: "completed"},
	}
}

// ListAssets returns assets matching the filter.
func (s *MemoryStore) ListAssets(filter models.AssetFilter) ([]models.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.Asset
	for _, a := range s.assets {
		if filter.Status != "" && a.Status != filter.Status {
			continue
		}
		if filter.Category != "" && a.Category != filter.Category {
			continue
		}
		if filter.Department != "" && a.Department != filter.Department {
			continue
		}
		if filter.Search != "" && !strings.Contains(strings.ToLower(a.Name), strings.ToLower(filter.Search)) {
			continue
		}
		result = append(result, a)
	}
	return result, nil
}

// GetAsset returns a single asset by ID.
func (s *MemoryStore) GetAsset(id string) (*models.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	asset, ok := s.assets[id]
	if !ok {
		return nil, fmt.Errorf("asset not found: %s", id)
	}
	return &asset, nil
}

// CreateAsset adds a new asset.
func (s *MemoryStore) CreateAsset(asset *models.Asset) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if asset.ID == "" {
		s.nextID++
		asset.ID = fmt.Sprintf("A%03d", s.nextID)
	}
	asset.CreatedAt = time.Now()
	asset.UpdatedAt = time.Now()
	s.assets[asset.ID] = *asset
	return nil
}

// UpdateAsset updates an existing asset.
func (s *MemoryStore) UpdateAsset(asset *models.Asset) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.assets[asset.ID]; !ok {
		return fmt.Errorf("asset not found: %s", asset.ID)
	}
	asset.UpdatedAt = time.Now()
	s.assets[asset.ID] = *asset
	return nil
}

// DeleteAsset removes an asset.
func (s *MemoryStore) DeleteAsset(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.assets[id]; !ok {
		return fmt.Errorf("asset not found: %s", id)
	}
	delete(s.assets, id)
	return nil
}

// GetAssetStats returns aggregate statistics.
func (s *MemoryStore) GetAssetStats() (*models.AssetStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &models.AssetStats{
		ByCategory:   make(map[string]int),
		ByDepartment: make(map[string]int),
	}

	for _, a := range s.assets {
		stats.Total++
		stats.TotalValue += a.CurrentValue
		stats.ByCategory[a.Category]++
		stats.ByDepartment[a.Department]++

		switch a.Status {
		case "active":
			stats.Active++
		case "maintenance":
			stats.Maintenance++
		case "retired":
			stats.Retired++
		}
	}
	return stats, nil
}

// ListMaintenance returns maintenance records for an asset.
func (s *MemoryStore) ListMaintenance(assetID string) ([]models.MaintenanceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.maintenance[assetID], nil
}

// ListAllMaintenance returns all maintenance records.
func (s *MemoryStore) ListAllMaintenance() ([]models.MaintenanceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.MaintenanceRecord
	for _, records := range s.maintenance {
		result = append(result, records...)
	}
	return result, nil
}

// CreateMaintenance adds a new maintenance record.
func (s *MemoryStore) CreateMaintenance(record *models.MaintenanceRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	record.ID = fmt.Sprintf("M%03d", s.nextID)
	record.CreatedAt = time.Now()
	s.maintenance[record.AssetID] = append(s.maintenance[record.AssetID], *record)
	return nil
}

// ListAuditEntries returns audit entries for an asset.
func (s *MemoryStore) ListAuditEntries(assetID string) ([]models.AuditEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.audit[assetID], nil
}

// CreateAuditEntry adds an audit entry.
func (s *MemoryStore) CreateAuditEntry(entry *models.AuditEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	entry.ID = fmt.Sprintf("AU%03d", s.nextID)
	entry.Timestamp = time.Now()
	s.audit[entry.AssetID] = append(s.audit[entry.AssetID], *entry)
	return nil
}

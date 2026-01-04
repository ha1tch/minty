package store

import (
	"time"

	"github.com/ha1tch/insurance-quote/internal/models"
)

// Store provides in-memory data storage.
type Store struct {
	Coverages []models.Coverage
	Plans     []models.Plan
	Quotes    []models.Quote
	Claims    []models.Claim
}

// New creates a new store with sample data.
func New() *Store {
	s := &Store{}
	s.initCoverages()
	s.initPlans()
	s.initQuotes()
	s.initClaims()
	return s
}

func (s *Store) initCoverages() {
	s.Coverages = []models.Coverage{
		{
			ID:          "auto",
			Type:        models.CoverageAuto,
			Name:        "Auto Insurance",
			Description: "Comprehensive coverage for your vehicles including collision, liability, and roadside assistance.",
			BasePrice:   89.99,
			Icon:        "truck",
		},
		{
			ID:          "home",
			Type:        models.CoverageHome,
			Name:        "Home Insurance",
			Description: "Protect your home and belongings from damage, theft, and natural disasters.",
			BasePrice:   125.00,
			Icon:        "home-modern",
		},
		{
			ID:          "life",
			Type:        models.CoverageLife,
			Name:        "Life Insurance",
			Description: "Financial security for your loved ones with flexible term and whole life options.",
			BasePrice:   45.00,
			Icon:        "heart",
		},
		{
			ID:          "business",
			Type:        models.CoverageBusiness,
			Name:        "Business Insurance",
			Description: "Complete protection for your business including liability, property, and workers' compensation.",
			BasePrice:   299.00,
			Icon:        "building-office",
		},
	}
}

func (s *Store) initPlans() {
	s.Plans = []models.Plan{
		// Auto plans
		{ID: "auto-basic", CoverageType: "auto", Name: "Basic Auto", Tier: "basic", Price: 59.99, Deductible: 1000, Coverage: 50000, Features: []string{"Liability coverage", "Uninsured motorist", "24/7 claims"}},
		{ID: "auto-standard", CoverageType: "auto", Name: "Standard Auto", Tier: "standard", Price: 89.99, Deductible: 500, Coverage: 100000, Features: []string{"Liability coverage", "Collision coverage", "Uninsured motorist", "Rental car", "24/7 claims"}, Popular: true},
		{ID: "auto-premium", CoverageType: "auto", Name: "Premium Auto", Tier: "premium", Price: 149.99, Deductible: 250, Coverage: 250000, Features: []string{"Full coverage", "Gap coverage", "Roadside assistance", "New car replacement", "Accident forgiveness", "24/7 concierge"}},
		
		// Home plans
		{ID: "home-basic", CoverageType: "home", Name: "Basic Home", Tier: "basic", Price: 85.00, Deductible: 2500, Coverage: 150000, Features: []string{"Dwelling coverage", "Personal property", "Liability protection"}},
		{ID: "home-standard", CoverageType: "home", Name: "Standard Home", Tier: "standard", Price: 125.00, Deductible: 1000, Coverage: 300000, Features: []string{"Dwelling coverage", "Personal property", "Liability protection", "Additional living expenses", "Medical payments"}, Popular: true},
		{ID: "home-premium", CoverageType: "home", Name: "Premium Home", Tier: "premium", Price: 225.00, Deductible: 500, Coverage: 500000, Features: []string{"Full replacement cost", "Extended coverage", "Identity theft", "Home business", "Equipment breakdown", "Water backup"}},
		
		// Life plans
		{ID: "life-basic", CoverageType: "life", Name: "Term Life 10", Tier: "basic", Price: 25.00, Deductible: 0, Coverage: 100000, Features: []string{"10-year term", "Level premiums", "Convertible"}},
		{ID: "life-standard", CoverageType: "life", Name: "Term Life 20", Tier: "standard", Price: 45.00, Deductible: 0, Coverage: 250000, Features: []string{"20-year term", "Level premiums", "Convertible", "Accelerated death benefit"}, Popular: true},
		{ID: "life-premium", CoverageType: "life", Name: "Whole Life", Tier: "premium", Price: 125.00, Deductible: 0, Coverage: 500000, Features: []string{"Lifetime coverage", "Cash value growth", "Dividends", "Loan options", "Estate planning"}},
		
		// Business plans
		{ID: "biz-basic", CoverageType: "business", Name: "Starter Business", Tier: "basic", Price: 199.00, Deductible: 2500, Coverage: 100000, Features: []string{"General liability", "Property coverage", "Business interruption"}},
		{ID: "biz-standard", CoverageType: "business", Name: "Growing Business", Tier: "standard", Price: 349.00, Deductible: 1000, Coverage: 500000, Features: []string{"General liability", "Property coverage", "Business interruption", "Professional liability", "Cyber liability"}, Popular: true},
		{ID: "biz-premium", CoverageType: "business", Name: "Enterprise", Tier: "premium", Price: 599.00, Deductible: 500, Coverage: 1000000, Features: []string{"Comprehensive liability", "Property coverage", "Business interruption", "Directors & officers", "Employment practices", "Cyber liability", "Umbrella coverage"}},
	}
}

func (s *Store) initQuotes() {
	now := time.Now()
	s.Quotes = []models.Quote{
		{ID: "Q001", CustomerName: "John Smith", Email: "john@example.com", Phone: "555-0101", CoverageType: "auto", PlanID: "auto-standard", Status: models.QuoteStatusApproved, Premium: 89.99, CreatedAt: now.AddDate(0, 0, -5), ExpiresAt: now.AddDate(0, 0, 25)},
		{ID: "Q002", CustomerName: "Sarah Johnson", Email: "sarah@example.com", Phone: "555-0102", CoverageType: "home", PlanID: "home-premium", Status: models.QuoteStatusPending, Premium: 225.00, CreatedAt: now.AddDate(0, 0, -2), ExpiresAt: now.AddDate(0, 0, 28)},
		{ID: "Q003", CustomerName: "Mike Williams", Email: "mike@example.com", Phone: "555-0103", CoverageType: "life", PlanID: "life-standard", Status: models.QuoteStatusDraft, Premium: 45.00, CreatedAt: now.AddDate(0, 0, -1), ExpiresAt: now.AddDate(0, 0, 29)},
		{ID: "Q004", CustomerName: "Emily Brown", Email: "emily@example.com", Phone: "555-0104", CoverageType: "auto", PlanID: "auto-premium", Status: models.QuoteStatusApproved, Premium: 149.99, CreatedAt: now.AddDate(0, 0, -10), ExpiresAt: now.AddDate(0, 0, 20)},
		{ID: "Q005", CustomerName: "David Lee", Email: "david@example.com", Phone: "555-0105", CoverageType: "business", PlanID: "biz-standard", Status: models.QuoteStatusDeclined, Premium: 349.00, CreatedAt: now.AddDate(0, 0, -15), ExpiresAt: now.AddDate(0, 0, 15)},
		{ID: "Q006", CustomerName: "Lisa Chen", Email: "lisa@example.com", Phone: "555-0106", CoverageType: "home", PlanID: "home-standard", Status: models.QuoteStatusPending, Premium: 125.00, CreatedAt: now.AddDate(0, 0, -3), ExpiresAt: now.AddDate(0, 0, 27)},
		{ID: "Q007", CustomerName: "Tom Anderson", Email: "tom@example.com", Phone: "555-0107", CoverageType: "auto", PlanID: "auto-basic", Status: models.QuoteStatusExpired, Premium: 59.99, CreatedAt: now.AddDate(0, 0, -45), ExpiresAt: now.AddDate(0, 0, -15)},
		{ID: "Q008", CustomerName: "Amy Wilson", Email: "amy@example.com", Phone: "555-0108", CoverageType: "life", PlanID: "life-premium", Status: models.QuoteStatusApproved, Premium: 125.00, CreatedAt: now.AddDate(0, 0, -7), ExpiresAt: now.AddDate(0, 0, 23)},
	}
}

func (s *Store) initClaims() {
	s.Claims = []models.Claim{
		{ID: "CLM001", PolicyNumber: "POL-2024-001", CustomerName: "John Smith", Type: "collision", Status: models.ClaimStatusApproved, Amount: 3500.00, Filed: "2024-11-15", Description: "Rear-end collision at intersection"},
		{ID: "CLM002", PolicyNumber: "POL-2024-002", CustomerName: "Sarah Johnson", Type: "water", Status: models.ClaimStatusInProgress, Amount: 12000.00, Filed: "2024-12-01", Description: "Pipe burst causing water damage to kitchen"},
		{ID: "CLM003", PolicyNumber: "POL-2024-003", CustomerName: "Mike Williams", Type: "theft", Status: models.ClaimStatusOpen, Amount: 2500.00, Filed: "2024-12-10", Description: "Laptop and equipment stolen from vehicle"},
		{ID: "CLM004", PolicyNumber: "POL-2024-001", CustomerName: "John Smith", Type: "glass", Status: models.ClaimStatusClosed, Amount: 450.00, Filed: "2024-09-20", Description: "Windshield replacement"},
		{ID: "CLM005", PolicyNumber: "POL-2024-004", CustomerName: "Emily Brown", Type: "collision", Status: models.ClaimStatusDenied, Amount: 8000.00, Filed: "2024-10-05", Description: "Single vehicle accident - policy lapsed"},
		{ID: "CLM006", PolicyNumber: "POL-2024-005", CustomerName: "David Lee", Type: "liability", Status: models.ClaimStatusInProgress, Amount: 25000.00, Filed: "2024-11-28", Description: "Customer slip and fall at business premises"},
		{ID: "CLM007", PolicyNumber: "POL-2024-002", CustomerName: "Sarah Johnson", Type: "fire", Status: models.ClaimStatusOpen, Amount: 45000.00, Filed: "2024-12-15", Description: "Kitchen fire damage"},
		{ID: "CLM008", PolicyNumber: "POL-2024-006", CustomerName: "Lisa Chen", Type: "weather", Status: models.ClaimStatusApproved, Amount: 7500.00, Filed: "2024-11-10", Description: "Storm damage to roof and fence"},
		{ID: "CLM009", PolicyNumber: "POL-2024-007", CustomerName: "Tom Anderson", Type: "collision", Status: models.ClaimStatusClosed, Amount: 4200.00, Filed: "2024-08-15", Description: "Parking lot collision"},
		{ID: "CLM010", PolicyNumber: "POL-2024-008", CustomerName: "Amy Wilson", Type: "medical", Status: models.ClaimStatusApproved, Amount: 15000.00, Filed: "2024-10-20", Description: "Medical expenses claim"},
	}
}

// GetCoverage returns a coverage by ID.
func (s *Store) GetCoverage(id string) *models.Coverage {
	for _, c := range s.Coverages {
		if c.ID == id {
			return &c
		}
	}
	return nil
}

// GetPlansByType returns all plans for a coverage type.
func (s *Store) GetPlansByType(coverageType string) []models.Plan {
	var plans []models.Plan
	for _, p := range s.Plans {
		if p.CoverageType == coverageType {
			plans = append(plans, p)
		}
	}
	return plans
}

// GetPlan returns a plan by ID.
func (s *Store) GetPlan(id string) *models.Plan {
	for _, p := range s.Plans {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// ClaimsAsMapSlice returns claims as []map[string]interface{} for mintydyn.
func (s *Store) ClaimsAsMapSlice() []map[string]interface{} {
	result := make([]map[string]interface{}, len(s.Claims))
	for i, c := range s.Claims {
		result[i] = map[string]interface{}{
			"id":           c.ID,
			"policyNumber": c.PolicyNumber,
			"customerName": c.CustomerName,
			"type":         c.Type,
			"status":       string(c.Status),
			"amount":       c.Amount,
			"filed":        c.Filed,
			"description":  c.Description,
		}
	}
	return result
}

// PlansAsMapSlice returns plans as []map[string]interface{} for mintydyn.
func (s *Store) PlansAsMapSlice() []map[string]interface{} {
	result := make([]map[string]interface{}, len(s.Plans))
	for i, p := range s.Plans {
		result[i] = map[string]interface{}{
			"id":           p.ID,
			"coverageType": p.CoverageType,
			"name":         p.Name,
			"tier":         p.Tier,
			"price":        p.Price,
			"deductible":   p.Deductible,
			"coverage":     p.Coverage,
			"popular":      p.Popular,
		}
	}
	return result
}

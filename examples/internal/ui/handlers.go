// Package ui provides the web UI for AssetTrack using minty.
package ui

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	mi "github.com/ha1tch/minty"
	mdy "github.com/ha1tch/minty/mintydyn"
	"github.com/ha1tch/assettrack/internal/models"
	"github.com/ha1tch/assettrack/internal/store"
)

// Handler holds dependencies for UI handlers.
type Handler struct {
	store  store.Store
	logger *slog.Logger
	theme  mdy.DynamicTheme
}

// NewHandler creates a new UI handler.
func NewHandler(s store.Store, logger *slog.Logger) *Handler {
	return &Handler{
		store:  s,
		logger: logger,
		theme:  mdy.NewTailwindDarkTheme(),
	}
}

// Router returns the UI router.
func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.Dashboard)
	r.Get("/assets", h.AssetList)
	r.Get("/assets/new", h.AssetNew)
	r.Post("/assets/new", h.AssetCreate)
	r.Get("/assets/{id}", h.AssetDetail)
	r.Post("/assets/{id}", h.AssetUpdate)
	r.Get("/maintenance", h.Maintenance)
	r.Get("/reports", h.Reports)
	r.Get("/settings", h.Settings)
	r.Post("/settings", h.SettingsSave)

	return r
}

// render converts a minty.H to HTTP response.
func (h *Handler) render(w http.ResponseWriter, page mi.H) {
	var buf bytes.Buffer
	if err := mi.Render(page, &buf); err != nil {
		h.logger.Error("render failed", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

// =============================================================================
// PAGE HANDLERS
// =============================================================================

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetAssetStats()
	if err != nil {
		h.logger.Error("failed to get stats", slog.Any("error", err))
		stats = &models.AssetStats{}
	}

	page := h.pageLayout("dashboard", "Dashboard", "Overview of your asset portfolio", func(b *mi.Builder) mi.Node {
		return b.Div(mi.Class("space-y-6"),
			// Stats cards
			b.Div(mi.Class("grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4"),
				statCard(b, "Total Assets", fmt.Sprintf("%d", stats.Total), "+2 this month", true, "assets"),
				statCard(b, "Active", fmt.Sprintf("%d", stats.Active), "92% of total", true, "check"),
				statCard(b, "Maintenance", fmt.Sprintf("%d", stats.Maintenance), "-1 from last week", true, "maintenance"),
				statCard(b, "Total Value", fmt.Sprintf("$%.0fK", stats.TotalValue/1000), "+5% this quarter", true, "dashboard"),
			),
			// Category breakdown
			b.Div(mi.Class("grid grid-cols-1 lg:grid-cols-3 gap-6"),
				b.Div(mi.Class("lg:col-span-2 bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4"),
					b.H3(mi.Class("text-lg font-medium text-gray-900 dark:text-white mb-4"), "Assets by Category"),
					b.Div(mi.Class("space-y-3"),
						categoryBar(b, "Laptops", stats.ByCategory["Laptops"], 50),
						categoryBar(b, "Monitors", stats.ByCategory["Monitors"], 20),
						categoryBar(b, "Servers", stats.ByCategory["Servers"], 10),
						categoryBar(b, "Network", stats.ByCategory["Network"], 10),
						categoryBar(b, "Printers", stats.ByCategory["Printers"], 10),
					),
				),
				b.Div(mi.Class("bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4"),
					b.H3(mi.Class("text-lg font-medium text-gray-900 dark:text-white mb-4"), "Recent Activity"),
					b.Div(mi.Class("space-y-3"),
						activityItem(b, "MacBook Pro", "Battery replaced", "2 hours ago"),
						activityItem(b, "Dell Server", "Scheduled maintenance", "Yesterday"),
						activityItem(b, "HP Printer", "Toner replaced", "2 days ago"),
						activityItem(b, "ThinkPad X1", "Assigned to Jane", "3 days ago"),
					),
				),
			),
		)
	})

	h.render(w, page)
}

func (h *Handler) AssetList(w http.ResponseWriter, r *http.Request) {
	assets, err := h.store.ListAssets(models.AssetFilter{})
	if err != nil {
		h.logger.Error("failed to list assets", slog.Any("error", err))
		assets = []models.Asset{}
	}

	page := h.pageLayout("assets", "Asset Inventory", "Manage and track all company assets", func(b *mi.Builder) mi.Node {
		// Combined filter component using mintydyn
		// - ServerRenderedData mode filters pre-rendered table rows
		// - TextFilter for search, SelectFilter for status
		// - No hand-written JavaScript needed!
		assetFilter := mdy.Dyn("asset-filter").
			ServerRenderedData(".asset-row", "#asset-count").
			TextFilter("name", "Search").
			SelectFilter("status", "Status", []string{"active", "maintenance", "retired"}).
			Theme(h.theme).
			Minified().
			Build()

		return b.Div(
			// Toolbar with Add button and search (search connected to filter)
			b.Div(mi.Class("flex items-center justify-between mb-4"),
				b.Div(mi.Class("flex items-center gap-2"),
					b.A(mi.Href("/assets/new"), mi.Class("inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"),
						icon("add")(b), "Add Asset",
					),
					b.Button(mi.Class("inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-500 dark:text-gray-400 bg-transparent border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"), mi.Type("button"),
						icon("export")(b), "Export",
					),
				),
				// Search input is now empty - filter controls are generated by mintydyn
			),
			// Filter component (generates tabs/controls and handles filtering)
			b.Div(mi.Class("bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 mb-4"),
				b.Div(mi.Class("px-4 py-4"), assetFilter(b)),
				b.Div(mi.Class("px-4 pb-2 text-sm text-gray-500 dark:text-gray-400"),
					b.Span(mi.ID("asset-count"), fmt.Sprintf("Showing %d assets", len(assets))),
				),
			),
			// Asset table (rows have data-* attributes for filtering)
			b.Div(mi.Class("bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden"),
				h.assetTable(b, assets),
			),
			// No filter script needed - mintydyn generates it!
		)
	})

	h.render(w, page)
}

func (h *Handler) AssetDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	asset, err := h.store.GetAsset(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	records, _ := h.store.ListMaintenance(id)

	page := h.pageLayout("assets", "Asset: "+asset.Name, asset.Tag+" • "+asset.Category, func(b *mi.Builder) mi.Node {
		states := h.buildAssetDetailStates(b, asset, records)

		detailTabs := mdy.Dyn("asset-detail-tabs").
			States(states).
			Theme(h.theme).
			Minified().
			Build()

		return b.Div(
			// Breadcrumb
			b.Div(mi.Class("flex items-center gap-2 text-sm text-gray-500 mb-4"),
				b.A(mi.Href("/assets"), mi.Class("hover:text-gray-700 dark:hover:text-gray-200"), "Assets"),
				b.Span("›"),
				b.Span(mi.Class("text-gray-900 dark:text-white"), asset.Name),
			),
			// Main card
			b.Form(mi.Class("bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700"), mi.Method("POST"), mi.Action("/assets/"+asset.ID),
				detailTabs(b),
				// Actions
				b.Div(mi.Class("flex items-center justify-between px-6 py-4 bg-gray-50 dark:bg-gray-900/50 border-t border-gray-200 dark:border-gray-700"),
					b.A(mi.Href("/assets"), mi.Class("px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600"), "Cancel"),
					b.Button(mi.Class("px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"), mi.Type("submit"), "Save Changes"),
				),
			),
		)
	})

	h.render(w, page)
}

func (h *Handler) Maintenance(w http.ResponseWriter, r *http.Request) {
	assets, _ := h.store.ListAssets(models.AssetFilter{})
	
	type recordWithAsset struct {
		AssetID   string
		AssetName string
		Record    models.MaintenanceRecord
	}

	var allRecords []recordWithAsset
	for _, asset := range assets {
		records, _ := h.store.ListMaintenance(asset.ID)
		for _, r := range records {
			allRecords = append(allRecords, recordWithAsset{
				AssetID:   asset.ID,
				AssetName: asset.Name,
				Record:    r,
			})
		}
	}

	page := h.pageLayout("maintenance", "Maintenance", "Track and schedule asset maintenance", func(b *mi.Builder) mi.Node {
		// Use mintydyn with server-rendered filtering
		// No hand-written JavaScript needed!
		maintFilter := mdy.Dyn("maint-filter").
			ServerRenderedData(".maint-row", "").
			SelectFilter("status", "Status", []string{"pending", "completed"}).
			Theme(h.theme).
			Minified().
			Build()

		rows := make([]mi.Node, len(allRecords))
		for i, item := range allRecords {
			rows[i] = b.Tr(mi.Class("hover:bg-gray-50 dark:hover:bg-gray-700 maint-row"), mi.Data("status", item.Record.Status),
				b.Td(mi.Class("px-4 py-3"),
					b.A(mi.Href("/assets/"+item.AssetID), mi.Class("text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300"), item.AssetName),
				),
				b.Td(mi.Class("px-4 py-3 text-sm text-gray-900 dark:text-gray-100"), item.Record.Date),
				b.Td(mi.Class("px-4 py-3"), b.Span(mi.Class("px-2 py-0.5 text-xs rounded border bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800"), item.Record.Type)),
				b.Td(mi.Class("px-4 py-3 text-sm text-gray-600 dark:text-gray-400"), item.Record.Description),
				b.Td(mi.Class("px-4 py-3 text-sm text-gray-900 dark:text-gray-100"), fmt.Sprintf("$%.2f", item.Record.Cost)),
				b.Td(mi.Class("px-4 py-3"), statusBadge(b, item.Record.Status)),
			)
		}

		return b.Div(
			b.Div(mi.Class("bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 mb-4 p-4"),
				maintFilter(b),
			),
			b.Div(mi.Class("bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden"),
				b.Table(mi.Class("w-full"),
					b.Thead(mi.Class("bg-gray-50 dark:bg-gray-900/50 border-b border-gray-200 dark:border-gray-700"),
						b.Tr(
							b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Asset"),
							b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Date"),
							b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Type"),
							b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Description"),
							b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Cost"),
							b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Status"),
						),
					),
					b.Tbody(mi.Class("divide-y divide-gray-200 dark:divide-gray-700"), mi.NewFragment(rows...)),
				),
			),
		)
	})

	h.render(w, page)
}

func (h *Handler) Reports(w http.ResponseWriter, r *http.Request) {
	page := h.pageLayout("reports", "Reports", "Generate and view asset reports", func(b *mi.Builder) mi.Node {
		return b.Div(mi.Class("grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"),
			reportCard(b, "Asset Inventory", "Complete list of all assets", "📋"),
			reportCard(b, "Depreciation Report", "Asset value over time", "📉"),
			reportCard(b, "Maintenance Summary", "Service history and costs", "🔧"),
			reportCard(b, "Department Assets", "Assets by department", "🏢"),
			reportCard(b, "Warranty Expiring", "Assets with expiring warranty", "⚠️"),
			reportCard(b, "Cost Analysis", "Total cost of ownership", "💰"),
		)
	})

	h.render(w, page)
}

func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
	page := h.pageLayout("settings", "Settings", "Configure application settings", func(b *mi.Builder) mi.Node {
		states := []mdy.ComponentState{
			{ID: "general", Label: "General", Active: true, Content: func(b *mi.Builder) mi.Node {
				return b.Div(mi.Class("p-6 space-y-4"),
					formField(b, "Company Name", "company", "text", "", "Acme Corporation", false),
					formField(b, "Default Currency", "currency", "text", "", "USD", false),
					formField(b, "Date Format", "dateformat", "text", "", "YYYY-MM-DD", false),
				)
			}},
			{ID: "notifications", Label: "Notifications", Content: func(b *mi.Builder) mi.Node {
				return b.Div(mi.Class("p-6 space-y-4"),
					b.Div(mi.Class("flex items-center justify-between py-3 border-b border-gray-200 dark:border-gray-700"),
						b.Div(
							b.P(mi.Class("text-sm font-medium text-gray-900 dark:text-white"), "Maintenance Reminders"),
							b.P(mi.Class("text-xs text-gray-500 dark:text-gray-400"), "Get notified before scheduled maintenance"),
						),
						b.Input(mi.Type("checkbox"), mi.Class("h-4 w-4 text-blue-600 rounded"), mi.Checked()),
					),
					b.Div(mi.Class("flex items-center justify-between py-3 border-b border-gray-200 dark:border-gray-700"),
						b.Div(
							b.P(mi.Class("text-sm font-medium text-gray-900 dark:text-white"), "Warranty Expiry Alerts"),
							b.P(mi.Class("text-xs text-gray-500 dark:text-gray-400"), "Alert when warranty is about to expire"),
						),
						b.Input(mi.Type("checkbox"), mi.Class("h-4 w-4 text-blue-600 rounded"), mi.Checked()),
					),
				)
			}},
			{ID: "integrations", Label: "Integrations", Content: func(b *mi.Builder) mi.Node {
				return b.Div(mi.Class("p-6"),
					b.P(mi.Class("text-gray-500 dark:text-gray-400"), "No integrations configured."),
				)
			}},
		}

		settingsTabs := mdy.Dyn("settings-tabs").
			States(states).
			Theme(h.theme).
			Minified().
			Build()

		return b.Div(mi.Class("bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700"),
			settingsTabs(b),
		)
	})

	h.render(w, page)
}

// =============================================================================
// FORM HANDLERS
// =============================================================================

func (h *Handler) AssetNew(w http.ResponseWriter, r *http.Request) {
	asset := &models.Asset{
		Status: "active",
	}
	
	page := h.pageLayout("assets", "New Asset", "Create a new asset record", func(b *mi.Builder) mi.Node {
		states := h.buildAssetDetailStates(b, asset, nil)

		detailTabs := mdy.Dyn("asset-detail-tabs").
			States(states).
			Theme(h.theme).
			Minified().
			Build()

		return b.Div(
			b.Div(mi.Class("flex items-center gap-2 text-sm text-gray-500 mb-4"),
				b.A(mi.Href("/assets"), mi.Class("hover:text-gray-700 dark:hover:text-gray-200"), "Assets"),
				b.Span("›"),
				b.Span(mi.Class("text-gray-900 dark:text-white"), "New Asset"),
			),
			b.Form(mi.Class("bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700"), mi.Method("POST"), mi.Action("/assets/new"),
				detailTabs(b),
				b.Div(mi.Class("flex items-center justify-between px-6 py-4 bg-gray-50 dark:bg-gray-900/50 border-t border-gray-200 dark:border-gray-700"),
					b.A(mi.Href("/assets"), mi.Class("px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600"), "Cancel"),
					b.Button(mi.Class("px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"), mi.Type("submit"), "Create Asset"),
				),
			),
		)
	})

	h.render(w, page)
}

func (h *Handler) AssetCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	asset := &models.Asset{
		Tag:          r.FormValue("tag"),
		Name:         r.FormValue("name"),
		Category:     r.FormValue("category"),
		Status:       r.FormValue("status"),
		Department:   r.FormValue("department"),
		AssignedTo:   r.FormValue("assigned"),
		Location:     r.FormValue("location"),
		Vendor:       r.FormValue("vendor"),
		Model:        r.FormValue("model"),
		SerialNumber: r.FormValue("serial"),
		PurchaseDate: r.FormValue("purchasedate"),
		Warranty:     r.FormValue("warranty"),
		Notes:        r.FormValue("notes"),
	}

	if cost := r.FormValue("purchasecost"); cost != "" {
		fmt.Sscanf(cost, "%f", &asset.PurchaseCost)
	}
	if val := r.FormValue("currentvalue"); val != "" {
		fmt.Sscanf(val, "%f", &asset.CurrentValue)
	}

	if err := h.store.CreateAsset(asset); err != nil {
		h.logger.Error("failed to create asset", "error", err)
		http.Error(w, "Failed to create asset", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/assets/"+asset.ID, http.StatusSeeOther)
}

func (h *Handler) AssetUpdate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	existing, err := h.store.GetAsset(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Update fields from form
	existing.Tag = r.FormValue("tag")
	existing.Name = r.FormValue("name")
	existing.Category = r.FormValue("category")
	existing.Status = r.FormValue("status")
	existing.Department = r.FormValue("department")
	existing.AssignedTo = r.FormValue("assigned")
	existing.Location = r.FormValue("location")
	existing.Vendor = r.FormValue("vendor")
	existing.Model = r.FormValue("model")
	existing.SerialNumber = r.FormValue("serial")
	existing.PurchaseDate = r.FormValue("purchasedate")
	existing.Warranty = r.FormValue("warranty")
	existing.Notes = r.FormValue("notes")

	if cost := r.FormValue("purchasecost"); cost != "" {
		fmt.Sscanf(cost, "%f", &existing.PurchaseCost)
	}
	if val := r.FormValue("currentvalue"); val != "" {
		fmt.Sscanf(val, "%f", &existing.CurrentValue)
	}

	if err := h.store.UpdateAsset(existing); err != nil {
		h.logger.Error("failed to update asset", "error", err)
		http.Error(w, "Failed to update asset", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/assets/"+id, http.StatusSeeOther)
}

func (h *Handler) SettingsSave(w http.ResponseWriter, r *http.Request) {
	// Settings would be saved to a config store
	// For now, just redirect back
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

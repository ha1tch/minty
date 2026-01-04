package ui

import (
	"fmt"
	"strings"

	mi "github.com/ha1tch/minty"
	mdy "github.com/ha1tch/minty/mintydyn"
	"github.com/ha1tch/assettrack/internal/models"
)

// =============================================================================
// LAYOUT
// =============================================================================

const globalCSS = `
.icon { font-style: normal; }
::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track { background: #f1f1f1; }
.dark ::-webkit-scrollbar-track { background: #1f2937; }
::-webkit-scrollbar-thumb { background: #c1c1c1; border-radius: 3px; }
.dark ::-webkit-scrollbar-thumb { background: #4b5563; }
*, *::before, *::after { transition: background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease; }
`

// darkMode provides theme toggling using minty's built-in dark mode support.
// Uses Tailwind's class-based dark mode with localStorage persistence.
var darkMode = mi.DarkModeTailwind(
	mi.DarkModeStorage("darkMode"), // Match existing localStorage key
	mi.DarkModeMinify(),
)

func (h *Handler) pageLayout(activePage, title, subtitle string, content mi.H) mi.H {
	return func(b *mi.Builder) mi.Node {
		return mi.NewFragment(
			mi.Raw("<!DOCTYPE html>"),
			b.Html(mi.Lang("en"),
				b.Head(
					b.Title("AssetTrack - "+title),
					b.Meta(mi.Charset("UTF-8")),
					b.Meta(mi.Name("viewport"), mi.Content("width=device-width, initial-scale=1")),
					b.Script(mi.Src("https://cdn.tailwindcss.com")),
					b.Script(mi.Raw(`tailwind.config = { darkMode: 'class' }`)),
					b.Style(mi.Raw(globalCSS)),
					darkMode.Script(b), // Uses minty's DarkMode API
				),
				b.Body(mi.Class("bg-gray-100 dark:bg-gray-900 transition-colors"),
					b.Div(mi.Class("flex"),
						sidebar(b, activePage),
						b.Div(mi.Class("flex-1 ml-64 min-h-screen"),
							header(b, title, subtitle),
							b.Main(mi.Class("p-6"), content(b)),
						),
					),
				),
			),
		)
	}
}

func sidebar(b *mi.Builder, activePage string) mi.Node {
	navItems := []struct{ Icon, Label, Href, ID string }{
		{"dashboard", "Dashboard", "/", "dashboard"},
		{"assets", "Assets", "/assets", "assets"},
		{"maintenance", "Maintenance", "/maintenance", "maintenance"},
		{"reports", "Reports", "/reports", "reports"},
		{"settings", "Settings", "/settings", "settings"},
	}

	navNodes := make([]mi.Node, len(navItems))
	for i, item := range navItems {
		class := "flex items-center gap-3 px-4 py-2.5 text-sm font-medium rounded-lg transition-colors"
		if item.ID == activePage {
			class += " bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400"
		} else {
			class += " text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200"
		}
		navNodes[i] = b.A(mi.Href(item.Href), mi.Class(class), icon(item.Icon)(b), item.Label)
	}

	return b.Aside(mi.Class("w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 min-h-screen fixed left-0 top-0"),
		b.Div(mi.Class("p-4 border-b border-gray-200 dark:border-gray-700"),
			b.H1(mi.Class("text-xl font-bold text-gray-900 dark:text-white"), "AssetTrack"),
			b.P(mi.Class("text-xs text-gray-500 dark:text-gray-400"), "Enterprise Asset Management"),
		),
		b.Nav(mi.Class("p-4 space-y-1"), mi.NewFragment(navNodes...)),
		b.Div(mi.Class("absolute bottom-0 left-0 w-64 p-4 border-t border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800"),
			b.Div(mi.Class("flex items-center gap-3"),
				b.Div(mi.Class("w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white text-sm font-medium"), "JD"),
				b.Div(
					b.P(mi.Class("text-sm font-medium text-gray-900 dark:text-white"), "John Doe"),
					b.P(mi.Class("text-xs text-gray-500 dark:text-gray-400"), "Administrator"),
				),
			),
		),
	)
}

func header(b *mi.Builder, title, subtitle string) mi.Node {
	return b.Header(mi.Class("bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-4"),
		b.Div(mi.Class("flex items-center justify-between"),
			b.Div(
				b.H2(mi.Class("text-2xl font-bold text-gray-900 dark:text-white"), title),
				b.P(mi.Class("text-sm text-gray-500 dark:text-gray-400"), subtitle),
			),
			b.Div(mi.Class("flex items-center gap-4"),
				b.Div(mi.Class("relative"),
					b.Span(mi.Class("absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400"), icon("search")(b)),
					b.Input(
						mi.Type("search"), mi.Placeholder("Search..."),
						mi.Class("pl-10 pr-4 py-2 text-sm border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 w-64"),
					),
				),
				// Dark mode toggle using minty's DarkMode API
				darkMode.Toggle(b,
					mi.Class("p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"),
					mi.Attr("title", "Toggle dark mode"),
				),
				b.Button(mi.Class("p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 relative"), mi.Type("button"),
					icon("notification")(b),
					b.Span(mi.Class("absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full")),
				),
			),
		),
	)
}

// =============================================================================
// COMPONENTS
// =============================================================================

func icon(name string) mi.H {
	icons := map[string]string{
		"dashboard": "📊", "assets": "💻", "maintenance": "🔧",
		"reports": "📈", "settings": "⚙️", "users": "👥",
		"search": "🔍", "filter": "⏳",
		"edit": "✏️", "delete": "🗑️", "view": "👁️",
		"export": "📤", "import": "📥", "refresh": "🔄",
		"notification": "🔔", "check": "✓", "warning": "⚠️",
	}

	if name == "add" {
		return func(b *mi.Builder) mi.Node {
			return mi.Raw(`<svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>`)
		}
	}

	ic := icons[name]
	if ic == "" {
		ic = "•"
	}
	return func(b *mi.Builder) mi.Node {
		return b.Span(mi.Class("icon"), ic)
	}
}

func statusBadge(b *mi.Builder, status string) mi.Node {
	colors := map[string]string{
		"active":      "bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300",
		"maintenance": "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-300",
		"retired":     "bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400",
		"pending":     "bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300",
		"completed":   "bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300",
	}
	colorClass := colors[status]
	if colorClass == "" {
		colorClass = "bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400"
	}
	return b.Span(mi.Class("px-2 py-1 text-xs font-medium rounded-full "+colorClass), status)
}

func statCard(b *mi.Builder, title, value, change string, positive bool, iconName string) mi.Node {
	changeColor := "text-green-600 dark:text-green-400"
	if !positive {
		changeColor = "text-red-600 dark:text-red-400"
	}
	return b.Div(mi.Class("bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4"),
		b.Div(mi.Class("flex items-center justify-between"),
			b.Div(
				b.P(mi.Class("text-sm font-medium text-gray-500 dark:text-gray-400"), title),
				b.P(mi.Class("text-2xl font-semibold text-gray-900 dark:text-white mt-1"), value),
				b.P(mi.Class("text-sm mt-1 "+changeColor), change),
			),
			b.Div(mi.Class("text-3xl opacity-20"), icon(iconName)(b)),
		),
	)
}

func categoryBar(b *mi.Builder, name string, count int, percent int) mi.Node {
	return b.Div(
		b.Div(mi.Class("flex justify-between text-sm mb-1"),
			b.Span(mi.Class("text-gray-700 dark:text-gray-300"), name),
			b.Span(mi.Class("text-gray-500 dark:text-gray-400"), fmt.Sprintf("%d assets", count)),
		),
		b.Div(mi.Class("w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2"),
			b.Div(mi.Class("bg-blue-600 h-2 rounded-full"), mi.Style(fmt.Sprintf("width: %d%%", percent))),
		),
	)
}

func activityItem(b *mi.Builder, asset, action, time string) mi.Node {
	return b.Div(mi.Class("flex items-start gap-3 py-2 border-b border-gray-100 dark:border-gray-700 last:border-0"),
		b.Div(mi.Class("w-2 h-2 mt-2 rounded-full bg-blue-500")),
		b.Div(
			b.P(mi.Class("text-sm text-gray-900 dark:text-white"), asset+" - "+action),
			b.P(mi.Class("text-xs text-gray-500 dark:text-gray-400"), time),
		),
	)
}

func reportCard(b *mi.Builder, title, desc, iconEmoji string) mi.Node {
	return b.Div(mi.Class("bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6 hover:shadow-md dark:hover:shadow-lg dark:hover:shadow-gray-900/50 transition-shadow cursor-pointer"),
		b.Div(mi.Class("text-3xl mb-4"), iconEmoji),
		b.H3(mi.Class("text-lg font-medium text-gray-900 dark:text-white"), title),
		b.P(mi.Class("text-sm text-gray-500 dark:text-gray-400 mt-1"), desc),
		b.Div(mi.Class("mt-4 text-sm text-blue-600 dark:text-blue-400"), "Generate report →"),
	)
}

func formField(b *mi.Builder, label, name, fieldType, placeholder, value string, required bool) mi.Node {
	id := "field-" + name
	attrs := []mi.Attribute{
		mi.Class("w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"),
		mi.ID(id), mi.Name(name), mi.Type(fieldType), mi.Placeholder(placeholder), mi.Value(value),
	}
	if required {
		attrs = append(attrs, mi.Required())
	}
	labelClass := "block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
	if required {
		labelClass += " after:content-['*'] after:ml-0.5 after:text-red-500"
	}
	return b.Div(mi.Class("mb-4"),
		b.Label(mi.Class(labelClass), mi.For(id), label),
		b.Input(attrs...),
	)
}

func selectField(b *mi.Builder, label, name string, options []struct{ Value, Text string }, selected string, required bool) mi.Node {
	id := "field-" + name
	optionNodes := make([]mi.Node, len(options)+1)
	optionNodes[0] = b.Option(mi.Value(""), "Select...")
	for i, opt := range options {
		attrs := []interface{}{mi.Value(opt.Value)}
		if opt.Value == selected {
			attrs = append(attrs, mi.Selected())
		}
		attrs = append(attrs, opt.Text)
		optionNodes[i+1] = b.Option(attrs...)
	}
	labelClass := "block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
	if required {
		labelClass += " after:content-['*'] after:ml-0.5 after:text-red-500"
	}
	selectAttrs := []interface{}{
		mi.Class("w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"),
		mi.ID(id), mi.Name(name),
	}
	if required {
		selectAttrs = append(selectAttrs, mi.Required())
	}
	selectAttrs = append(selectAttrs, mi.NewFragment(optionNodes...))
	return b.Div(mi.Class("mb-4"),
		b.Label(mi.Class(labelClass), mi.For(id), label),
		b.Select(selectAttrs...),
	)
}

func textareaField(b *mi.Builder, label, name, placeholder, value string, rows int) mi.Node {
	id := "field-" + name
	return b.Div(mi.Class("mb-4"),
		b.Label(mi.Class("block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"), mi.For(id), label),
		b.Textarea(
			mi.Class("w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"),
			mi.ID(id), mi.Name(name), mi.Placeholder(placeholder), mi.Rows(rows),
			value,
		),
	)
}

// =============================================================================
// ASSET TABLE
// =============================================================================

func (h *Handler) assetTable(b *mi.Builder, assets []models.Asset) mi.Node {
	rows := make([]mi.Node, len(assets))
	for i, asset := range assets {
		rows[i] = b.Tr(
			mi.Class("hover:bg-gray-50 dark:hover:bg-gray-700 asset-row"),
			mi.Data("status", asset.Status),
			mi.Data("category", asset.Category),
			mi.Data("name", strings.ToLower(asset.Name)),
			b.Td(mi.Class("px-4 py-3"), b.Input(mi.Type("checkbox"), mi.Class("rounded border-gray-300 dark:border-gray-600"))),
			b.Td(mi.Class("px-4 py-3"),
				b.A(mi.Href("/assets/"+asset.ID), mi.Class("block"),
					b.P(mi.Class("font-medium text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300"), asset.Name),
					b.P(mi.Class("text-xs text-gray-500 dark:text-gray-400"), asset.Tag),
				),
			),
			b.Td(mi.Class("px-4 py-3 text-sm text-gray-600 dark:text-gray-400"), asset.Category),
			b.Td(mi.Class("px-4 py-3"), statusBadge(b, asset.Status)),
			b.Td(mi.Class("px-4 py-3 text-sm text-gray-600 dark:text-gray-400"), asset.Location),
			b.Td(mi.Class("px-4 py-3 text-sm text-gray-600 dark:text-gray-400"), asset.AssignedTo),
			b.Td(mi.Class("px-4 py-3 text-sm text-gray-600 dark:text-gray-400"), fmt.Sprintf("$%.2f", asset.CurrentValue)),
			b.Td(mi.Class("px-4 py-3"),
				b.Div(mi.Class("flex items-center gap-2"),
					b.A(mi.Href("/assets/"+asset.ID), mi.Class("p-1 text-gray-400 hover:text-blue-600"), mi.Attr("title", "View"), icon("view")(b)),
					b.A(mi.Href("/assets/"+asset.ID+"/edit"), mi.Class("p-1 text-gray-400 hover:text-blue-600"), mi.Attr("title", "Edit"), icon("edit")(b)),
				),
			),
		)
	}

	return b.Table(mi.Class("w-full"),
		b.Thead(mi.Class("bg-gray-50 dark:bg-gray-900/50 border-b border-gray-200 dark:border-gray-700"),
			b.Tr(
				b.Th(mi.Class("px-4 py-3 text-left w-10"), b.Input(mi.Type("checkbox"), mi.Class("rounded border-gray-300 dark:border-gray-600"))),
				b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Asset"),
				b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Category"),
				b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Status"),
				b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Location"),
				b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Assigned To"),
				b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Value"),
				b.Th(mi.Class("px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"), "Actions"),
			),
		),
		b.Tbody(mi.Class("divide-y divide-gray-200 dark:divide-gray-700"), mi.NewFragment(rows...)),
	)
}

// =============================================================================
// ASSET DETAIL STATES
// =============================================================================

func (h *Handler) buildAssetDetailStates(b *mi.Builder, asset *models.Asset, records []models.MaintenanceRecord) []mdy.ComponentState {
	categories := []struct{ Value, Text string }{
		{"Laptops", "Laptops"}, {"Monitors", "Monitors"}, {"Servers", "Servers"},
		{"Network", "Network Equipment"}, {"Printers", "Printers"}, {"Other", "Other"},
	}
	statuses := []struct{ Value, Text string }{
		{"active", "Active"}, {"maintenance", "Maintenance"}, {"retired", "Retired"},
	}
	departments := []struct{ Value, Text string }{
		{"Engineering", "Engineering"}, {"Sales", "Sales"}, {"Marketing", "Marketing"},
		{"Finance", "Finance"}, {"HR", "HR"}, {"IT Operations", "IT Operations"}, {"Shared", "Shared"},
	}

	return []mdy.ComponentState{
		{
			ID: "details", Label: "Details", Active: true,
			Content: func(b *mi.Builder) mi.Node {
				return b.Div(mi.Class("p-6"),
					b.Div(mi.Class("grid grid-cols-1 md:grid-cols-2 gap-6"),
						b.Div(
							b.H4(mi.Class("text-sm font-medium text-gray-900 dark:text-white mb-4"), "Basic Information"),
							formField(b, "Asset Tag", "tag", "text", "", asset.Tag, true),
							formField(b, "Asset Name", "name", "text", "", asset.Name, true),
							selectField(b, "Category", "category", categories, asset.Category, true),
							selectField(b, "Status", "status", statuses, asset.Status, true),
						),
						b.Div(
							b.H4(mi.Class("text-sm font-medium text-gray-900 dark:text-white mb-4"), "Assignment"),
							selectField(b, "Department", "department", departments, asset.Department, true),
							formField(b, "Assigned To", "assigned", "text", "", asset.AssignedTo, false),
							formField(b, "Location", "location", "text", "", asset.Location, true),
						),
					),
					b.Div(mi.Class("grid grid-cols-1 md:grid-cols-2 gap-6 mt-6"),
						b.Div(
							b.H4(mi.Class("text-sm font-medium text-gray-900 dark:text-white mb-4"), "Hardware"),
							formField(b, "Vendor", "vendor", "text", "", asset.Vendor, false),
							formField(b, "Model", "model", "text", "", asset.Model, false),
							formField(b, "Serial Number", "serial", "text", "", asset.SerialNumber, false),
						),
						b.Div(
							b.H4(mi.Class("text-sm font-medium text-gray-900 dark:text-white mb-4"), "Warranty"),
							formField(b, "Purchase Date", "purchasedate", "date", "", asset.PurchaseDate, false),
							formField(b, "Warranty Expiry", "warranty", "date", "", asset.Warranty, false),
						),
					),
					b.Div(mi.Class("mt-6"),
						textareaField(b, "Notes", "notes", "Additional notes...", asset.Notes, 3),
					),
				)
			},
		},
		{
			ID: "financial", Label: "Financial",
			Content: func(b *mi.Builder) mi.Node {
				depreciation := asset.PurchaseCost - asset.CurrentValue
				depPercent := 0.0
				if asset.PurchaseCost > 0 {
					depPercent = (asset.CurrentValue / asset.PurchaseCost) * 100
				}
				return b.Div(mi.Class("p-6"),
					b.Div(mi.Class("grid grid-cols-1 md:grid-cols-3 gap-6"),
						b.Div(
							b.H4(mi.Class("text-sm font-medium text-gray-900 dark:text-white mb-4"), "Purchase"),
							formField(b, "Purchase Date", "purchasedate", "date", "", asset.PurchaseDate, false),
							formField(b, "Purchase Cost", "purchasecost", "number", "", fmt.Sprintf("%.2f", asset.PurchaseCost), false),
						),
						b.Div(
							b.H4(mi.Class("text-sm font-medium text-gray-900 dark:text-white mb-4"), "Current Value"),
							formField(b, "Current Value", "currentvalue", "number", "", fmt.Sprintf("%.2f", asset.CurrentValue), false),
							b.Div(mi.Class("mt-4 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-100 dark:border-blue-800 rounded-lg"),
								b.P(mi.Class("text-sm text-blue-800 dark:text-blue-300"), fmt.Sprintf("Depreciation: $%.2f", depreciation)),
								b.P(mi.Class("text-xs text-blue-600 dark:text-blue-400 mt-1"), fmt.Sprintf("%.1f%% of original value", depPercent)),
							),
						),
						b.Div(
							b.H4(mi.Class("text-sm font-medium text-gray-900 dark:text-white mb-4"), "Summary"),
							summaryItem(b, "Original Cost", fmt.Sprintf("$%.2f", asset.PurchaseCost)),
							summaryItem(b, "Current Value", fmt.Sprintf("$%.2f", asset.CurrentValue)),
							summaryItem(b, "Total Depreciation", fmt.Sprintf("$%.2f", depreciation)),
						),
					),
				)
			},
		},
		{
			ID: "maintenance", Label: "Maintenance",
			Content: func(b *mi.Builder) mi.Node {
				return b.Div(mi.Class("p-6"),
					b.Div(mi.Class("flex justify-between items-center mb-4"),
						b.H4(mi.Class("text-sm font-medium text-gray-900 dark:text-white"), "Maintenance History"),
						b.Button(mi.Class("inline-flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"), mi.Type("button"),
							icon("add")(b), "Schedule Maintenance",
						),
					),
					maintenanceTable(b, records),
					maintenanceSummary(b, records),
				)
			},
		},
		{
			ID: "history", Label: "History",
			Content: func(b *mi.Builder) mi.Node {
				return b.Div(mi.Class("p-6"),
					b.H4(mi.Class("text-sm font-medium text-gray-900 dark:text-white mb-4"), "Audit Trail"),
					b.Div(mi.Class("space-y-4"),
						historyEntry(b, "2025-01-03 14:32", "John Doe", "Updated", "Changed status to 'active'"),
						historyEntry(b, "2025-01-02 09:15", "System", "Maintenance", "Scheduled maintenance completed"),
						historyEntry(b, "2024-12-15 11:20", "Jane Smith", "Reassigned", "Transferred to John Smith"),
						historyEntry(b, asset.PurchaseDate+" 09:00", "System", "Created", "Asset record created"),
					),
				)
			},
		},
	}
}

func summaryItem(b *mi.Builder, label, value string) mi.Node {
	return b.Div(mi.Class("flex justify-between py-2 border-b border-gray-100 dark:border-gray-700"),
		b.Span(mi.Class("text-sm text-gray-500 dark:text-gray-400"), label),
		b.Span(mi.Class("text-sm font-medium text-gray-900 dark:text-white"), value),
	)
}

func maintenanceTable(b *mi.Builder, records []models.MaintenanceRecord) mi.Node {
	if len(records) == 0 {
		return b.Div(mi.Class("text-center py-8 text-gray-500 dark:text-gray-400"), b.P("No maintenance records"))
	}
	rows := make([]mi.Node, len(records))
	for i, r := range records {
		rows[i] = b.Tr(mi.Class("hover:bg-gray-50 dark:hover:bg-gray-700"),
			b.Td(mi.Class("px-4 py-3 text-sm text-gray-900 dark:text-gray-100"), r.Date),
			b.Td(mi.Class("px-4 py-3"), b.Span(mi.Class("px-2 py-0.5 text-xs rounded border bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800"), r.Type)),
			b.Td(mi.Class("px-4 py-3 text-sm text-gray-600 dark:text-gray-400"), r.Description),
			b.Td(mi.Class("px-4 py-3 text-sm text-gray-900 dark:text-gray-100"), fmt.Sprintf("$%.2f", r.Cost)),
			b.Td(mi.Class("px-4 py-3"), statusBadge(b, r.Status)),
		)
	}
	return b.Table(mi.Class("w-full text-sm"),
		b.Thead(mi.Class("bg-gray-50 dark:bg-gray-900/50"),
			b.Tr(
				b.Th(mi.Class("px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400"), "Date"),
				b.Th(mi.Class("px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400"), "Type"),
				b.Th(mi.Class("px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400"), "Description"),
				b.Th(mi.Class("px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400"), "Cost"),
				b.Th(mi.Class("px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400"), "Status"),
			),
		),
		b.Tbody(mi.Class("divide-y divide-gray-200 dark:divide-gray-700"), mi.NewFragment(rows...)),
	)
}

func maintenanceSummary(b *mi.Builder, records []models.MaintenanceRecord) mi.Node {
	var total float64
	for _, r := range records {
		total += r.Cost
	}
	lastService := "N/A"
	if len(records) > 0 && len(records[0].Date) >= 7 {
		lastService = records[0].Date[:7]
	}
	return b.Div(mi.Class("mt-6 pt-6 border-t border-gray-200 dark:border-gray-700 grid grid-cols-3 gap-4"),
		b.Div(mi.Class("text-center p-4 bg-gray-50 dark:bg-gray-900/50 rounded-lg"),
			b.P(mi.Class("text-2xl font-semibold text-gray-900 dark:text-white"), fmt.Sprintf("%d", len(records))),
			b.P(mi.Class("text-sm text-gray-500 dark:text-gray-400"), "Records"),
		),
		b.Div(mi.Class("text-center p-4 bg-gray-50 dark:bg-gray-900/50 rounded-lg"),
			b.P(mi.Class("text-2xl font-semibold text-gray-900 dark:text-white"), fmt.Sprintf("$%.0f", total)),
			b.P(mi.Class("text-sm text-gray-500 dark:text-gray-400"), "Total Cost"),
		),
		b.Div(mi.Class("text-center p-4 bg-gray-50 dark:bg-gray-900/50 rounded-lg"),
			b.P(mi.Class("text-2xl font-semibold text-gray-900 dark:text-white"), lastService),
			b.P(mi.Class("text-sm text-gray-500 dark:text-gray-400"), "Last Service"),
		),
	)
}

func historyEntry(b *mi.Builder, timestamp, user, action, details string) mi.Node {
	return b.Div(mi.Class("flex gap-4 p-3 bg-gray-50 dark:bg-gray-900/50 rounded-lg"),
		b.Div(mi.Class("flex-shrink-0 w-2 h-2 mt-2 rounded-full bg-blue-500")),
		b.Div(mi.Class("flex-1"),
			b.Div(mi.Class("flex items-center gap-2 mb-1"),
				b.Span(mi.Class("text-sm font-medium text-gray-900 dark:text-white"), user),
				b.Span(mi.Class("px-2 py-0.5 text-xs rounded border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-300"), action),
			),
			b.P(mi.Class("text-sm text-gray-600 dark:text-gray-400"), details),
			b.P(mi.Class("text-xs text-gray-400 mt-1"), timestamp),
		),
	)
}

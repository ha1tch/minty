# Minty System Documentation - Part 10
## Presentation Layer Architecture: Clean UI Separation

---

### Table of Contents
1. [Presentation Layer Overview](#presentation-layer-overview)
2. [Separation of Concerns Architecture](#separation-of-concerns-architecture)
3. [Data Preparation Patterns](#data-preparation-patterns)
4. [Presentation Adapters](#presentation-adapters)
5. [Cross-Domain UI Composition](#cross-domain-ui-composition)
6. [Component Design Patterns](#component-design-patterns)
7. [Testing Presentation Layer](#testing-presentation-layer)
8. [Advanced Presentation Patterns](#advanced-presentation-patterns)

---

## Presentation Layer Overview

The Minty System's presentation layer architecture provides a clean separation between business logic and UI concerns, enabling maintainable, testable, and flexible user interfaces. This layer serves as the bridge between pure domain logic and visual presentation, handling data transformation, UI component generation, and theme integration.

### Architectural Boundaries

```
┌─────────────────────────────────────────────────┐
│                 Domain Layer                    │ ← Pure business logic
│            (Zero UI Dependencies)               │
├─────────────────────────────────────────────────┤
│                Data Preparation                 │ ← Business data → Display data
│           (Domain-Specific Functions)           │
├─────────────────────────────────────────────────┤
│                Presentation Layer               │ ← UI component generation
│          (Theme-Aware UI Components)            │
├─────────────────────────────────────────────────┤
│                   Theme Layer                   │ ← Visual styling
│          (Bootstrap, Tailwind, Custom)          │
├─────────────────────────────────────────────────┤
│                  HTML Output                    │ ← Final rendered HTML
└─────────────────────────────────────────────────┘
```

### Key Principles

**Clean Separation**: Domain logic has no knowledge of UI concerns; presentation logic has no business rules.

**Data Transformation**: Business entities are transformed into display-ready data structures before UI generation.

**Theme Integration**: Presentation components work with any theme through the theme interface.

**Composability**: UI components can be composed to build complex interfaces from simple building blocks.

**Testability**: Each layer can be tested independently with clear boundaries and dependencies.

---

## Separation of Concerns Architecture

### Domain Independence

Domain packages contain pure business logic with zero UI dependencies:

```go
// mintyfin domain - pure business logic
package mintyfin

// NO imports of UI packages
import (
    "time"
    "errors"
    "github.com/ha1tch/mintyex" // Only shared utilities
)

type Account struct {
    ID      string        `json:"id"`
    Name    string        `json:"name"`
    Balance mintyex.Money `json:"balance"`
    Status  string        `json:"status"`
    Type    string        `json:"type"`
}

func (fs *FinanceService) CreateAccount(name, accountType string, 
    initialBalance mintyex.Money, customerID string) (*Account, error) {
    // Pure business logic - no UI concerns
    account := Account{
        ID:      generateID("acc"),
        Name:    name,
        Balance: initialBalance,
        Status:  mintyex.StatusActive,
        Type:    accountType,
    }
    
    // Business validation
    if errors := ValidateAccount(account); errors.HasErrors() {
        return nil, errors
    }
    
    fs.accounts = append(fs.accounts, account)
    return &account, nil
}
```

### Presentation Adapter Pattern

Presentation adapters bridge domain data and UI components:

```go
// mintyfinui - presentation adapter
package mintyfinui

import (
    mi "github.com/ha1tch/minty"
    mui "github.com/ha1tch/mintyui"
    miex "github.com/ha1tch/mintyex"
    mifi "github.com/ha1tch/mintyfin"  // Domain import
)

// Convert domain account to UI card component
func AccountSummaryCard(theme mui.Theme, account mifi.Account) mi.H {
    // Data preparation - domain data to display data
    displayData := mifi.PrepareAccountForDisplay(account)
    
    // UI generation using theme
    return theme.Card(account.Name, func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("mifi_account_summary"),
            b.Div(mi.Class("mifi_account_info"),
                b.P(mi.Class("mifi_account_type"), 
                    displayData.TypeDisplay),
                b.Div(mi.Class("mifi_account_balance"),
                    b.Strong(displayData.FormattedBalance),
                ),
                StatusBadge(theme, displayData.StatusDisplay, 
                    displayData.StatusClass),
            ),
            b.Div(mi.Class("mifi_account_actions"),
                mui.DomainButton(theme, "mifi", "View Details", "secondary",
                    mi.Href("/accounts/"+account.ID)),
            ),
        )
    })
}
```

### Responsibility Distribution

**Domain Layer Responsibilities:**
- Business rules and validation
- Entity lifecycle management
- Data integrity and consistency
- Business calculations and logic
- Domain-specific operations

**Data Preparation Responsibilities:**
- Transform business entities to display structures
- Format data for human consumption
- Calculate display-specific values
- Prepare data for UI consumption

**Presentation Layer Responsibilities:**
- Generate UI components from display data
- Apply theme-specific styling
- Handle UI composition and layout
- Manage user interaction patterns

**Theme Layer Responsibilities:**
- Provide visual styling and branding
- Implement CSS framework integration
- Handle responsive design concerns
- Manage visual consistency

---

## Data Preparation Patterns

### Display Data Structures

Data preparation transforms business entities into display-ready structures:

```go
// Business entity (domain layer)
type Account struct {
    ID          string
    Name        string
    Balance     mintyex.Money
    Status      string
    Type        string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Display data structure (presentation layer)
type AccountDisplayData struct {
    ID                  string
    Name               string
    FormattedBalance   string    // "$1,234.56"
    StatusDisplay      string    // "Active"
    StatusClass        string    // "status-active"
    TypeDisplay        string    // "Checking Account"
    TypeIcon           string    // "💳"
    FormattedCreatedAt string    // "January 15, 2024"
    IsOverdrawn        bool      
    DaysOld            int
}
```

### Data Preparation Functions

Domain packages provide functions to prepare data for display:

```go
// mintyfin domain - data preparation function
func PrepareAccountForDisplay(account Account) AccountDisplayData {
    return AccountDisplayData{
        ID:               account.ID,
        Name:            account.Name,
        FormattedBalance: account.Balance.Format(),
        StatusDisplay:   formatStatus(account.Status),
        StatusClass:     getStatusClass(account.Status),
        TypeDisplay:     formatAccountType(account.Type),
        TypeIcon:        getAccountTypeIcon(account.Type),
        FormattedCreatedAt: account.CreatedAt.Format("January 2, 2006"),
        IsOverdrawn:     account.Balance.IsNegative(),
        DaysOld:         int(time.Since(account.CreatedAt).Hours() / 24),
    }
}

func formatStatus(status string) string {
    switch status {
    case mintyex.StatusActive:
        return "Active"
    case mintyex.StatusInactive:
        return "Inactive"
    case mintyex.StatusPending:
        return "Pending Activation"
    case mintyex.StatusSuspended:
        return "Suspended"
    default:
        return "Unknown"
    }
}

func getStatusClass(status string) string {
    switch status {
    case mintyex.StatusActive:
        return "status-active"
    case mintyex.StatusInactive:
        return "status-inactive"
    case mintyex.StatusPending:
        return "status-pending"
    case mintyex.StatusSuspended:
        return "status-suspended"
    default:
        return "status-unknown"
    }
}

func getAccountTypeIcon(accountType string) string {
    switch accountType {
    case "checking":
        return "💳"
    case "savings":
        return "🏦"
    case "investment":
        return "📈"
    case "credit":
        return "💸"
    default:
        return "📄"
    }
}
```

### Complex Data Preparation

More sophisticated display data with computed values:

```go
type TransactionDisplayData struct {
    ID                    string
    FormattedAmount       string
    AmountClass           string  // "amount-credit" or "amount-debit"
    Description           string
    FormattedDate         string
    RelativeDate          string  // "2 days ago"
    CategoryDisplay       string
    CategoryIcon          string
    StatusDisplay         string
    StatusClass           string
    IsRecent              bool
    IsPending             bool
    ShowDetails           bool
}

func PrepareTransactionForDisplay(transaction Transaction, 
    currentUser User) TransactionDisplayData {
    
    now := time.Now()
    
    return TransactionDisplayData{
        ID:              transaction.ID,
        FormattedAmount: transaction.Amount.Format(),
        AmountClass:     getAmountClass(transaction.Type),
        Description:     transaction.Description,
        FormattedDate:   transaction.Date.Format("Jan 2, 2006"),
        RelativeDate:    formatRelativeDate(transaction.Date, now),
        CategoryDisplay: formatCategory(transaction.Category),
        CategoryIcon:    getCategoryIcon(transaction.Category),
        StatusDisplay:   formatTransactionStatus(transaction.Status),
        StatusClass:     getTransactionStatusClass(transaction.Status),
        IsRecent:        now.Sub(transaction.Date) < 7*24*time.Hour,
        IsPending:       transaction.Status == mintyex.StatusPending,
        ShowDetails:     canViewTransactionDetails(transaction, currentUser),
    }
}

func formatRelativeDate(date time.Time, now time.Time) string {
    diff := now.Sub(date)
    
    switch {
    case diff < time.Hour:
        return "Just now"
    case diff < 24*time.Hour:
        hours := int(diff.Hours())
        if hours == 1 {
            return "1 hour ago"
        }
        return fmt.Sprintf("%d hours ago", hours)
    case diff < 7*24*time.Hour:
        days := int(diff.Hours() / 24)
        if days == 1 {
            return "Yesterday"
        }
        return fmt.Sprintf("%d days ago", days)
    default:
        return date.Format("Jan 2, 2006")
    }
}
```

### Dashboard Data Preparation

Preparing complex dashboard data with multiple entity types:

```go
type DashboardData struct {
    AccountSummaries    []AccountDisplayData
    RecentTransactions  []TransactionDisplayData
    PendingInvoices     []InvoiceDisplayData
    TotalBalance        string
    MonthlyChange       string
    MonthlyChangeClass  string
    AlertCount          int
    HasCriticalAlerts   bool
    QuickActions        []QuickActionData
}

func PrepareDashboardData(service *FinanceService, userID string) DashboardData {
    accounts := service.GetAccountsByUser(userID)
    transactions := service.GetRecentTransactions(userID, 10)
    invoices := service.GetPendingInvoices(userID)
    
    // Calculate aggregated values
    totalBalance := calculateTotalBalance(accounts)
    monthlyChange := calculateMonthlyChange(accounts, service)
    alerts := service.GetAlerts(userID)
    
    return DashboardData{
        AccountSummaries: mintyex.Map(accounts, PrepareAccountForDisplay),
        RecentTransactions: mintyex.Map(transactions, func(t Transaction) TransactionDisplayData {
            return PrepareTransactionForDisplay(t, service.GetUser(userID))
        }),
        PendingInvoices: mintyex.Map(invoices, PrepareInvoiceForDisplay),
        TotalBalance: totalBalance.Format(),
        MonthlyChange: formatMonthlyChange(monthlyChange),
        MonthlyChangeClass: getChangeClass(monthlyChange),
        AlertCount: len(alerts),
        HasCriticalAlerts: hasCriticalAlerts(alerts),
        QuickActions: prepareQuickActions(userID),
    }
}
```

---

## Presentation Adapters

### Finance UI Adapter (mintyfinui)

The finance presentation adapter converts finance domain data into UI components:

```go
package mintyfinui

import (
    mi "github.com/ha1tch/minty"
    mui "github.com/ha1tch/mintyui"
    miex "github.com/ha1tch/mintyex"
    mifi "github.com/ha1tch/mintyfin"
)

const Domain = "mifi"

// Account UI Components
func AccountsTable(theme mui.Theme, accounts []mifi.Account) mi.H {
    headers := []string{"Account", "Type", "Balance", "Status", "Actions"}
    
    rows := miex.Map(accounts, func(account mifi.Account) []string {
        displayData := mifi.PrepareAccountForDisplay(account)
        actions := fmt.Sprintf(`
            <div class="mifi_account_table_actions">
                <a href="/accounts/%s" class="mifi_view_button">View</a>
                <a href="/accounts/%s/edit" class="mifi_edit_button">Edit</a>
            </div>`, account.ID, account.ID)
        
        return []string{
            account.Name,
            displayData.TypeDisplay,
            displayData.FormattedBalance,
            displayData.StatusDisplay,
            actions,
        }
    })
    
    return theme.Table(headers, rows)
}

// Transaction UI Components  
func TransactionList(theme mui.Theme, transactions []mifi.Transaction) mi.H {
    return func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("mifi_transaction_list"),
            miex.Map(transactions, func(txn mifi.Transaction) mi.H {
                return TransactionItem(theme, txn)
            })...,
        )
    }
}

func TransactionItem(theme mui.Theme, transaction mifi.Transaction) mi.H {
    displayData := mifi.PrepareTransactionForDisplay(transaction)
    
    return func(b *mi.Builder) mi.Node {
        itemClass := "mifi_transaction_item"
        if displayData.IsPending {
            itemClass += " mifi_transaction_pending"
        }
        if displayData.IsRecent {
            itemClass += " mifi_transaction_recent"
        }
        
        return b.Div(mi.Class(itemClass),
            b.Div(mi.Class("mifi_transaction_info"),
                b.Div(mi.Class("mifi_transaction_description"),
                    b.Span(mi.Class("mifi_category_icon"), displayData.CategoryIcon),
                    b.Span(transaction.Description),
                ),
                b.Div(mi.Class("mifi_transaction_meta"),
                    b.Small(displayData.RelativeDate),
                    b.Span(mi.Class(displayData.StatusClass), displayData.StatusDisplay),
                ),
            ),
            b.Div(mi.Class("mifi_transaction_amount " + displayData.AmountClass),
                b.Strong(displayData.FormattedAmount),
            ),
        )
    }
}

// Dashboard Components
func FinancialDashboard(theme mui.Theme, dashboardData mifi.DashboardData, 
    recentTxns []mifi.Transaction, pendingInvoices []mifi.Invoice) mi.H {
    
    return mui.Dashboard(theme, "Financial Dashboard",
        // Sidebar
        func(b *mi.Builder) mi.Node {
            return theme.Sidebar(func(b *mi.Builder) mi.Node {
                return FinancialNavigation(theme)
            })
        },
        
        // Main content
        func(b *mi.Builder) mi.Node {
            return b.Div(mi.Class("mifi_dashboard_main"),
                // Financial metrics
                FinancialMetrics(theme, dashboardData),
                
                // Recent transactions
                theme.Card("Recent Transactions", func(b *mi.Builder) mi.Node {
                    return TransactionList(theme, recentTxns)
                }),
                
                // Pending invoices
                miex.RenderIf(pendingInvoices, len(pendingInvoices) > 0,
                    func(invoice mifi.Invoice) mi.H {
                        return theme.Card("Pending Invoices", func(b *mi.Builder) mi.Node {
                            return InvoicesTable(theme, pendingInvoices)
                        })
                    })...,
            )
        },
    )
}
```

### Logistics UI Adapter (mintymoveui)

```go
package mintymoveui

import (
    mi "github.com/ha1tch/minty"
    mui "github.com/ha1tch/mintyui"
    miex "github.com/ha1tch/mintyex"
    mimo "github.com/ha1tch/mintymove"
)

const Domain = "mimo"

// Shipment UI Components
func ShipmentCard(theme mui.Theme, shipment mimo.Shipment) mi.H {
    displayData := mimo.PrepareShipmentForDisplay(shipment)
    
    return theme.Card("Shipment " + shipment.TrackingCode, func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("mimo_shipment_card"),
            b.Div(mi.Class("mimo_shipment_route"),
                b.Div(mi.Class("mimo_origin"),
                    b.Strong("From: "),
                    b.Span(shipment.Origin.FormatOneLine()),
                ),
                b.Div(mi.Class("mimo_destination"),
                    b.Strong("To: "),
                    b.Span(shipment.Destination.FormatOneLine()),
                ),
            ),
            
            b.Div(mi.Class("mimo_shipment_details"),
                b.P("Carrier: ", shipment.Carrier),
                b.P("Service: ", displayData.ServiceTypeDisplay),
                b.P("Cost: ", shipment.Cost.Format()),
                b.P("Weight: ", displayData.FormattedWeight),
            ),
            
            b.Div(mi.Class("mimo_shipment_status"),
                StatusBadge(theme, displayData.StatusDisplay, displayData.StatusClass),
                miex.If(displayData.HasTracking,
                    TrackingButton(theme, shipment.TrackingCode),
                ),
            ),
            
            miex.If(displayData.EstimatedDelivery != "",
                b.Div(mi.Class("mimo_delivery_info"),
                    b.P("Estimated Delivery: ", displayData.EstimatedDelivery),
                ),
            ),
        )
    })
}

// Vehicle Management Components
func VehicleFleetView(theme mui.Theme, vehicles []mimo.Vehicle) mi.H {
    return func(b *mi.Builder) mi.Node {
        // Group vehicles by status
        vehiclesByStatus := miex.GroupBy(vehicles, func(v mimo.Vehicle) string {
            return v.Status
        })
        
        return b.Div(mi.Class("mimo_fleet_overview"),
            b.H2("Vehicle Fleet"),
            
            miex.Map(vehiclesByStatus, func(status string, statusVehicles []mimo.Vehicle) mi.H {
                return theme.Card(fmt.Sprintf("%s Vehicles", strings.Title(status)), 
                    func(b *mi.Builder) mi.Node {
                        return b.Div(mi.Class("mimo_vehicle_grid"),
                            miex.Map(statusVehicles, func(v mimo.Vehicle) mi.H {
                                return VehicleCard(theme, v)
                            })...,
                        )
                    },
                )
            })...,
        )
    }
}

// Route Optimization UI
func RouteOptimizationPanel(theme mui.Theme, route mimo.Route, 
    optimization mimo.RouteOptimization) mi.H {
    
    return theme.Card("Route Optimization", func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("mimo_route_optimization"),
            b.Div(mi.Class("mimo_route_stats"),
                b.P("Total Distance: ", optimization.TotalDistance),
                b.P("Estimated Time: ", optimization.EstimatedTime),
                b.P("Fuel Cost: ", optimization.FuelCost.Format()),
                b.P("Optimized Stops: ", fmt.Sprintf("%d", len(route.Stops))),
            ),
            
            b.Div(mi.Class("mimo_route_stops"),
                b.H4("Delivery Sequence"),
                b.Ol(
                    miex.Map(route.Stops, func(stop mimo.RouteStop) mi.H {
                        return func(b *mi.Builder) mi.Node {
                            return b.Li(mi.Class("mimo_route_stop"),
                                b.Strong(stop.Address.FormatOneLine()),
                                b.P("ETA: ", stop.EstimatedArrival.Format("3:04 PM")),
                                b.Small(fmt.Sprintf("%d packages", len(stop.Packages))),
                            )
                        }
                    })...,
                ),
            ),
        )
    })
}
```

### E-commerce UI Adapter (mintycartui)

```go
package mintycartui

import (
    mi "github.com/ha1tch/minty"
    mui "github.com/ha1tch/mintyui"
    miex "github.com/ha1tch/mintyex"
    mica "github.com/ha1tch/mintycart"
)

const Domain = "mica"

// Product Catalog Components
func ProductGrid(theme mui.Theme, products []mica.Product) mi.H {
    return func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("mica_product_grid"),
            miex.ChunkAndRender(products, 3, func(productGroup []mica.Product) mi.H {
                return func(b *mi.Builder) mi.Node {
                    return b.Div(mi.Class("mica_product_row"),
                        miex.Map(productGroup, func(p mica.Product) mi.H {
                            return ProductCard(theme, p)
                        })...,
                    )
                }
            })...,
        )
    }
}

func ProductCard(theme mui.Theme, product mica.Product) mi.H {
    displayData := mica.PrepareProductForDisplay(product)
    
    return theme.Card("", func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("mica_product_card"),
            b.Div(mi.Class("mica_product_image"),
                b.Img(mi.Src(displayData.PrimaryImageURL), 
                      mi.Alt(product.Name),
                      mi.Class("mica_product_img")),
                miex.If(displayData.IsOnSale,
                    b.Div(mi.Class("mica_sale_badge"), "SALE"),
                ),
            ),
            
            b.Div(mi.Class("mica_product_info"),
                b.H3(mi.Class("mica_product_name"), product.Name),
                b.P(mi.Class("mica_product_description"), 
                    truncateText(product.Description, 100)),
                
                b.Div(mi.Class("mica_product_pricing"),
                    miex.If(displayData.IsOnSale,
                        b.Span(mi.Class("mica_original_price"), 
                               displayData.OriginalPrice),
                    ),
                    b.Span(mi.Class("mica_current_price"), 
                           displayData.CurrentPrice),
                ),
                
                b.Div(mi.Class("mica_product_meta"),
                    InventoryIndicator(theme, product.Inventory),
                    CategoryBadge(theme, displayData.CategoryDisplay),
                ),
            ),
            
            b.Div(mi.Class("mica_product_actions"),
                miex.If(displayData.CanPurchase,
                    mui.DomainButton(theme, Domain, "Add to Cart", "primary",
                        mi.DataAttr("product-id", product.ID),
                        mi.OnClick("addToCart(this)")),
                    b.Span(mi.Class("mica_out_of_stock"), "Out of Stock"),
                ),
                
                mui.DomainButton(theme, Domain, "View Details", "secondary",
                    mi.Href("/products/" + product.ID)),
            ),
        )
    })
}

// Shopping Cart Components
func ShoppingCart(theme mui.Theme, cart mica.Cart, 
    products []mica.Product) mi.H {
    
    displayData := mica.PrepareCartForDisplay(cart, products)
    
    return theme.Card("Shopping Cart", func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("mica_shopping_cart"),
            miex.RenderIf(cart.Items, len(cart.Items) > 0,
                func(item mica.CartItem) mi.H {
                    product := findProductByID(products, item.ProductID)
                    return CartItem(theme, item, product)
                })...,
            
            miex.If(len(cart.Items) == 0,
                b.Div(mi.Class("mica_empty_cart"),
                    b.P("Your cart is empty"),
                    mui.DomainButton(theme, Domain, "Continue Shopping", "primary",
                        mi.Href("/products")),
                ),
                
                // Cart summary
                b.Div(mi.Class("mica_cart_summary"),
                    b.Div(mi.Class("mica_cart_totals"),
                        b.P("Subtotal: ", displayData.FormattedSubtotal),
                        b.P("Tax: ", displayData.FormattedTax),
                        b.P("Shipping: ", displayData.FormattedShipping),
                        b.Hr(),
                        b.P(mi.Class("mica_cart_total"),
                            b.Strong("Total: ", displayData.FormattedTotal)),
                    ),
                    
                    b.Div(mi.Class("mica_cart_actions"),
                        mui.DomainButton(theme, Domain, "Update Cart", "secondary",
                            mi.OnClick("updateCart()")),
                        mui.DomainButton(theme, Domain, "Checkout", "primary",
                            mi.Href("/checkout")),
                    ),
                ),
            ),
        )
    })
}
```

---

## Cross-Domain UI Composition

### Unified Dashboard

Combining multiple domain UIs into a cohesive interface:

```go
func UnifiedBusinessDashboard(services *ApplicationServices, theme mui.Theme, 
    user User) mi.H {
    
    return mui.Dashboard(theme, "Business Command Center",
        // Unified sidebar
        UnifiedNavigation(theme, user),
        
        // Main dashboard content
        func(b *mi.Builder) mi.Node {
            return b.Div(mi.Class("unified_dashboard"),
                // Cross-domain metrics
                CrossDomainMetrics(theme, services, user),
                
                // Domain-specific sections
                b.Div(mi.Class("domain_sections"),
                    // Finance section
                    b.Section(mi.Class("finance_section"),
                        b.H2("Finance Overview"),
                        FinanceOverview(theme, services.Finance, user.ID),
                    ),
                    
                    // Logistics section  
                    b.Section(mi.Class("logistics_section"),
                        b.H2("Logistics Status"),
                        LogisticsOverview(theme, services.Logistics, user.ID),
                    ),
                    
                    // E-commerce section
                    b.Section(mi.Class("ecommerce_section"),
                        b.H2("Sales Dashboard"),
                        EcommerceOverview(theme, services.Ecommerce, user.ID),
                    ),
                ),
            )
        },
    )
}

func CrossDomainMetrics(theme mui.Theme, services *ApplicationServices, 
    user User) mi.H {
    
    return func(b *mi.Builder) mi.Node {
        // Gather cross-domain data
        totalRevenue := calculateTotalRevenue(services, user.ID)
        pendingOrders := services.Ecommerce.GetPendingOrderCount(user.ID)
        activeShipments := services.Logistics.GetActiveShipmentCount()
        accountBalance := services.Finance.GetTotalAccountBalance(user.ID)
        
        return b.Div(mi.Class("cross_domain_metrics"),
            mui.StatsCard(theme, "Total Revenue", totalRevenue.Format(), 
                "This month"),
            mui.StatsCard(theme, "Pending Orders", 
                fmt.Sprintf("%d", pendingOrders), "Awaiting fulfillment"),
            mui.StatsCard(theme, "Active Shipments", 
                fmt.Sprintf("%d", activeShipments), "In transit"),
            mui.StatsCard(theme, "Account Balance", accountBalance.Format(), 
                "Available funds"),
        )
    }
}
```

### Cross-Domain Workflows

UI components that span multiple domains:

```go
func OrderFulfillmentWorkflow(theme mui.Theme, services *ApplicationServices, 
    orderID string) mi.H {
    
    // Gather data from multiple domains
    order := services.Ecommerce.GetOrder(orderID)
    invoice := services.Finance.GetInvoiceByOrderID(orderID)
    shipment := services.Logistics.GetShipmentByOrderID(orderID)
    
    return theme.Card("Order Fulfillment", func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("order_fulfillment_workflow"),
            // Order step
            WorkflowStep(theme, "order", "Order Placed", 
                order.Status, func(b *mi.Builder) mi.Node {
                    return b.Div(
                        b.P("Order #", order.Number),
                        b.P("Customer: ", order.Customer.Name),
                        b.P("Total: ", order.Total.Format()),
                        b.P("Items: ", fmt.Sprintf("%d", len(order.Items))),
                    )
                }),
            
            // Invoice step
            WorkflowStep(theme, "invoice", "Invoice Created", 
                getInvoiceStatus(invoice), func(b *mi.Builder) mi.Node {
                    if invoice != nil {
                        return b.Div(
                            b.P("Invoice #", invoice.Number),
                            b.P("Amount: ", invoice.Amount.Format()),
                            b.P("Due Date: ", invoice.DueDate.Format("Jan 2, 2006")),
                        )
                    }
                    return b.P("Invoice pending creation")
                }),
            
            // Payment step
            WorkflowStep(theme, "payment", "Payment Processed", 
                getPaymentStatus(invoice), func(b *mi.Builder) mi.Node {
                    if invoice != nil && invoice.PaidAt != nil {
                        return b.Div(
                            b.P("Paid: ", invoice.PaidAt.Format("Jan 2, 2006")),
                            b.P("Amount: ", invoice.Amount.Format()),
                        )
                    }
                    return b.P("Payment pending")
                }),
            
            // Shipment step
            WorkflowStep(theme, "shipment", "Shipment Created", 
                getShipmentStatus(shipment), func(b *mi.Builder) mi.Node {
                    if shipment != nil {
                        return b.Div(
                            b.P("Tracking: ", shipment.TrackingCode),
                            b.P("Carrier: ", shipment.Carrier),
                            b.P("Status: ", shipment.Status),
                        )
                    }
                    return b.P("Shipment pending creation")
                }),
        )
    })
}

func WorkflowStep(theme mui.Theme, stepID, title, status string, 
    content mi.H) mi.H {
    
    return func(b *mi.Builder) mi.Node {
        stepClass := "workflow_step"
        statusIcon := ""
        
        switch status {
        case "completed":
            stepClass += " step_completed"
            statusIcon = "✅"
        case "pending":
            stepClass += " step_pending"
            statusIcon = "⏳"
        case "error":
            stepClass += " step_error"
            statusIcon = "❌"
        default:
            stepClass += " step_inactive"
            statusIcon = "⚪"
        }
        
        return b.Div(mi.Class(stepClass), mi.DataAttr("step", stepID),
            b.Div(mi.Class("step_header"),
                b.Span(mi.Class("step_icon"), statusIcon),
                b.H4(mi.Class("step_title"), title),
                b.Span(mi.Class("step_status"), strings.Title(status)),
            ),
            b.Div(mi.Class("step_content"),
                content(b),
            ),
        )
    }
}
```

---

## Component Design Patterns

### Atomic Design Pattern

Building UI components from simple to complex:

```go
// Atoms - Basic building blocks
func StatusBadge(theme mui.Theme, status, statusClass string) mi.H {
    return theme.Badge(status, statusClass)
}

func ActionButton(theme mui.Theme, text, action, variant string) mi.H {
    return mui.DomainButton(theme, "shared", text, variant,
        mi.OnClick(action))
}

// Molecules - Simple component combinations
func MetricCard(theme mui.Theme, label, value, description string) mi.H {
    return theme.Card(label, func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("metric_card_content"),
            b.Div(mi.Class("metric_value"), value),
            b.Div(mi.Class("metric_description"), description),
        )
    })
}

func DataRow(theme mui.Theme, label, value string, actions []mi.H) mi.H {
    return func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("data_row"),
            b.Div(mi.Class("data_label"), label),
            b.Div(mi.Class("data_value"), value),
            b.Div(mi.Class("data_actions"),
                miex.Map(actions, func(action mi.H) mi.Node {
                    return action(b)
                })...,
            ),
        )
    }
}

// Organisms - Complex component compositions
func DataTable(theme mui.Theme, title string, headers []string, 
    rows [][]string, actions TableActions) mi.H {
    
    return theme.Card(title, func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("data_table_container"),
            miex.If(actions.ShowFilters,
                TableFilters(theme, actions.Filters),
            ),
            
            theme.Table(headers, rows),
            
            miex.If(actions.ShowPagination,
                theme.Pagination(actions.CurrentPage, actions.TotalPages, 
                    actions.BaseURL),
            ),
        )
    })
}

// Templates - Page-level layouts
func StandardPage(theme mui.Theme, title string, breadcrumbs []mui.BreadcrumbItem,
    content mi.H, sidebar mi.H) mi.H {
    
    return theme.Container(func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("standard_page_layout"),
            b.Header(mi.Class("page_header"),
                b.H1(mi.Class("page_title"), title),
                theme.Breadcrumbs(breadcrumbs),
            ),
            
            b.Main(mi.Class("page_main"),
                b.Div(mi.Class("page_content"),
                    content(b),
                ),
                miex.If(sidebar != nil,
                    b.Aside(mi.Class("page_sidebar"),
                        sidebar(b),
                    ),
                ),
            ),
        )
    })
}
```

### Conditional Rendering Patterns

```go
// Permission-based rendering
func ConditionalContent(theme mui.Theme, user User, requiredRole string, 
    content mi.H, fallback mi.H) mi.H {
    
    if hasRole(user, requiredRole) {
        return content
    }
    return fallback
}

// Data-dependent rendering
func DataDependentComponent(theme mui.Theme, data interface{}, 
    renderData func(interface{}) mi.H, 
    renderEmpty func() mi.H,
    renderError func(error) mi.H) mi.H {
    
    switch v := data.(type) {
    case error:
        return renderError(v)
    case nil:
        return renderEmpty()
    default:
        return renderData(v)
    }
}

// Status-based rendering
func StatusRenderer(theme mui.Theme, status string, 
    renderers map[string]func() mi.H,
    defaultRenderer func() mi.H) mi.H {
    
    if renderer, exists := renderers[status]; exists {
        return renderer()
    }
    return defaultRenderer()
}
```

---

## Testing Presentation Layer

### Unit Testing UI Components

```go
func TestAccountSummaryCard(t *testing.T) {
    // Mock theme for testing
    mockTheme := &MockTheme{}
    
    account := mifi.Account{
        ID:      "acc_123",
        Name:    "Test Checking",
        Balance: mintyex.NewMoney(1500.00, "USD"),
        Status:  mintyex.StatusActive,
        Type:    "checking",
    }
    
    // Test component generation
    component := mintyfinui.AccountSummaryCard(mockTheme, account)
    html := mi.Render(component)
    
    // Verify content
    assert.Contains(t, html, "Test Checking")
    assert.Contains(t, html, "$1500.00")
    assert.Contains(t, html, "acc_123")
    
    // Verify CSS classes
    assert.Contains(t, html, "mifi_account_summary")
    assert.Contains(t, html, "mifi_account_info")
}

func TestDataPreparation(t *testing.T) {
    account := mifi.Account{
        Name:      "Savings Account",
        Balance:   mintyex.NewMoney(-100.00, "USD"),
        Status:    mintyex.StatusActive,
        Type:      "savings",
        CreatedAt: time.Now().AddDate(0, 0, -30),
    }
    
    displayData := mifi.PrepareAccountForDisplay(account)
    
    assert.Equal(t, "Savings Account", displayData.Name)
    assert.Equal(t, "-$100.00", displayData.FormattedBalance)
    assert.Equal(t, "Active", displayData.StatusDisplay)
    assert.Equal(t, "Savings Account", displayData.TypeDisplay)
    assert.True(t, displayData.IsOverdrawn)
    assert.Equal(t, 30, displayData.DaysOld)
}
```

### Integration Testing with Multiple Themes

```go
func TestMultiThemeCompatibility(t *testing.T) {
    account := mifi.Account{
        Name:    "Test Account",
        Balance: mintyex.NewMoney(1000.00, "USD"),
        Status:  mintyex.StatusActive,
        Type:    "checking",
    }
    
    themes := []struct {
        name  string
        theme mui.Theme
    }{
        {"Bootstrap", bootstrap.NewBootstrapTheme()},
        {"Tailwind", tailwind.NewTailwindTheme()},
        {"Custom", NewCustomTheme()},
    }
    
    for _, tt := range themes {
        t.Run(tt.name, func(t *testing.T) {
            component := mintyfinui.AccountSummaryCard(tt.theme, account)
            html := mi.Render(component)
            
            // Essential content should be present regardless of theme
            assert.Contains(t, html, "Test Account")
            assert.Contains(t, html, "$1000.00")
            
            // Should not contain errors or empty content
            assert.NotContains(t, html, "undefined")
            assert.NotContains(t, html, "null")
            assert.NotEmpty(t, html)
        })
    }
}
```

### Mock Theme for Testing

```go
type MockTheme struct{}

func (t *MockTheme) GetName() string    { return "Mock" }
func (t *MockTheme) GetVersion() string { return "1.0.0" }

func (t *MockTheme) Button(text, variant string, attrs ...mi.Attribute) mi.H {
    return func(b *mi.Builder) mi.Node {
        return b.Button(mi.Class("mock-button"), text)
    }
}

func (t *MockTheme) Card(title string, content mi.H) mi.H {
    return func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("mock-card"),
            miex.If(title != "", b.H3(title)),
            content(b),
        )
    }
}

func (t *MockTheme) Table(headers []string, rows [][]string) mi.H {
    return func(b *mi.Builder) mi.Node {
        return b.Table(mi.Class("mock-table"),
            b.Thead(
                b.Tr(
                    miex.Map(headers, func(h string) mi.Node {
                        return b.Th(h)
                    })...,
                ),
            ),
            b.Tbody(
                miex.Map(rows, func(row []string) mi.Node {
                    return b.Tr(
                        miex.Map(row, func(cell string) mi.Node {
                            return b.Td(cell)
                        })...,
                    )
                })...,
            ),
        )
    }
}
```

---

## Advanced Presentation Patterns

### Lazy Loading Components

```go
func LazyLoadingList(theme mui.Theme, initialItems []interface{}, 
    loadMoreURL string, itemRenderer func(interface{}) mi.H) mi.H {
    
    return func(b *mi.Builder) mi.Node {
        return b.Div(mi.Class("lazy_loading_list"),
            mi.DataAttr("load-more-url", loadMoreURL),
            
            b.Div(mi.Class("list_items"),
                miex.Map(initialItems, func(item interface{}) mi.H {
                    return itemRenderer(item)
                })...,
            ),
            
            b.Div(mi.Class("load_more_container"),
                theme.Button("Load More", "secondary",
                    mi.Class("load_more_button"),
                    mi.OnClick("loadMoreItems(this)")),
                
                b.Div(mi.Class("loading_spinner"), 
                    mi.Style("display: none;"),
                    "Loading..."),
            ),
        )
    }
}
```

### Responsive Component Patterns

```go
func ResponsiveGrid(theme mui.Theme, items []interface{}, 
    itemRenderer func(interface{}) mi.H,
    breakpoints map[string]int) mi.H {
    
    return func(b *mi.Builder) mi.Node {
        gridClass := "responsive_grid"
        for breakpoint, columns := range breakpoints {
            gridClass += fmt.Sprintf(" %s-cols-%d", breakpoint, columns)
        }
        
        return b.Div(mi.Class(gridClass),
            miex.Map(items, func(item interface{}) mi.H {
                return func(b *mi.Builder) mi.Node {
                    return b.Div(mi.Class("grid_item"),
                        itemRenderer(item)(b),
                    )
                }
            })...,
        )
    }
}

// Usage
func ProductCatalog(theme mui.Theme, products []mica.Product) mi.H {
    return ResponsiveGrid(theme, 
        miex.Map(products, func(p mica.Product) interface{} { return p }),
        func(item interface{}) mi.H {
            product := item.(mica.Product)
            return ProductCard(theme, product)
        },
        map[string]int{
            "sm": 1,  // 1 column on small screens
            "md": 2,  // 2 columns on medium screens
            "lg": 3,  // 3 columns on large screens
            "xl": 4,  // 4 columns on extra large screens
        })
}
```

### Error Boundary Pattern

```go
func ErrorBoundary(theme mui.Theme, content mi.H, 
    onError func(error) mi.H) mi.H {
    
    return func(b *mi.Builder) mi.Node {
        // In a real implementation, this would use error recovery
        // For now, we'll demonstrate the pattern
        defer func() {
            if r := recover(); r != nil {
                // Handle panics in component rendering
                err := fmt.Errorf("component error: %v", r)
                return onError(err)(b)
            }
        }()
        
        return content(b)
    }
}

func DefaultErrorHandler(theme mui.Theme) func(error) mi.H {
    return func(err error) mi.H {
        return theme.Card("Error", func(b *mi.Builder) mi.Node {
            return b.Div(mi.Class("error_boundary"),
                b.P("An error occurred while rendering this component:"),
                b.Code(err.Error()),
                theme.Button("Reload", "primary", 
                    mi.OnClick("window.location.reload()")),
            )
        })
    }
}
```

---

## Summary

The Minty System's presentation layer architecture provides:

**Clean Separation**: Complete independence between business logic and UI concerns, enabling maintainable and testable applications.

**Data Transformation**: Systematic conversion of business entities to display-ready structures with formatting and computed values.

**Component Composition**: Rich set of UI components that can be composed to build complex interfaces from simple building blocks.

**Theme Integration**: Seamless integration with multiple theme systems while maintaining consistent component APIs.

**Cross-Domain Support**: Unified interfaces that span multiple business domains while maintaining clear architectural boundaries.

**Testing Excellence**: Clear testing strategies for both individual components and cross-theme compatibility.

The presentation layer serves as the crucial bridge between the Minty System's powerful business domains and its flexible theme system, enabling the development of sophisticated user interfaces that remain maintainable and adaptable over time.

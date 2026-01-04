# InsureQuote

A comprehensive insurance quote application demonstrating all mintydyn patterns with SVG icons and Tailwind CSS.

## Quick Start

```bash
# Run the application
make run

# Or with Go directly
go run ./cmd/insurequote

# Open http://localhost:8080
```

## Building

```bash
# Build binary
make build
./bin/insurequote

# Build for specific platform
make build-linux
make build-darwin
make build-windows

# Build all platforms
make build-all
```

## Configuration

Set the `PORT` environment variable to change the listening port (default: 8080):

```bash
PORT=3000 make run
```

## Project Structure

```
insurance-quote/
├── cmd/
│   └── insurequote/
│       └── main.go           # Entry point
├── internal/
│   ├── models/
│   │   └── models.go         # Data structures
│   ├── store/
│   │   └── store.go          # Sample data
│   └── ui/
│       ├── handlers.go       # HTTP handlers + mintydyn components
│       └── icons.go          # Heroicons SVG system
├── minty/                    # minty framework (local copy)
├── configs/                  # Configuration files
├── go.mod
├── Makefile
└── README.md
```

## Purpose

This application serves as a "kitchen sink" example demonstrating all mintydyn patterns:

| Pattern | Location | Description |
|---------|----------|-------------|
| **States** | Quote wizard, Settings tabs, Compare plans | Tab-like navigation with content panels |
| **Rules** | Quote form | Conditional field visibility based on selections |
| **ClientFilterable** | Claims page | JSON data with client-side filtering |

## Pages

### Dashboard (`/`)
Overview with stats cards, quick actions, and coverage type selection.

### Get Quote (`/quote`)
**Demonstrates: States + Rules**

- **States pattern**: Wizard steps (Coverage → Details → Customize → Review)
- **Rules pattern**: Form fields show/hide based on coverage type
  - Auto: Vehicle info, accident history
  - Home: Property info, pool coverage
  - Life: Health info, smoker notice
  - Business: Company info, premises fields

### Claims (`/claims`)
**Demonstrates: ClientFilterable (Data pattern)**

- JSON data array passed to component
- Client-side filtering by customer, status, type
- Pagination support

### Compare Plans (`/compare`)
**Demonstrates: States with pre-rendered content**

- Tabs for each coverage type (All, Auto, Home, Life, Business)
- SVG icons in tab labels

### Settings (`/settings`)
**Demonstrates: States**

- Profile, Notifications, Security, Billing tabs
- Toggle switches for notification preferences

## Icon System

Uses Heroicons (MIT licensed) via embedded SVG paths:

```go
// As minty Node (in HTML generation)
Icon("shield-check", "w-5 h-5 text-blue-600")

// As HTML string (for mintydyn state icons)
IconHTML("truck", "w-4 h-4 inline")
```

Available icons include: `home`, `shield-check`, `truck`, `heart`, `building-office`, `plus`, `check`, `x-mark`, `pencil`, `trash`, `arrow-right`, `check-circle`, `exclamation-triangle`, `bell`, `user`, `calendar`, and more.

## Dark Mode

Uses minty's built-in `DarkModeTailwind()`:

```go
var darkMode = mi.DarkModeTailwind(mi.DarkModeMinify())

// In <head>:
darkMode.Script(b)

// Toggle button:
darkMode.Toggle(b, mi.Class("..."))
```

Dark mode state persists in localStorage and is applied immediately on page load (no flash).

## Key Code Examples

### Rules Pattern (Form Dependencies)

```go
formRules := mdy.Dyn("quote-form-rules").
    Rules([]mdy.DependencyRule{
        mdy.ShowWhen("coverage-type", "equals", "auto", "auto-fields"),
        mdy.ShowWhen("coverage-type", "equals", "home", "home-fields"),
        mdy.ShowWhen("has-accidents", "equals", true, "accident-details"),
    }).
    Theme(h.theme).
    Build()
```

### Client-Side Filtering

```go
claimsFilter := mdy.Dyn("claims-filter").
    Data(mdy.FilterableDataset{
        Items: h.store.ClaimsAsMapSlice(),
        Schema: mdy.FilterSchema{
            Fields: []mdy.FilterableField{
                mdy.TextField("customerName", "Customer"),
                mdy.SelectField("status", "Status", []string{"open", "in-progress", "approved"}),
            },
        },
        Options: mdy.FilterOptions{
            EnableSearch:     true,
            EnablePagination: true,
            ItemsPerPage:     5,
        },
    }).
    Theme(h.theme).
    Build()
```

### States with Icons

```go
states := []mdy.ComponentState{
    {ID: "all", Label: "All Plans", Active: true, Content: h.planGrid(b, "")},
    {ID: "auto", Label: "Auto", Icon: IconHTML("truck", "w-4 h-4 inline"), Content: h.planGrid(b, "auto")},
}
```

## Comparison with AssetTrack

| Feature | AssetTrack | InsureQuote |
|---------|------------|-------------|
| States (tabs) | ✓ | ✓ |
| ServerRenderedFilter | ✓ | — |
| ClientFilterable | — | ✓ |
| Rules (dependencies) | — | ✓ |
| SVG icons | — | ✓ |
| Emoji icons | ✓ | — |

Together, both examples demonstrate the full range of mintydyn capabilities.

## License

MIT

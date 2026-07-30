# Minty

**Minty is a fresh, ultra-concise, type-safe HTML generation library for Go web applications. It provides a fluent API for creating HTML components without traditional templates, offering compile-time safety and excellent IDE support.**

## Features

- **Type-safe HTML generation** - Catch errors at compile time, not runtime
- **Fluent builder pattern** - Intuitive, chainable API
- **Dynamic client-side components** - Without writing any Javascript by hand
- **HTMX integration** - First-class support for HTMX attributes and patterns
- **Theme system** - Pluggable themes (Bootstrap, Tailwind, Bulma, Material Design)
- **Domain libraries** - Pre-built components for common business domains
- **Control flow helpers** - If, IfElse, Each, Map, Filter, and more
- **Zero dependencies** - Pure Go, no external requirements

## What's new in 0.2.0 and 0.2.1

Type-safe conditional attributes (`IfT`/`IfElseT`), composite conditions
in `mintydyn` (`allOf`/`anyOf`), a `Validate()` pre-flight check for
config-level mistakes, `Builder.Debug` for informative panics, and two
fixes from [@mogsie](https://github.com/mogsie) (deterministic attribute
order; a clearer failure on a common mistake). The new conditional-
attribute path also measures 10-15% faster than the equivalent existing
API, and 4-5x faster than an earlier explored approach. 0.2.1 adds
`in`/`notIn` list-membership operators to `mintydyn`'s conditions.

Full details, measured performance data, and credits:
**[What's New in 0.2.0 and 0.2.1](docs/minty-12-whats-new-0.2.0.md)**.

If you're deciding how much to trust type safety here, or debugging
something that compiled fine but behaved unexpectedly:
**[Type Safety: What It Covers, and What It Doesn't](docs/minty-13-type-safety.md)**.

## Installation

```bash
go get github.com/ha1tch/minty
```

## Quick Start

```go
package main

import (
    "os"
    mi "github.com/ha1tch/minty"
)

func main() {
    // Create a simple page
    page := func(b *mi.Builder) mi.Node {
        return b.Html(
            b.Head(b.Title("Hello Minty")),
            b.Body(
                b.H1(mi.Class("title"), "Welcome!"),
                b.P("This is a paragraph."),
                b.A(mi.Href("/about"), "Learn more"),
            ),
        )
    }
    
    // Render to stdout
    mi.Render(page, os.Stdout)
}
```

## Package Structure

```
github.com/ha1tch/minty
├── /                    # Core library (HTML builder, attributes, HTMX)
├── mintytypes/          # Pure business types (Money, Address, Status, etc.)
├── mintyex/             # Extensions (UI helpers, re-exports mintytypes)  
├── mintyui/             # UI component abstractions (Theme interface)
├── mintydyn/            # Dynamic client-side components (states, data, rules)
├── domains/             # Business domain libraries (depend only on mintytypes)
│   ├── mintyfin/        # Finance domain (accounts, transactions, invoices)
│   ├── mintycart/       # E-commerce domain (products, carts, orders)
│   └── mintymove/       # Logistics domain (shipments, tracking, vehicles)
├── presentation/        # UI adapters (domain → themed components)
│   ├── mintyfinui/
│   ├── mintycartui/
│   └── mintymoveui/
├── themes/              # Theme implementations
│   ├── bootstrap/       # Bootstrap 5 theme
│   ├── tailwind/        # Tailwind CSS theme
│   ├── bulma/           # Bulma CSS theme
│   └── material/        # Material Design theme
├── examples/            # Example applications
└── docs/                # Comprehensive documentation
```

### Clean Architecture

The dependency graph enforces clean architecture:

```
        mintytypes (pure - no dependencies)
        ↗          ↖
    minty           domains/*
        ↖          ↗
         mintyex
            ↑
      presentation/*
```

- **mintytypes**: Pure business types with zero external dependencies
- **domains**: Business logic depends only on mintytypes (no UI knowledge)
- **mintyex**: UI helpers + re-exports mintytypes for convenience
- **presentation**: Adapts domain data to themed UI components

## Core Concepts

### The Builder Pattern

All HTML elements are created through the `Builder` type. Use the global `B` instance or create your own:

```go
import mi "github.com/ha1tch/minty"

// Using global builder
div := mi.B.Div(mi.Class("container"), "Hello")

// Using builder in templates (most common)
template := func(b *mi.Builder) mi.Node {
    return b.Div(mi.Class("container"),
        b.H1("Title"),
        b.P("Content"),
    )
}
```

### Attributes

Attributes are created using helper functions:

```go
b.A(
    mi.Href("/page"),
    mi.Class("nav-link"),
    mi.Target("_blank"),
    mi.Rel("noopener"),
    "Click me",
)
```

### HTMX Integration

First-class HTMX support:

```go
b.Button(
    mi.Class("btn"),
    mi.HtmxPost("/api/submit"),
    mi.HtmxTarget("#result"),
    mi.HtmxSwap("innerHTML"),
    mi.HtmxIndicator("#spinner"),
    "Submit",
)
```

### Control Flow

Conditional rendering:

```go
mi.If(isLoggedIn, userGreeting)
mi.IfElse(hasItems, itemsList, emptyMessage)
mi.Each(items, func(item Item) mi.H {
    return func(b *mi.Builder) mi.Node {
        return b.Li(item.Name)
    }
})
```

Type-safe conditional attributes, for self-closing elements that only
accept `...Attribute` (like `Input`) -- see
[What's New in 0.2.0](docs/minty-12-whats-new-0.2.0.md#type-safe-conditional-attributes-ift-and-ifelset)
for nesting and the full story:

```go
b.Input(mi.IfT(maybe, mi.Class("highlighted")))
b.Input(mi.IfElseT(isValid, mi.Class("valid"), mi.Class("invalid")))
```

## Themes

Use pre-built themes for consistent styling:

```go
import (
    mi "github.com/ha1tch/minty"
    "github.com/ha1tch/minty/themes/bootstrap"
)

theme := bootstrap.NewBootstrapTheme()

// Use themed components
button := theme.Button("Click me", "primary")
card := theme.Card("Title", content)
form := theme.FormInput("Email", "email", "email")
```

## Documentation

Comprehensive documentation is available in the `/docs` directory:

1. [Introduction](docs/minty-01-intro.md) - Getting started
2. [Design Philosophy](docs/minty-02-design.md) - Architecture decisions
3. [Architecture](docs/minty-03-architecture.md) - System design
4. [Syntax & API](docs/minty-04-syntax-and-api.md) - Complete API reference
5. [Components](docs/minty-05-components.md) - Component library
6. [HTMX Integration](docs/minty-06-htmx-integration.md) - HTMX patterns
7. [Business Domains](docs/minty-07-business-domains.md) - Domain libraries
8. [Iterators](docs/minty-08-iterators.md) - Collection utilities
9. [Themes](docs/minty-09-themes.md) - Theme system
10. [Presentation Layer](docs/minty-10-presentation-layer.md) - UI patterns
11. [JavaScript Integration](docs/minty-11-javascript-integration.md) - JS interop
12. [What's New in 0.2.0 and 0.2.1](docs/minty-12-whats-new-0.2.0.md) - New APIs, measured performance, credits
13. [Type Safety](docs/minty-13-type-safety.md) - What's covered, what isn't, and why

## Examples

Run the simple example:

```bash
cd examples/simple
go run main.go
```

## Development Status

All packages compile and tests pass.

### Stable (Ready for use)
- Core library (HTML builder, attributes)
- HTMX integration
- Theme system (Bootstrap, Tailwind, Bulma, Material)
- Domain libraries (mintyfin, mintycart, mintymove)
- Extensions (mintyex)
- UI abstractions (mintyui)
- Dynamic components (mintydyn) -- states, filterable data, dependency rules, including composite (allOf/anyOf) conditions
- Presentation adapters (mintycartui, mintyfinui, mintymoveui)

## Standard Import Aliases

For consistency across the codebase, use these import aliases:

```go
import (
    mi   "github.com/ha1tch/minty"           // Core library
    mt   "github.com/ha1tch/minty/mintytypes" // Pure business types
    miex "github.com/ha1tch/minty/mintyex"   // Extensions (includes mt re-exports)
    mui  "github.com/ha1tch/minty/mintyui"   // UI components
    
    // Domain packages (import mt, not miex)
    mifi "github.com/ha1tch/minty/domains/mintyfin"   // Finance
    mica "github.com/ha1tch/minty/domains/mintycart"  // E-commerce
    mimo "github.com/ha1tch/minty/domains/mintymove"  // Logistics
)
```

**Note**: Domain packages import `mintytypes` directly (as `mt`) to maintain clean architecture. Presentation layers can import `mintyex` which re-exports all types for convenience.

## License

Apache 2.0

https://github.com/ha1tch/minty?tab=Apache-2.0-1-ov-file

## Author

Copyright (C)2026 haitch

h@ual.li 

https://oldbytes.space/@haitchfive 


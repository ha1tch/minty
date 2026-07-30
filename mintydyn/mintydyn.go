// Package mintydyn provides dynamic UI components with automatic pattern detection.
// It generates minimal, targeted JavaScript for client-side interactivity while
// maintaining server-side rendering as the source of truth.
//
// Recommended import alias:
//
//	import mdy "github.com/ha1tch/minty/mintydyn"
//
// The package supports three primary patterns that can be combined:
//   - States: Tab-like interfaces with pre-rendered or dynamic content
//   - Data: Filterable datasets with client-side or server-side filtering
//   - Rules: Dependency management for form fields and UI elements
//
// Pattern detection is automatic based on what data you provide.
package mintydyn

import (
	"encoding/json"
)

// =============================================================================
// GENERIC TYPE CONSTRAINTS
// =============================================================================

// States constraint defines valid types for state/tab data.
// Accepts slices, maps, or rich collections of ComponentState.
type States interface {
	~[]ComponentState | ~map[string]ComponentState | ComponentStateCollection
}

// Data constraint defines valid types for filterable data.
// Accepts various slice types or structured datasets.
type Data interface {
	~[]map[string]interface{} | ~[]interface{} | FilterableDataset | DataCollection
}

// Rules constraint defines valid types for dependency rules.
// Accepts slices or rich collections of DependencyRule.
type Rules interface {
	~[]DependencyRule | RuleCollection
}

// =============================================================================
// CORE DATA STRUCTURES
// =============================================================================

// ComponentState represents a single state (tab, view, panel) in a dynamic component.
type ComponentState struct {
	ID        string                 `json:"id"`
	Label     string                 `json:"label,omitempty"`
	Content   interface{}            `json:"content,omitempty"` // Can be string, mi.H, or mi.Node
	Active    bool                   `json:"active"`
	Disabled  bool                   `json:"disabled,omitempty"`
	Icon      string                 `json:"icon,omitempty"`
	Condition *StateCondition        `json:"condition,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// StateCondition defines when a state should be available or visible.
//
// AllOf/AnyOf make this a composite condition tree: when either is set,
// Field/Operator/Value/Component on this node are ignored and every entry
// in AllOf must be true (AND) or at least one entry in AnyOf must be true
// (OR) instead. Composites nest arbitrarily -- an AllOf entry can itself
// have its own AllOf/AnyOf. Every existing StateCondition{Field: ...} value
// is a valid leaf under this shape, so this is purely additive: no current
// caller or serialized config changes meaning.
type StateCondition struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"` // equals, notEquals, contains, greaterThan, lessThan
	Value     interface{} `json:"value"`
	Component string      `json:"component,omitempty"` // External component reference

	AllOf []StateCondition `json:"allOf,omitempty"`
	AnyOf []StateCondition `json:"anyOf,omitempty"`
}

// ComponentStateCollection provides rich state management with transitions.
type ComponentStateCollection struct {
	States      []ComponentState       `json:"states"`
	DefaultID   string                 `json:"defaultId,omitempty"`
	Transitions map[string][]string    `json:"transitions,omitempty"` // Valid state transitions
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// =============================================================================
// DEPENDENCY RULES
// =============================================================================

// DependencyRule defines how UI elements respond to changes in other elements.
type DependencyRule struct {
	ID          string             `json:"id"`
	Priority    int                `json:"priority,omitempty"` // Higher priority wins conflicts
	Trigger     TriggerCondition   `json:"trigger"`
	Actions     []DependencyAction `json:"actions"`
	Description string             `json:"description,omitempty"`
}

// TriggerCondition specifies when a rule should fire.
//
// AllOf/AnyOf make this a composite condition tree: when either is set,
// ComponentID/Condition/Value on this node are ignored and every entry in
// AllOf must be true (AND) or at least one entry in AnyOf must be true (OR)
// instead. Composites nest arbitrarily. Event/Debounce stay meaningful on
// the top-level rule regardless of whether Trigger is a leaf or a
// composite, since they govern when re-evaluation happens, not what's
// evaluated. Every existing TriggerCondition{ComponentID: ...} value is a
// valid leaf under this shape, so this is purely additive: no current
// caller or serialized config changes meaning.
type TriggerCondition struct {
	ComponentID string      `json:"componentId"`
	Event       string      `json:"event"`     // change, click, focus, blur
	Condition   string      `json:"condition"` // equals, notEquals, contains, greaterThan, lessThan, checked, unchecked, empty, notEmpty
	Value       interface{} `json:"value"`
	Debounce    int         `json:"debounce,omitempty"` // Milliseconds

	AllOf []TriggerCondition `json:"allOf,omitempty"`
	AnyOf []TriggerCondition `json:"anyOf,omitempty"`
}

// DependencyAction specifies what happens when a rule fires.
type DependencyAction struct {
	TargetID  string      `json:"targetId"`
	Action    string      `json:"action"` // show, hide, enable, disable, addClass, removeClass, setValue, setText, setHTML, focus, blur
	Value     interface{} `json:"value,omitempty"`
	Condition string      `json:"condition,omitempty"` // Additional condition for action
}

// RuleCollection provides rich rule management with grouping.
type RuleCollection struct {
	Rules      []DependencyRule       `json:"rules"`
	Groups     map[string][]string    `json:"groups,omitempty"` // Rule groupings
	Priorities map[string]int         `json:"priorities,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// =============================================================================
// FILTERABLE DATA
// =============================================================================

// FilterableDataset combines data with schema and options for filtering.
type FilterableDataset struct {
	Items   []map[string]interface{} `json:"items"`
	Schema  FilterSchema             `json:"schema"`
	Options FilterOptions            `json:"options"`
}

// DataCollection is a simpler data container with optional schema.
type DataCollection struct {
	Items    []map[string]interface{} `json:"items"`
	Schema   FilterSchema             `json:"schema,omitempty"`
	Metadata map[string]interface{}   `json:"metadata,omitempty"`
}

// FilterSchema defines the filterable fields in a dataset.
type FilterSchema struct {
	Fields []FilterableField `json:"fields"`
}

// FilterableField describes a single filterable field.
type FilterableField struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // text, range, multiselect, boolean, select
	Label        string      `json:"label"`
	Options      []string    `json:"options,omitempty"` // For select/multiselect
	Range        *RangeInfo  `json:"range,omitempty"`   // For range type
	Searchable   bool        `json:"searchable,omitempty"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
}

// RangeInfo defines min/max/step for range filters.
type RangeInfo struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Step float64 `json:"step"`
}

// FilterOptions controls filtering behavior.
type FilterOptions struct {
	EnableSearch     bool   `json:"enableSearch"`
	EnableSort       bool   `json:"enableSort"`
	ItemsPerPage     int    `json:"itemsPerPage"`
	EnablePagination bool   `json:"enablePagination"`
	ClientSide       bool   `json:"clientSide"`             // Force client-side even for large datasets
	ServerRendered   bool   `json:"serverRendered"`         // Data is pre-rendered in HTML, just show/hide
	RowSelector      string `json:"rowSelector"`            // CSS selector for data rows (e.g., ".asset-row")
	CounterSelector  string `json:"counterSelector"`        // CSS selector for count display (e.g., "#asset-count")
	ItemTemplate     string `json:"itemTemplate,omitempty"` // JS template for rendering items (uses ${field} syntax)
}

// =============================================================================
// COMPONENT OPTIONS
// =============================================================================

// DynamicOptions configures component behavior.
type DynamicOptions struct {
	// Size limits for pattern detection
	MaxStates   int `json:"maxStates"`
	MaxDataSize int `json:"maxDataSize"`

	// Pattern control
	FallbackPattern  string `json:"fallbackPattern"`
	StrictValidation bool   `json:"strictValidation"`
	PerformanceMode  string `json:"performanceMode"` // speed, memory, balanced

	// JavaScript output
	MinifyJS bool `json:"minifyJs,omitempty"` // Minify generated JavaScript

	// Custom attributes for container
	CustomAttributes map[string]string `json:"customAttributes,omitempty"`

	// External library integration
	ExternalScripts  []ExternalScript `json:"externalScripts,omitempty"`
	ExternalRegistry []string         `json:"externalRegistry,omitempty"` // Names to reserve in this.externals

	// Lifecycle hooks
	Hooks ComponentHooks `json:"hooks,omitempty"`

	// General metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ExternalScript defines an external JavaScript dependency.
type ExternalScript struct {
	Src      string `json:"src"`
	Async    bool   `json:"async,omitempty"`
	Defer    bool   `json:"defer,omitempty"`
	OnLoad   string `json:"onLoad,omitempty"`   // JS code to run when loaded
	Required bool   `json:"required,omitempty"` // Block component init until loaded
}

// ComponentHooks provides lifecycle callbacks.
type ComponentHooks struct {
	BeforeInit        string            `json:"beforeInit,omitempty"`
	AfterInit         string            `json:"afterInit,omitempty"`
	BeforeStateChange string            `json:"beforeStateChange,omitempty"` // Receives {from, to}, return false to cancel
	AfterStateChange  string            `json:"afterStateChange,omitempty"`  // Receives {from, to}
	BeforeFilter      string            `json:"beforeFilter,omitempty"`      // Receives {field, value}
	AfterFilter       string            `json:"afterFilter,omitempty"`       // Receives {field, value, resultCount}
	OnDestroy         string            `json:"onDestroy,omitempty"`
	StateHooks        map[string]string `json:"stateHooks,omitempty"` // Per-state callbacks: stateID -> JS code
}

// =============================================================================
// PATTERN DETECTION
// =============================================================================

// DetectedPattern contains the result of automatic pattern detection.
type DetectedPattern struct {
	HasStates      bool     `json:"hasStates"`
	HasData        bool     `json:"hasData"`
	HasRules       bool     `json:"hasRules"`
	HasRenderer    bool     `json:"hasRenderer"`
	PrimaryPattern string   `json:"primaryPattern"`
	StateCount     int      `json:"stateCount"`
	DataSize       int      `json:"dataSize"`
	RuleCount      int      `json:"ruleCount"`
	Optimizations  []string `json:"optimizations"`
}

// Pattern constants
const (
	PatternEmpty             = "empty"
	PatternPreRenderedStates = "pre-rendered-states"
	PatternDynamicStates     = "dynamic-states"
	PatternClientFilterable  = "client-filterable"
	PatternServerFilterable  = "server-filterable"
	PatternDependencyOnly    = "dependency-only"
	PatternStatefulData      = "stateful-data"
	PatternFilterableStates  = "filterable-states"
	PatternDependentStates   = "dependent-states"
	PatternDependentData     = "dependent-data"
	PatternComplete          = "complete"
)

// =============================================================================
// COMPONENT RENDERER
// =============================================================================

// ComponentRenderer defines how to render individual data items.
// Used with filterable data patterns.
type ComponentRenderer func(item map[string]interface{}) interface{}

// =============================================================================
// JSON HELPERS
// =============================================================================

// MustJSON marshals a value to JSON, panicking on error.
// Useful for embedding config in generated scripts.
func MustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic("mintydyn: failed to marshal JSON: " + err.Error())
	}
	return string(b)
}

// JSONOrEmpty marshals a value to JSON, returning empty object on error.
func JSONOrEmpty(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

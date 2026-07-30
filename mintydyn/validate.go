package mintydyn

import (
	"errors"
	"fmt"
)

// =============================================================================
// VALIDATE -- the practical strict-mode equivalent
// =============================================================================
//
// Go's type system cannot reject any of the footguns this catches at
// compile time: an empty []TriggerCondition, an unrecognized operator
// string, or two mutually-exclusive fields both being set are all
// perfectly well-typed Go values. There is no dependent-type or
// refinement-type mechanism in Go to express "this slice must be
// non-empty" or "these two fields are mutually exclusive" as a type
// constraint -- these are constraints on *values*, not *types*, so no
// strict-mode compiler flag could exist for them, regardless of what
// tooling might promise.
//
// Validate is the closest practical equivalent: a Go-callable pre-flight
// check, walking the exact same condition trees the generated JavaScript
// evaluates at runtime, catching the exact same six footguns the
// generated JS's own console.warn guards catch (see javascript.go's
// warnIfMalformedConditionNode) -- but surfaced as a Go error, callable
// from a unit test or a startup check, rather than only visible as a
// browser console.warn that requires someone to have dev tools open, at
// runtime, potentially in production, to ever see. Call it in a test:
//
//	func TestMyRules(t *testing.T) {
//	    if err := builder.Validate(); err != nil {
//	        t.Fatal(err)
//	    }
//	}
//
// Every issue found is fail-safe at runtime regardless (the generated JS
// always resolves to a defined result, never an unhandled exception from
// a malformed condition tree) -- Validate exists to make mistakes visible
// early, not because leaving them unvalidated would crash anything.
func (db *DynamicBuilder[S, D, R]) Validate() error {
	issues := db.collectIssues()
	errs := make([]error, len(issues))
	for i, issue := range issues {
		errs[i] = issue
	}
	return errors.Join(errs...)
}

// IssueKind categorizes what kind of problem an Issue describes.
type IssueKind int

const (
	// IssueUnrecognizedOperator: an operator/condition string that isn't
	// one of the values the generated JS evaluator actually recognizes.
	// Value is the offending string; Valid holds every recognized
	// operator.
	IssueUnrecognizedOperator IssueKind = iota
	// IssueAllOfAndAnyOfBothSet: a condition node has both allOf and
	// anyOf set. allOf silently wins at runtime; anyOf is ignored.
	// Value/Valid are not meaningful for this kind (empty/nil).
	IssueAllOfAndAnyOfBothSet
	// IssueCompositeAndLeafBothSet: a node has both a composite
	// (allOf/anyOf) and leaf fields (field/component/componentId/
	// operator/condition) set. The leaf fields are silently ignored at
	// runtime. Value/Valid are not meaningful for this kind.
	IssueCompositeAndLeafBothSet
	// IssueEmptyAllOf: an allOf slice is present but has zero entries --
	// evaluates as vacuously true (condition met) at runtime.
	// Value/Valid are not meaningful for this kind.
	IssueEmptyAllOf
	// IssueEmptyAnyOf: an anyOf slice is present but has zero entries --
	// evaluates as vacuously false (condition not met) at runtime.
	// Value/Valid are not meaningful for this kind.
	IssueEmptyAnyOf
	// IssueDeadRule: a rule's trigger condition references no
	// componentId anywhere in its tree, so it can never be registered
	// against any trigger and can never fire. Value/Valid are not
	// meaningful for this kind.
	IssueDeadRule
)

func (k IssueKind) String() string {
	switch k {
	case IssueUnrecognizedOperator:
		return "unrecognized_operator"
	case IssueAllOfAndAnyOfBothSet:
		return "allof_and_anyof_both_set"
	case IssueCompositeAndLeafBothSet:
		return "composite_and_leaf_both_set"
	case IssueEmptyAllOf:
		return "empty_allof"
	case IssueEmptyAnyOf:
		return "empty_anyof"
	case IssueDeadRule:
		return "dead_rule"
	default:
		return "unknown"
	}
}

// Issue describes a single detected problem in a condition tree. It
// implements error, so a []Issue can be joined via errors.Join exactly
// as a []error could -- Validate() does exactly this, unchanged in
// behaviour from before this type existed.
//
// Value and Valid are populated only for enum-style issues -- currently
// just IssueUnrecognizedOperator, the one issue kind that is genuinely
// "this value must be one of a known set." The other five kinds are
// structural (two fields both set, a slice being empty, a tree
// referencing nothing) rather than "value not in set X", so forcing a
// Valid list onto them would be inventing information that doesn't
// naturally exist for that kind of mistake.
type Issue struct {
	Kind    IssueKind
	Context string   // e.g. `rule "r1" (allOf[1])`
	Value   string   // the offending value, only for enum-style issues
	Valid   []string // the full set of valid alternatives, only for enum-style issues
}

func (i Issue) Error() string {
	switch i.Kind {
	case IssueUnrecognizedOperator:
		return fmt.Sprintf("%s: unrecognized operator %q -- likely a typo, evaluates as false at runtime with no error", i.Context, i.Value)
	case IssueAllOfAndAnyOfBothSet:
		return fmt.Sprintf("%s: condition has both allOf and anyOf set -- allOf wins, anyOf is ignored entirely", i.Context)
	case IssueCompositeAndLeafBothSet:
		return fmt.Sprintf("%s: condition has both a composite (allOf/anyOf) and leaf fields set -- the leaf fields are ignored entirely", i.Context)
	case IssueEmptyAllOf:
		return fmt.Sprintf("%s: allOf is an empty (non-nil) slice -- this evaluates as condition MET (vacuous AND), likely not intended", i.Context)
	case IssueEmptyAnyOf:
		return fmt.Sprintf("%s: anyOf is an empty (non-nil) slice -- this evaluates as condition NOT met (vacuous OR)", i.Context)
	case IssueDeadRule:
		return fmt.Sprintf("%s: trigger condition references no componentId anywhere -- this rule can never be registered against any trigger and will never fire", i.Context)
	default:
		return fmt.Sprintf("%s: unknown issue", i.Context)
	}
}

// IssueHandler is called once for every Issue found during validation --
// in addition to, not instead of, the aggregated error Validate() and
// ValidateWithHandler() both still return. Receiving the full Valid list
// alongside the offending Value (for enum-style issues) means a handler
// can build its own message, metric, or decision without needing to
// re-derive "what would have been valid here" itself.
type IssueHandler func(Issue)

// ValidateWithHandler runs the exact same checks as Validate, additionally
// calling handler once for every issue found, as each one is discovered.
// handler may be nil, in which case this behaves identically to Validate.
//
//	builder.ValidateWithHandler(func(issue mintydyn.Issue) {
//	    metrics.Inc("mintydyn.validation_issue", "kind", issue.Kind.String())
//	    if issue.Kind == mintydyn.IssueUnrecognizedOperator {
//	        log.Printf("did you mean one of %v instead of %q?", issue.Valid, issue.Value)
//	    }
//	})
func (db *DynamicBuilder[S, D, R]) ValidateWithHandler(handler IssueHandler) error {
	issues := db.collectIssues()
	errs := make([]error, len(issues))
	for i, issue := range issues {
		if handler != nil {
			handler(issue)
		}
		errs[i] = issue
	}
	return errors.Join(errs...)
}

func (db *DynamicBuilder[S, D, R]) collectIssues() []Issue {
	var issues []Issue

	for _, state := range db.extractStates() {
		if state.Condition != nil {
			issues = append(issues, collectStateConditionIssues(*state.Condition, fmt.Sprintf("state %q", state.ID))...)
		}
	}

	for _, rule := range db.extractRules() {
		triggerIDs := collectTriggerComponentIDs(rule.Trigger)
		if len(triggerIDs) == 0 {
			issues = append(issues, Issue{Kind: IssueDeadRule, Context: fmt.Sprintf("rule %q", rule.ID)})
		}
		issues = append(issues, collectTriggerConditionIssues(rule.Trigger, fmt.Sprintf("rule %q", rule.ID))...)
	}

	return issues
}

func collectStateConditionIssues(c StateCondition, context string) []Issue {
	var issues []Issue
	hasLeaf := c.Field != "" || c.Component != "" || c.Operator != ""

	if len(c.AllOf) > 0 && len(c.AnyOf) > 0 {
		issues = append(issues, Issue{Kind: IssueAllOfAndAnyOfBothSet, Context: context})
	}
	if (len(c.AllOf) > 0 || len(c.AnyOf) > 0) && hasLeaf {
		issues = append(issues, Issue{Kind: IssueCompositeAndLeafBothSet, Context: context})
	}
	if c.AllOf != nil && len(c.AllOf) == 0 {
		issues = append(issues, Issue{Kind: IssueEmptyAllOf, Context: context})
	}
	if c.AnyOf != nil && len(c.AnyOf) == 0 {
		issues = append(issues, Issue{Kind: IssueEmptyAnyOf, Context: context})
	}
	if len(c.AllOf) == 0 && len(c.AnyOf) == 0 && c.Operator != "" && !isKnownOperator(c.Operator) {
		issues = append(issues, Issue{Kind: IssueUnrecognizedOperator, Context: context, Value: c.Operator, Valid: knownOperators})
	}

	for i, sub := range c.AllOf {
		issues = append(issues, collectStateConditionIssues(sub, fmt.Sprintf("%s (allOf[%d])", context, i))...)
	}
	for i, sub := range c.AnyOf {
		issues = append(issues, collectStateConditionIssues(sub, fmt.Sprintf("%s (anyOf[%d])", context, i))...)
	}
	return issues
}

func collectTriggerConditionIssues(c TriggerCondition, context string) []Issue {
	var issues []Issue
	hasLeaf := c.ComponentID != "" || c.Condition != ""

	if len(c.AllOf) > 0 && len(c.AnyOf) > 0 {
		issues = append(issues, Issue{Kind: IssueAllOfAndAnyOfBothSet, Context: context})
	}
	if (len(c.AllOf) > 0 || len(c.AnyOf) > 0) && hasLeaf {
		issues = append(issues, Issue{Kind: IssueCompositeAndLeafBothSet, Context: context})
	}
	if c.AllOf != nil && len(c.AllOf) == 0 {
		issues = append(issues, Issue{Kind: IssueEmptyAllOf, Context: context})
	}
	if c.AnyOf != nil && len(c.AnyOf) == 0 {
		issues = append(issues, Issue{Kind: IssueEmptyAnyOf, Context: context})
	}
	if len(c.AllOf) == 0 && len(c.AnyOf) == 0 && c.Condition != "" && !isKnownOperator(c.Condition) {
		issues = append(issues, Issue{Kind: IssueUnrecognizedOperator, Context: context, Value: c.Condition, Valid: knownOperators})
	}

	for i, sub := range c.AllOf {
		issues = append(issues, collectTriggerConditionIssues(sub, fmt.Sprintf("%s (allOf[%d])", context, i))...)
	}
	for i, sub := range c.AnyOf {
		issues = append(issues, collectTriggerConditionIssues(sub, fmt.Sprintf("%s (anyOf[%d])", context, i))...)
	}
	return issues
}

// collectTriggerComponentIDs mirrors javascript.go's generated
// collectTriggerComponentIds exactly, in Go, over the real
// []TriggerCondition tree rather than the JSON it serializes to -- used
// here to detect a rule that can never fire, the same check
// processRules's own generated-JS warning performs at runtime.
func collectTriggerComponentIDs(node TriggerCondition) []string {
	var out []string
	if len(node.AllOf) > 0 {
		for _, c := range node.AllOf {
			out = append(out, collectTriggerComponentIDs(c)...)
		}
		return out
	}
	if len(node.AnyOf) > 0 {
		for _, c := range node.AnyOf {
			out = append(out, collectTriggerComponentIDs(c)...)
		}
		return out
	}
	if node.ComponentID != "" {
		out = append(out, node.ComponentID)
	}
	return out
}

// knownOperators is the full, valid operator list -- exposed as a slice
// so an Issue describing an unrecognized operator can carry the complete
// set of valid alternatives, not just a pass/fail answer. Kept in sync
// with the generated JavaScript's own switch by hand;
// TestIsKnownOperator_MatchesGeneratedJS cross-checks both this list and
// isKnownOperator itself against the generated source.
var knownOperators = []string{
	"equals", "notEquals", "contains", "greaterThan", "lessThan",
	"checked", "unchecked", "empty", "notEmpty",
}

// isKnownOperator mirrors evaluateConditionOperator_%s's own switch
// exactly -- the set of operators the generated JS actually recognizes.
//
// Dispatches via length first, then a short switch only among the
// candidates that length narrows to, rather than one flat switch
// comparing against all nine unconditionally: six of the nine operators
// are uniquely identified by len(op) alone, before any string content
// comparison happens at all; only length 8 (contains/lessThan/notEmpty)
// and length 9 (notEquals/unchecked) still need a further comparison,
// and even then among at most 3 candidates rather than 9.
//
// Measured, not assumed: BenchmarkIsKnownOperator_NestedSwitch vs
// BenchmarkIsKnownOperator_FlatSwitch shows no measurable difference
// between this and a single flat switch at this scale (9 short strings)
// -- Go's compiler already dispatches a flat string switch efficiently
// on its own, so the manual length-first split doesn't add anything
// further here. What both switch forms measurably beat is a naive linear
// []string scan (BenchmarkIsKnownOperator_LinearScan): roughly 3.5x
// slower on a mixed workload, and about 14x slower specifically for the
// non-matching ("miss") case, where a linear scan must exhaust the whole
// list before concluding "not found," while either switch form resolves
// a miss in near-constant time. The real, measured win here is
// switch-over-linear-scan; the further nested-vs-flat split is kept for
// the documentation value of showing the technique and because it's not
// worse, not because it's faster.
func isKnownOperator(op string) bool {
	switch len(op) {
	case 5:
		return op == "empty"
	case 6:
		return op == "equals"
	case 7:
		return op == "checked"
	case 8:
		switch op {
		case "contains", "lessThan", "notEmpty":
			return true
		default:
			return false
		}
	case 9:
		switch op {
		case "notEquals", "unchecked":
			return true
		default:
			return false
		}
	case 11:
		return op == "greaterThan"
	default:
		return false
	}
}

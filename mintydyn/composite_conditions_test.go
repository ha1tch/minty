package mintydyn

import (
	"encoding/json"
	"strings"
	"testing"
)

// =============================================================================
// Go-level tests: AllOf/AnyOf JSON round-tripping
// =============================================================================
//
// These are genuinely executable Go tests, unlike the generated-JS checks
// below -- the composite condition tree itself is just Go structs with JSON
// tags, so its round-trip correctness is fully verifiable without node,
// unlike the JS evaluator logic that consumes it at runtime.

func TestTriggerCondition_AllOfRoundTrips(t *testing.T) {
	original := TriggerCondition{
		AllOf: []TriggerCondition{
			{ComponentID: "a", Condition: "equals", Value: "x"},
			{ComponentID: "b", Condition: "greaterThan", Value: 5},
		},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var decoded TriggerCondition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(decoded.AllOf) != 2 {
		t.Fatalf("expected 2 AllOf entries, got %d", len(decoded.AllOf))
	}
	if decoded.AllOf[0].ComponentID != "a" || decoded.AllOf[1].ComponentID != "b" {
		t.Errorf("AllOf entries did not round-trip correctly: %+v", decoded.AllOf)
	}
}

func TestTriggerCondition_NestedCompositeRoundTrips(t *testing.T) {
	// Composites nesting arbitrarily is a specific, documented claim --
	// verify a genuinely nested tree (AnyOf containing an AllOf) survives
	// a JSON round trip intact, not just a single flat level.
	original := TriggerCondition{
		AnyOf: []TriggerCondition{
			{ComponentID: "a", Condition: "equals", Value: 1},
			{AllOf: []TriggerCondition{
				{ComponentID: "b", Condition: "equals", Value: 2},
				{ComponentID: "c", Condition: "equals", Value: 3},
			}},
		},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var decoded TriggerCondition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(decoded.AnyOf) != 2 {
		t.Fatalf("expected 2 AnyOf entries, got %d", len(decoded.AnyOf))
	}
	nested := decoded.AnyOf[1]
	if len(nested.AllOf) != 2 || nested.AllOf[0].ComponentID != "b" || nested.AllOf[1].ComponentID != "c" {
		t.Errorf("nested AllOf did not survive the round trip: %+v", nested)
	}
}

func TestTriggerCondition_PlainLeafStillOmitsCompositeFields(t *testing.T) {
	// Every existing TriggerCondition{ComponentID: ...} value is a valid
	// leaf under the new shape -- confirms omitempty means a plain,
	// pre-existing config doesn't grow allOf/anyOf keys in its JSON at
	// all, so this is genuinely additive, not just additive at the Go
	// type level.
	leaf := TriggerCondition{ComponentID: "x", Condition: "equals", Value: "y"}
	data, err := json.Marshal(leaf)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if strings.Contains(string(data), "allOf") || strings.Contains(string(data), "anyOf") {
		t.Errorf("plain leaf's JSON unexpectedly contains allOf/anyOf keys: %s", data)
	}
}

func TestStateCondition_AllOfRoundTrips(t *testing.T) {
	original := StateCondition{
		AllOf: []StateCondition{
			{Field: "a", Operator: "equals", Value: "x"},
			{Field: "b", Operator: "notEquals", Value: "y"},
		},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var decoded StateCondition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(decoded.AllOf) != 2 || decoded.AllOf[0].Field != "a" || decoded.AllOf[1].Field != "b" {
		t.Errorf("StateCondition AllOf did not round-trip correctly: %+v", decoded.AllOf)
	}
}

// =============================================================================
// Generated-JS tests: structural verification only
// =============================================================================
//
// Matches minify_test.go's own established convention: this repository's
// test suite is pure Go, zero external dependencies (the README's own
// stated principle) -- these tests verify the generated JavaScript
// textually/structurally, not by executing it. Full behavioral
// verification (composite allOf/anyOf logic, all six warning guards,
// nesting, multi-component rule registration) was done directly with a
// Node.js harness during development -- documented in CHANGELOG-worthy
// detail in the surrounding code comments -- but is deliberately not part
// of the permanent, node-dependent test suite.

func generatedJS(t *testing.T) string {
	t.Helper()
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test")
	return db.generateBaseClass() + "\n" + db.generateStatesManager() + "\n" + db.generateRulesManager()
}

func TestGeneratedJS_CompositeSupport_StatesManager(t *testing.T) {
	js := generatedJS(t)
	mustContain(t, js, "evaluateCondition(condition)")
	mustContain(t, js, "condition.allOf.every(c => this.evaluateCondition(c))")
	mustContain(t, js, "condition.anyOf.some(c => this.evaluateCondition(c))")
}

func TestGeneratedJS_CompositeSupport_RulesManager(t *testing.T) {
	js := generatedJS(t)
	mustContain(t, js, "evaluateTriggerCondition(trigger, value)")
	mustContain(t, js, "evaluateTriggerConditionNode(node)")
	mustContain(t, js, "trigger.allOf.every(c => this.evaluateTriggerConditionNode(c))")
	mustContain(t, js, "trigger.anyOf.some(c => this.evaluateTriggerConditionNode(c))")
	mustContain(t, js, "node.allOf.every(c => this.evaluateTriggerConditionNode(c))")
	mustContain(t, js, "node.anyOf.some(c => this.evaluateTriggerConditionNode(c))")
}

func TestGeneratedJS_MultiComponentRuleRegistration(t *testing.T) {
	js := generatedJS(t)
	mustContain(t, js, "collectTriggerComponentIds(node, out)")
	mustContain(t, js, "this.collectTriggerComponentIds(rule.trigger, triggerIds)")
}

func TestGeneratedJS_SharedOperatorEvaluator_NoDuplication(t *testing.T) {
	js := generatedJS(t)
	// The whole point of Part 2 of the originating proposal: one shared
	// implementation, not two near-identical switch statements. Assert
	// the de-duplication actually held, not just that it compiles --
	// count occurrences of the operator switch's most distinctive case
	// ('notEmpty', the operator that only ever existed in the
	// trigger-condition-specific switch before unification) and confirm
	// there's exactly one.
	count := strings.Count(js, "case 'notEmpty':")
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of the shared operator switch (case 'notEmpty'), got %d -- de-duplication may have regressed", count)
	}
}

func TestGeneratedJS_AllSixWarningGuardsPresent(t *testing.T) {
	// Six real, confirmed footguns found by direct, adversarial testing
	// during development (not hypothetical): unrecognized operator
	// strings, allOf+anyOf both set, composite+leaf fields both set,
	// empty allOf (vacuous true), empty anyOf (vacuous false), and a
	// rule referencing no componentId anywhere (silently dead, can never
	// fire). Each produces a defined, fail-safe result rather than a
	// crash, but was previously silent -- these guards make the mistake
	// visible via console.warn without changing behavior.
	js := generatedJS(t)
	checks := []struct {
		label string
		want  string
	}{
		{"unrecognized operator", "unrecognized condition operator"},
		{"allOf+anyOf both set", "allOf and anyOf set -- allOf wins"},
		{"composite+leaf fields both set", "composite (allOf/anyOf) and leaf fields"},
		{"empty allOf (vacuous true)", "allOf array is empty -- this evaluates as condition MET"},
		{"empty anyOf (vacuous false)", "anyOf array is empty -- this evaluates as condition NOT met"},
		{"dead rule (no componentId anywhere)", "references no componentId anywhere"},
	}
	for _, c := range checks {
		if !strings.Contains(js, c.want) {
			t.Errorf("missing warning guard for %s: expected generated JS to contain %q", c.label, c.want)
		}
	}
}

func TestGeneratedJS_WarningGuardsWiredIntoAllThreeEntryPoints(t *testing.T) {
	// The guards exist as one shared function (warnIfMalformedConditionNode),
	// but need to actually be called from all three places a composite
	// condition tree can be entered -- confirms the wiring, not just the
	// helper's own existence.
	js := generatedJS(t)
	mustContain(t, js, "warnIfMalformedConditionNode_test(condition, 'StatesManager.evaluateCondition')")
	mustContain(t, js, "warnIfMalformedConditionNode_test(trigger, 'RulesManager.evaluateTriggerCondition')")
	mustContain(t, js, "warnIfMalformedConditionNode_test(node, 'RulesManager.evaluateTriggerConditionNode')")
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected generated JS to contain %q, not found", needle)
	}
}

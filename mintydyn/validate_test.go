package mintydyn

import (
	"strings"
	"testing"
)

func TestValidate_CleanConfigProducesNoErrors(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{ComponentID: "x", Condition: "equals", Value: "y"}},
			{ID: "r2", Trigger: TriggerCondition{AllOf: []TriggerCondition{
				{ComponentID: "a", Condition: "equals", Value: 1},
				{ComponentID: "b", Condition: "greaterThan", Value: 2},
			}}},
		}).
		WithStates([]ComponentState{
			{ID: "s1", Condition: &StateCondition{Field: "f", Operator: "equals", Value: "v"}},
		})
	if err := db.Validate(); err != nil {
		t.Errorf("expected no errors for a clean config, got: %v", err)
	}
}

func TestValidate_DetectsUnrecognizedOperator(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{ComponentID: "x", Condition: "eqauls", Value: "y"}},
		})
	err := db.Validate()
	if err == nil || !strings.Contains(err.Error(), `unrecognized operator "eqauls"`) {
		t.Errorf("expected an unrecognized-operator error, got: %v", err)
	}
}

func TestValidate_DetectsAllOfAndAnyOfBothSet(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{
				AllOf: []TriggerCondition{{ComponentID: "a", Condition: "equals", Value: 1}},
				AnyOf: []TriggerCondition{{ComponentID: "b", Condition: "equals", Value: 2}},
			}},
		})
	err := db.Validate()
	if err == nil || !strings.Contains(err.Error(), "both allOf and anyOf set") {
		t.Errorf("expected an allOf+anyOf-both-set error, got: %v", err)
	}
}

func TestValidate_DetectsCompositeAndLeafBothSet(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{
				ComponentID: "x", Condition: "equals", Value: "y",
				AllOf: []TriggerCondition{{ComponentID: "a", Condition: "equals", Value: 1}},
			}},
		})
	err := db.Validate()
	if err == nil || !strings.Contains(err.Error(), "both a composite (allOf/anyOf) and leaf fields") {
		t.Errorf("expected a composite+leaf-both-set error, got: %v", err)
	}
}

func TestValidate_DetectsEmptyAllOf(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{AllOf: []TriggerCondition{}}},
		})
	err := db.Validate()
	if err == nil || !strings.Contains(err.Error(), "allOf is an empty") {
		t.Errorf("expected an empty-allOf error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "vacuous AND") {
		t.Errorf("expected the vacuous-AND explanation, got: %v", err)
	}
}

func TestValidate_DetectsEmptyAnyOf(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{AnyOf: []TriggerCondition{}}},
		})
	err := db.Validate()
	if err == nil || !strings.Contains(err.Error(), "anyOf is an empty") {
		t.Errorf("expected an empty-anyOf error, got: %v", err)
	}
}

func TestValidate_DetectsDeadRuleWithNoComponentIdAnywhere(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "dead-rule", Trigger: TriggerCondition{Condition: "equals", Value: 1}}, // no ComponentID
		})
	err := db.Validate()
	if err == nil || !strings.Contains(err.Error(), `rule "dead-rule"`) || !strings.Contains(err.Error(), "will never fire") {
		t.Errorf("expected a dead-rule error naming the rule, got: %v", err)
	}
}

func TestValidate_DetectsIssuesAtCorrectNestingDepth(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{AllOf: []TriggerCondition{
				{ComponentID: "a", Condition: "equals", Value: 1},
				{ComponentID: "b", Condition: "typo-operator", Value: 2}, // nested, index 1
			}}},
		})
	err := db.Validate()
	if err == nil || !strings.Contains(err.Error(), "allOf[1]") {
		t.Errorf("expected the error to identify the specific nested index (allOf[1]), got: %v", err)
	}
}

func TestValidate_StateConditionDetectsIssuesToo(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithStates([]ComponentState{
			{ID: "s1", Condition: &StateCondition{Operator: "bad-op", Field: "f", Value: "v"}},
		})
	err := db.Validate()
	if err == nil || !strings.Contains(err.Error(), `state "s1"`) || !strings.Contains(err.Error(), `unrecognized operator "bad-op"`) {
		t.Errorf("expected a StateCondition operator error naming the state, got: %v", err)
	}
}

func TestValidate_MultipleIssuesAllReported(t *testing.T) {
	// errors.Join should aggregate every issue found, not just the first.
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{Condition: "typo1", ComponentID: "x"}},
			{ID: "r2", Trigger: TriggerCondition{AllOf: []TriggerCondition{}}},
		})
	err := db.Validate()
	if err == nil {
		t.Fatal("expected errors")
	}
	msg := err.Error()
	if !strings.Contains(msg, "typo1") || !strings.Contains(msg, "allOf is an empty") {
		t.Errorf("expected both issues to be reported together, got: %v", msg)
	}
}

// TestIsKnownOperator_MatchesGeneratedJS is the cross-check promised in
// validate.go's own comment: isKnownOperator is a hand-maintained,
// separate list from evaluateConditionOperator_%s's own switch in the
// generated JavaScript. If they ever drift -- someone adds an operator to
// one and forgets the other -- this test catches it, rather than letting
// Validate() silently accept or reject operators the actual runtime
// evaluator disagrees with.
func TestIsKnownOperator_MatchesGeneratedJS(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test")
	js := db.generateBaseClass()

	knownOperators := []string{
		"equals", "notEquals", "contains", "greaterThan", "lessThan",
		"checked", "unchecked", "empty", "notEmpty",
	}
	for _, op := range knownOperators {
		if !isKnownOperator(op) {
			t.Errorf("isKnownOperator rejects %q, but it should be known", op)
		}
		caseLine := "case '" + op + "':"
		if !strings.Contains(js, caseLine) {
			t.Errorf("generated JS's operator switch is missing %q (isKnownOperator claims to know it, but the generated evaluator doesn't have a case for it)", caseLine)
		}
	}
	if isKnownOperator("definitelyNotAnOperator") {
		t.Error("isKnownOperator wrongly accepts a made-up operator name")
	}
}

func TestValidateWithHandler_CalledForEveryIssue(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{ComponentID: "x", Condition: "eqauls", Value: 1}},
			{ID: "r2", Trigger: TriggerCondition{AllOf: []TriggerCondition{}}},
		})

	var received []Issue
	err := db.ValidateWithHandler(func(issue Issue) {
		received = append(received, issue)
	})

	if err == nil {
		t.Fatal("expected an aggregated error still returned")
	}
	// r1 has one issue (unrecognized operator). r2's trigger is an empty
	// AllOf slice, which is itself flagged (IssueEmptyAllOf) *and* means
	// collectTriggerComponentIDs finds no componentId anywhere, so r2 is
	// also flagged as a dead rule (IssueDeadRule) -- three issues total,
	// not two. Confirmed by inspecting the actual output before fixing
	// this expectation, not just relaxing the assertion.
	if len(received) != 3 {
		t.Fatalf("expected the handler called exactly 3 times, got %d calls: %+v", len(received), received)
	}
}

func TestValidateWithHandler_PassesValueAndValidList(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{ComponentID: "x", Condition: "eqauls", Value: 1}},
		})

	var got Issue
	calls := 0
	db.ValidateWithHandler(func(issue Issue) {
		calls++
		got = issue
	})

	if calls != 1 {
		t.Fatalf("expected exactly 1 call, got %d", calls)
	}
	if got.Kind != IssueUnrecognizedOperator {
		t.Errorf("expected IssueUnrecognizedOperator, got %v", got.Kind)
	}
	if got.Value != "eqauls" {
		t.Errorf("expected Value %q, got %q", "eqauls", got.Value)
	}
	if len(got.Valid) != 9 {
		t.Fatalf("expected 9 valid operators, got %d: %v", len(got.Valid), got.Valid)
	}
	found := false
	for _, v := range got.Valid {
		if v == "equals" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Valid to contain the likely intended operator %q, got %v", "equals", got.Valid)
	}
}

func TestValidateWithHandler_StructuralIssuesHaveNoValueOrValid(t *testing.T) {
	// Confirms Value/Valid are deliberately left empty for non-enum-style
	// issues, rather than inventing something -- structural issues
	// (both allOf/anyOf set, empty allOf, etc.) have no natural "valid
	// alternatives" list.
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{AllOf: []TriggerCondition{}}},
		})

	var got Issue
	db.ValidateWithHandler(func(issue Issue) {
		got = issue
	})

	if got.Kind != IssueEmptyAllOf {
		t.Fatalf("expected IssueEmptyAllOf, got %v", got.Kind)
	}
	if got.Value != "" {
		t.Errorf("expected empty Value for a structural issue, got %q", got.Value)
	}
	if got.Valid != nil {
		t.Errorf("expected nil Valid for a structural issue, got %v", got.Valid)
	}
}

func TestValidateWithHandler_NilHandlerBehavesLikeValidate(t *testing.T) {
	db := New[[]ComponentState, []map[string]interface{}, []DependencyRule]("test").
		WithRules([]DependencyRule{
			{ID: "r1", Trigger: TriggerCondition{ComponentID: "x", Condition: "eqauls", Value: 1}},
		})

	errViaValidate := db.Validate()
	errViaHandler := db.ValidateWithHandler(nil)

	if errViaValidate.Error() != errViaHandler.Error() {
		t.Errorf("expected identical error output with a nil handler, got:\nValidate:            %v\nValidateWithHandler: %v", errViaValidate, errViaHandler)
	}
}

func TestIssueKind_StringNames(t *testing.T) {
	cases := map[IssueKind]string{
		IssueUnrecognizedOperator:    "unrecognized_operator",
		IssueAllOfAndAnyOfBothSet:    "allof_and_anyof_both_set",
		IssueCompositeAndLeafBothSet: "composite_and_leaf_both_set",
		IssueEmptyAllOf:              "empty_allof",
		IssueEmptyAnyOf:              "empty_anyof",
		IssueDeadRule:                "dead_rule",
	}
	for kind, want := range cases {
		if got := kind.String(); got != want {
			t.Errorf("IssueKind(%d).String() = %q, want %q", kind, got, want)
		}
	}
}

package minty

import (
	"strings"
	"testing"
)

// Directly reproduces mogsie's issue #14 scenario: Input previously
// couldn't accept conditional attributes since If/IfElse return Evaluation
// (typed any), which doesn't satisfy Attribute. IfT/IfElseT fix this.
func TestIfT_InputAcceptsConditionalAttribute(t *testing.T) {
	b := &Builder{}
	maybe := true
	node := b.Input(IfT(maybe, Class("this")))

	var sb strings.Builder
	if err := node.Render(&sb); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	got := sb.String()
	if !strings.Contains(got, `class="this"`) {
		t.Errorf("expected class=\"this\" in rendered output, got: %s", got)
	}
}

func TestIfElseT_InputAcceptsConditionalAttribute(t *testing.T) {
	b := &Builder{}
	node := b.Input(IfElseT(false, Class("this"), Class("that")))

	var sb strings.Builder
	if err := node.Render(&sb); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	got := sb.String()
	if !strings.Contains(got, `class="that"`) {
		t.Errorf("expected class=\"that\" (condition false), got: %s", got)
	}
	if strings.Contains(got, `class="this"`) {
		t.Errorf("did not expect class=\"this\" when condition is false, got: %s", got)
	}
}

// The exact nested case mogsie called "an abomination" -- IfElseT of an
// IfElseT. Confirms no special-casing is needed anywhere for this.
func TestIfElseT_NestedConditionalAttribute(t *testing.T) {
	b := &Builder{}
	inner := IfElseT(true, Class("inner-true"), Class("inner-false"))
	// Explicit type parameter needed here: Go can't infer a common T across
	// a concrete Attribute value and a TypedEvaluation[Attribute] value in
	// the same call -- confirmed directly, this failed to compile without
	// it. Worth documenting plainly since a real user hits this exact error.
	node := b.Input(IfElseT[Attribute](false, Class("outer-true"), inner))

	var sb strings.Builder
	if err := node.Render(&sb); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	got := sb.String()
	if !strings.Contains(got, `class="inner-true"`) {
		t.Errorf("expected class=\"inner-true\" (outer false -> inner, inner true), got: %s", got)
	}
}

// Confirms the existing, non-generic Evaluation/If/IfElse are completely
// untouched -- this is genuinely additive, not a replacement.
func TestExistingEvaluationAPIStillWorks(t *testing.T) {
	b := &Builder{}
	node := b.Div(b.IfElse(true, "still works", "nope"))

	var sb strings.Builder
	if err := node.Render(&sb); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	got := sb.String()
	if !strings.Contains(got, "still works") {
		t.Errorf("expected existing b.IfElse to still work, got: %s", got)
	}
}

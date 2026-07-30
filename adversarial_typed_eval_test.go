package minty_test

import (
	"strings"
	"testing"

	mi "github.com/ha1tch/minty"
)

// =============================================================================
// Adversarial regression tests for TypedEvaluation (IfT/IfElseT)
// =============================================================================
//
// These document two real, confirmed edge cases found while investigating
// TypedEvaluation's own type-safety guarantees after mogsie's original
// issue (#14) prompted a closer look than the initial implementation and
// test suite had covered. One was a genuine, silent bug (now fixed); the
// other is expected, unavoidable Go behavior worth confirming explicitly
// so it stays that way, not something to "fix" further.

// customClassAttr is a user's own custom Attribute implementation -- a
// realistic scenario, not a contrived one, since Attribute is a public
// interface users are meant to implement themselves.
type customClassAttr struct{ value string }

func (c customClassAttr) Apply(el *mi.Element) {
	// Deliberately applies even for a zero-value receiver (value == "").
	// This is the point of the test: Apply here never panics on a
	// zero value, so if TypedEvaluation ever mishandled this case again,
	// the failure would be silent, wrong output -- not a crash that's
	// easy to notice.
	if c.value != "" {
		mi.Class(c.value).Apply(el)
	} else {
		mi.Class("EMPTY-VALUE-WAS-APPLIED").Apply(el)
	}
}

// TestIfT_ConcreteTypeDoesNotApplyZeroValueOnFalseBranch is a regression
// test for a real, confirmed bug in an earlier version of this code, not
// a hypothetical.
//
// Root cause: TypedEvaluation originally decided whether to apply
// falseValue by comparing it against T's zero value
// (Attribute(e.falseValue) != nil). That's correct when T is inferred as
// the Attribute interface itself, whose zero value is genuinely nil. It
// silently breaks when T is inferred as a concrete struct type instead --
// which happens whenever a value is passed to IfT without an explicit
// Attribute(...) conversion, exactly as a user implementing their own
// Attribute type would naturally write it. A concrete struct's zero value
// is a real, non-nil value of that struct, not nil, so the old check
// wrongly treated "falseValue was never set" as "falseValue is the
// zero-value struct" and applied it -- with no error or panic to reveal
// the mistake, just wrong rendered output.
//
// Fixed by tracking whether falseValue was ever set with an explicit
// hasFalse bool, set by IfT (false) vs IfElseT (true), which never
// depends on what T happens to be. This test exists so that fix can never
// silently regress.
func TestIfT_ConcreteTypeDoesNotApplyZeroValueOnFalseBranch(t *testing.T) {
	b := &mi.Builder{}
	node := b.Input(mi.IfT(false, customClassAttr{value: "should-not-appear"}))
	var sb strings.Builder
	if err := node.Render(&sb); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	got := sb.String()
	if strings.Contains(got, "EMPTY-VALUE-WAS-APPLIED") {
		t.Fatalf("regression: zero-value concrete Attribute was silently applied on IfT's untaken branch, got: %s", got)
	}
	if strings.Contains(got, "should-not-appear") {
		t.Fatalf("wrong branch applied entirely, got: %s", got)
	}
	if got != `<input />` {
		t.Errorf("expected a bare <input />, got: %s", got)
	}
}

// nilAttr's Apply panics on a nil receiver, specifically so this test can
// observe whether TypedEvaluation ever calls Apply on a nil *nilAttr.
type nilAttr struct{ tag string }

func (n *nilAttr) Apply(el *mi.Element) {
	if n == nil {
		panic("Apply called on a nil *nilAttr receiver")
	}
}

// TestIfElseT_ExplicitTypedNilPanics documents expected, unavoidable Go
// behavior -- not a bug, and deliberately not "fixed" further. When a
// user explicitly passes a nil pointer as IfElseT's falseValue (as
// opposed to IfT's implicit, never-set case above), hasFalse is
// correctly true, so Apply is correctly called on it. If that nil
// pointer's own Apply method panics on a nil receiver, that is the
// user's own Attribute implementation's responsibility, matching
// ordinary Go interface-method-call semantics on any nil concrete
// receiver -- minty deliberately does not add nil-guarding magic on top
// of a value the caller explicitly, deliberately provided, since silently
// swallowing an explicitly-passed nil would be its own kind of surprise.
// This test exists so a future change doesn't accidentally paper over
// this and call it a fix.
func TestIfElseT_ExplicitTypedNilPanics(t *testing.T) {
	var p *nilAttr // an explicit, deliberate nil pointer
	b := &mi.Builder{}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected a panic (ordinary Go nil-receiver semantics for an explicitly-passed nil), got none")
		}
	}()

	node := b.Input(mi.IfElseT[mi.Attribute](false, mi.Class("x"), p))
	var sb strings.Builder
	node.Render(&sb) // panics inside p.Apply, not here
}

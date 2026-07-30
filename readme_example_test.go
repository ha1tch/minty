package minty_test

// Verifies the exact code examples shown in the README and in
// docs/minty-12-whats-new-0.2.0.md actually compile and behave as
// documented -- the simple IfT/IfElseT examples are in both; the nested
// example moved to Part 12 specifically (to avoid duplicating the same
// explanation in both places), but is still verified here.

import (
	"strings"
	"testing"

	mi "github.com/ha1tch/minty"
)

func TestReadmeExample_TypeSafeConditionalAttributes(t *testing.T) {
	b := &mi.Builder{}
	maybe := true
	node1 := b.Input(mi.IfT(maybe, mi.Class("highlighted")))
	var sb1 strings.Builder
	node1.Render(&sb1)
	if !strings.Contains(sb1.String(), `class="highlighted"`) {
		t.Errorf("IfT example failed: %s", sb1.String())
	}

	isValid := false
	node2 := b.Input(mi.IfElseT(isValid, mi.Class("valid"), mi.Class("invalid")))
	var sb2 strings.Builder
	node2.Render(&sb2)
	if !strings.Contains(sb2.String(), `class="invalid"`) {
		t.Errorf("IfElseT example failed: %s", sb2.String())
	}

	outer, inner := false, true
	node3 := b.Input(mi.IfElseT[mi.Attribute](outer, mi.Class("a"), mi.IfElseT(inner, mi.Class("b"), mi.Class("c"))))
	var sb3 strings.Builder
	node3.Render(&sb3)
	if !strings.Contains(sb3.String(), `class="b"`) {
		t.Errorf("nested IfElseT example failed: %s", sb3.String())
	}
}

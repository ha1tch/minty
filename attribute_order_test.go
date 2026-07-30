package minty

import (
	"strings"
	"testing"
)

// Proves the fix, not just that it looks right: without sorted keys, Go's
// randomized map iteration means the same element can render with
// different attribute orders across separate calls. Renders an element
// with several attributes many times and confirms the output string is
// byte-identical every time (github.com/mogsie, consistent-attribute-order).
func TestElementRender_DeterministicAttributeOrder(t *testing.T) {
	render := func() string {
		b := &Builder{}
		node := b.Input(Type("text"), Class("foo"), Name("bar"), ID("baz"), Placeholder("qux"))
		var sb strings.Builder
		if err := node.Render(&sb); err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		return sb.String()
	}

	first := render()
	for i := 0; i < 50; i++ {
		got := render()
		if got != first {
			t.Fatalf("attribute order not deterministic on iteration %d:\nfirst: %s\ngot:   %s", i, first, got)
		}
	}

	// Also confirm the actual order is alphabetical, not just stable.
	if !strings.Contains(first, `class="foo" id="baz" name="bar" placeholder="qux" type="text"`) {
		t.Errorf("expected alphabetically-sorted attributes, got: %s", first)
	}
}

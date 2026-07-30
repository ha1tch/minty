package minty

import (
	"strings"
	"testing"
)

// Proves the panic guard fires for the exact, realistic mistake it's meant
// to catch: passing If/IfElse's result (explicitly typed H) directly as a
// child argument, instead of calling it with a builder first. Without this
// guard, H silently falls through to the default case and renders as
// meaningless text (Go's %v formatting of a function value), rather than
// failing loudly at the actual mistake (github.com/mogsie,
// handle-misconstructed-tree).
//
// Important, worth documenting plainly: a bare func literal of the same
// signature (func(*Builder) Node{...}, not explicitly typed H) does NOT
// trigger this guard -- confirmed directly, a first version of this test
// used exactly that and failed to panic. Go's type switch matches on the
// value's actual dynamic type, and an anonymous func literal's inferred
// type is the unnamed func(*Builder) Node, not the named type H, even
// though they're structurally identical. This guard specifically covers
// values that are genuinely typed H, which is what If/IfElse actually
// return -- the realistic case, not every conceivable one.
func TestCreateElement_PanicsOnDirectHValue(t *testing.T) {
	b := &Builder{}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic when passing If's H result directly, got none")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "cannot use H type directly") {
			t.Errorf("expected a clear panic message about H, got: %v", r)
		}
	}()

	// If returns H; the mistake is passing it directly instead of calling
	// it with a builder: If(...)(b).
	b.Div(If(true, func(inner *Builder) Node { return inner.P("hello") }))
}

// Confirms the correct usage (calling H with a builder) still works fine
// and does not panic.
func TestCreateElement_CorrectHUsageDoesNotPanic(t *testing.T) {
	b := &Builder{}
	template := func(inner *Builder) Node { return inner.P("hello") }

	node := b.Div(template(b))
	var sb strings.Builder
	if err := node.Render(&sb); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.Contains(sb.String(), "hello") {
		t.Errorf("expected 'hello' in output, got: %s", sb.String())
	}
}

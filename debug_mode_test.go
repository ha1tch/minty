package minty_test

import (
	"errors"
	"strings"
	"testing"

	mi "github.com/ha1tch/minty"
)

// panicAttr always panics on Apply, deliberately, to exercise Debug mode's
// panic-recovery-and-wrap behavior directly.
type panicAttr struct{}

func (panicAttr) Apply(el *mi.Element) {
	panic("deliberate test panic inside a custom Attribute")
}

func TestBuilder_Debug_False_PanicsPlain(t *testing.T) {
	b := &mi.Builder{} // Debug is false, the zero value
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic")
		}
		if _, isWrapped := r.(*mi.ElementConstructionError); isWrapped {
			t.Errorf("did not expect Debug-mode wrapping when Debug is false, got: %v", r)
		}
		msg, ok := r.(string)
		if !ok || msg != "deliberate test panic inside a custom Attribute" {
			t.Errorf("expected the plain, unwrapped panic message, got: %v (%T)", r, r)
		}
	}()
	b.Div(panicAttr{})
}

func TestBuilder_Debug_True_WrapsWithContext(t *testing.T) {
	b := &mi.Builder{Debug: true}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic")
		}
		ece, ok := r.(*mi.ElementConstructionError)
		if !ok {
			t.Fatalf("expected *mi.ElementConstructionError, got %T: %v", r, r)
		}
		if ece.Tag != "div" {
			t.Errorf("expected Tag %q, got %q", "div", ece.Tag)
		}
		if ece.ArgIndex != 1 {
			t.Errorf("expected ArgIndex 1 (the second argument), got %d", ece.ArgIndex)
		}
		if !strings.Contains(ece.ArgType, "panicAttr") {
			t.Errorf("expected ArgType to mention panicAttr, got %q", ece.ArgType)
		}
		if ece.Original != "deliberate test panic inside a custom Attribute" {
			t.Errorf("expected Original to preserve the exact original panic value, got %v", ece.Original)
		}
		if len(ece.Stack) == 0 {
			t.Error("expected a non-empty captured stack trace")
		}
		if !strings.Contains(string(ece.Stack), "panicAttr") {
			t.Errorf("expected the captured stack to mention panicAttr's own Apply method, got:\n%s", ece.Stack)
		}
		errMsg := ece.Error()
		if !strings.Contains(errMsg, "div") || !strings.Contains(errMsg, "1") || !strings.Contains(errMsg, "deliberate test panic") {
			t.Errorf("Error() string missing expected context, got:\n%s", errMsg)
		}
	}()
	// A real (non-panicking) attribute at index 0, the panicking one at
	// index 1, to confirm the index is genuinely tracked per-argument,
	// not just always reporting 0.
	b.Div(mi.Class("real-attr-first"), panicAttr{})
}

func TestBuilder_Debug_True_SameOutputWhenNoPanic(t *testing.T) {
	// Debug mode must not change behavior at all when nothing panics --
	// only when something does.
	b := &mi.Builder{Debug: true}
	node := b.Input(mi.Class("x"), mi.ID("y"))
	var sb strings.Builder
	node.Render(&sb)
	got := sb.String()
	if !strings.Contains(got, `class="x"`) || !strings.Contains(got, `id="y"`) {
		t.Errorf("Debug mode changed normal (non-panicking) output: %s", got)
	}
}

func TestElementConstructionError_UnwrapsOriginalErrorPanic(t *testing.T) {
	b := &mi.Builder{Debug: true}
	sentinel := errors.New("sentinel underlying error")
	defer func() {
		r := recover()
		ece, ok := r.(*mi.ElementConstructionError)
		if !ok {
			t.Fatalf("expected *mi.ElementConstructionError, got %T", r)
		}
		if !errors.Is(ece, sentinel) {
			t.Errorf("expected errors.Is to find the wrapped sentinel error via Unwrap")
		}
	}()
	b.Div(errorPanicAttr{err: sentinel})
}

type errorPanicAttr struct{ err error }

func (e errorPanicAttr) Apply(el *mi.Element) {
	panic(e.err)
}

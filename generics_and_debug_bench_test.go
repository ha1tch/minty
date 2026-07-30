package minty_test

import (
	"io"
	"testing"

	mi "github.com/ha1tch/minty"
)

// =============================================================================
// Generics pathway (TypedEvaluation / IfT / IfElseT) benchmarks
// =============================================================================
//
// Measures the actual cost of the generic wrapper against a direct
// Attribute call and against the existing, non-generic Evaluation/If -- the
// three ways to express a conditional attribute, side by side, so the
// overhead of choosing type safety (IfT/IfElseT) is a measured number, not
// an assumption.

func BenchmarkDirectAttribute(b *testing.B) {
	bd := &mi.Builder{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node := bd.Input(mi.Class("x"))
		_ = node.Render(io.Discard)
	}
}

func BenchmarkExistingEvaluation_If(b *testing.B) {
	bd := &mi.Builder{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Div, not Input: the existing, non-generic If/IfElse can't be used
		// with Input at all (the problem IfT/IfElseT exist to solve) --
		// Div is the closest fair comparison available for the untyped path.
		node := bd.Div(bd.If(true, mi.Class("x")))
		_ = node.Render(io.Discard)
	}
}

func BenchmarkTypedEvaluation_IfT(b *testing.B) {
	bd := &mi.Builder{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node := bd.Input(mi.IfT(true, mi.Class("x")))
		_ = node.Render(io.Discard)
	}
}

func BenchmarkTypedEvaluation_IfElseT(b *testing.B) {
	bd := &mi.Builder{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node := bd.Input(mi.IfElseT(true, mi.Class("x"), mi.Class("y")))
		_ = node.Render(io.Discard)
	}
}

func BenchmarkTypedEvaluation_NestedTwoLevels(b *testing.B) {
	bd := &mi.Builder{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		inner := mi.IfElseT(true, mi.Class("a"), mi.Class("b"))
		node := bd.Input(mi.IfElseT[mi.Attribute](false, mi.Class("c"), inner))
		_ = node.Render(io.Discard)
	}
}

func BenchmarkTypedEvaluation_NestedFourLevels(b *testing.B) {
	bd := &mi.Builder{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l1 := mi.IfElseT(true, mi.Class("a"), mi.Class("b"))
		l2 := mi.IfElseT[mi.Attribute](false, mi.Class("c"), l1)
		l3 := mi.IfElseT[mi.Attribute](true, l2, mi.Class("d"))
		node := bd.Input(mi.IfElseT[mi.Attribute](false, mi.Class("e"), l3))
		_ = node.Render(io.Discard)
	}
}

// A realistic multi-attribute element, mixing several IfT/IfElseT calls --
// closer to what a real form input looks like than a single-attribute
// microbenchmark.
func BenchmarkTypedEvaluation_RealisticMultiAttribute(b *testing.B) {
	bd := &mi.Builder{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node := bd.Input(
			mi.Type("text"),
			mi.Name("email"),
			mi.IfElseT(true, mi.Class("valid"), mi.Class("invalid")),
			mi.IfT(false, mi.Attribute(mi.Required())),
			mi.IfElseT(false, mi.Placeholder("required"), mi.Placeholder("optional")),
		)
		_ = node.Render(io.Discard)
	}
}

// =============================================================================
// Builder.Debug overhead benchmarks
// =============================================================================
//
// Measures the actual cost of Debug mode's per-argument panic recovery on
// the non-panicking path (the common case -- Debug mode's whole value
// proposition is for the rare panicking case, so what matters here is how
// much it costs the other 99.9% of the time it's left on).

func BenchmarkCreateElement_Fast(b *testing.B) {
	bd := &mi.Builder{} // Debug: false, the zero value
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node := bd.Div(mi.Class("a"), mi.ID("b"), "text content")
		_ = node.Render(io.Discard)
	}
}

func BenchmarkCreateElement_Debug(b *testing.B) {
	bd := &mi.Builder{Debug: true}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node := bd.Div(mi.Class("a"), mi.ID("b"), "text content")
		_ = node.Render(io.Discard)
	}
}

// Same comparison with more arguments, since Debug mode's cost is
// per-argument (one recovery wrapper call per arg), not per-element --
// the overhead should scale with argument count, worth confirming rather
// than assuming.
func BenchmarkCreateElement_Fast_ManyArgs(b *testing.B) {
	bd := &mi.Builder{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node := bd.Div(
			mi.Class("a"), mi.ID("b"), mi.Type("c"), mi.Name("d"), mi.Placeholder("e"),
			"text1", "text2", "text3",
		)
		_ = node.Render(io.Discard)
	}
}

func BenchmarkCreateElement_Debug_ManyArgs(b *testing.B) {
	bd := &mi.Builder{Debug: true}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		node := bd.Div(
			mi.Class("a"), mi.ID("b"), mi.Type("c"), mi.Name("d"), mi.Placeholder("e"),
			"text1", "text2", "text3",
		)
		_ = node.Render(io.Discard)
	}
}

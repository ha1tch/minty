package mintydyn

import "testing"

// Three implementations of the exact same operator-validity check,
// benchmarked head to head, so the nested-switch dispatch's actual
// benefit (if any) is a measured number, not an assumption:
//   - isKnownOperator: the real, current implementation (length-first,
//     then a short switch among same-length candidates).
//   - isKnownOperatorFlatSwitch: a single flat switch comparing against
//     all nine operators unconditionally (what isKnownOperator looked
//     like before this change).
//   - isKnownOperatorLinearScan: a naive []string scan, included as a
//     deliberately worse baseline to calibrate the other two against.

func isKnownOperatorFlatSwitch(op string) bool {
	switch op {
	case "equals", "notEquals", "contains", "greaterThan", "lessThan",
		"checked", "unchecked", "empty", "notEmpty":
		return true
	default:
		return false
	}
}

func isKnownOperatorLinearScan(op string) bool {
	for _, k := range knownOperators {
		if op == k {
			return true
		}
	}
	return false
}

// benchInputs mixes hits and misses, and -- among the hits -- operators
// from every length bucket (including the two buckets, 8 and 9, that
// still need a within-bucket comparison), so the benchmark reflects a
// realistic distribution rather than only the cheapest or only the most
// expensive case.
var benchInputs = []string{
	"equals", "notEquals", "contains", "greaterThan", "lessThan",
	"checked", "unchecked", "empty", "notEmpty",
	"eqauls", "definitelyNotAnOperator", "", "x",
}

func BenchmarkIsKnownOperator_NestedSwitch(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, op := range benchInputs {
			_ = isKnownOperator(op)
		}
	}
}

func BenchmarkIsKnownOperator_FlatSwitch(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, op := range benchInputs {
			_ = isKnownOperatorFlatSwitch(op)
		}
	}
}

func BenchmarkIsKnownOperator_LinearScan(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, op := range benchInputs {
			_ = isKnownOperatorLinearScan(op)
		}
	}
}

// Separately: the worst case for each approach in isolation --
// non-matching input forces every implementation to do its maximum
// amount of work (linear scan checks all 9; flat switch's compiler-
// generated dispatch still has to rule everything out; nested switch
// still has to check the length bucket and, for buckets with multiple
// candidates, compare against each).
func BenchmarkIsKnownOperator_NestedSwitch_Miss(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = isKnownOperator("definitelyNotAnOperator")
	}
}

func BenchmarkIsKnownOperator_FlatSwitch_Miss(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = isKnownOperatorFlatSwitch("definitelyNotAnOperator")
	}
}

func BenchmarkIsKnownOperator_LinearScan_Miss(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = isKnownOperatorLinearScan("definitelyNotAnOperator")
	}
}

// TestThreeImplementationsAgree is a correctness guard for the benchmark
// comparison above: all three must classify every operator identically,
// or the benchmark would be comparing implementations that don't even
// do the same job.
func TestThreeImplementationsAgree(t *testing.T) {
	for _, op := range benchInputs {
		a := isKnownOperator(op)
		bResult := isKnownOperatorFlatSwitch(op)
		c := isKnownOperatorLinearScan(op)
		if a != bResult || bResult != c {
			t.Errorf("implementations disagree for %q: nested=%v flat=%v linear=%v", op, a, bResult, c)
		}
	}
}

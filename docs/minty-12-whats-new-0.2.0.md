# Minty Documentation - Part 12
## What's New in 0.2.0

> **Part of the Minty System**: This document covers everything new in
> the 0.2.0 release -- new APIs, fixed bugs (including one that shipped
> unnoticed since the original conditional-attributes work), measured
> performance data, and credit for the community contributions this
> release incorporates. All numbers in this document were measured
> directly on this codebase, not estimated -- see each section for how
> to reproduce them yourself.

---

### Table of Contents
1. [Summary](#summary)
2. [Type-safe conditional attributes: IfT and IfElseT](#type-safe-conditional-attributes-ift-and-ifelset)
3. [Performance](#performance)
4. [Composite conditions in mintydyn](#composite-conditions-in-mintydyn)
5. [Validate and ValidateWithHandler](#validate-and-validatewithhandler)
6. [Builder.Debug and ElementConstructionError](#builderdebug-and-elementconstructionerror)
7. [Deterministic attribute rendering order](#deterministic-attribute-rendering-order)
8. [Clearer failure on a common mistake](#clearer-failure-on-a-common-mistake)
9. [Credits](#credits)

---

## Summary

| Change | Kind | Breaking? |
|---|---|---|
| `IfT` / `IfElseT` / `TypedEvaluation[T Attribute]` | New API | No -- additive, `Evaluation`/`If`/`IfElse` unchanged |
| `mintydyn`: `AllOf` / `AnyOf` composite conditions | New API | No -- additive, existing configs unaffected |
| `mintydyn`: `Validate()` / `ValidateWithHandler()` / `Issue` | New API | No |
| `Builder.Debug` / `ElementConstructionError` | New API | No -- `Builder{}` (zero value) behaves exactly as before |
| Deterministic attribute rendering order | Behavior fix | No -- output was previously unordered; now it's sorted |
| Panic on passing `H` directly instead of calling it | Behavior fix | Only if code was relying on the previous silent-garbage-text behavior, which nothing sensibly could |

Nothing existing was renamed, removed, or had its signature changed. Every
addition in this release can be adopted incrementally, field by field,
attribute by attribute.

---

## Type-safe conditional attributes: IfT and IfElseT

**The problem** (reported by [@mogsie](https://github.com/mogsie),
[issue #14](https://github.com/ha1tch/minty/issues/14)): `Div` and other
heterogeneous elements accept `...interface{}`, so `b.Div(b.IfElse(cond,
mi.Class("a"), mi.Class("b")))` works fine. Self-closing elements like
`Input` are deliberately homogeneous (`...Attribute` only), and the
existing `Evaluation` returned by `If`/`IfElse` is typed as `any`, so it
doesn't satisfy `Attribute` -- `b.Input(b.If(cond, mi.Class("x")))`
simply doesn't compile.

**The fix**: `IfT`/`IfElseT`, new package-level generic functions
returning `TypedEvaluation[T Attribute]`, which implements `Apply`
directly:

```go
b.Input(mi.IfT(maybe, mi.Class("highlighted")))
b.Input(mi.IfElseT(isValid, mi.Class("valid"), mi.Class("invalid")))
```

Nesting works, but needs an explicit type parameter once you nest,
since Go can't infer a common `T` across a concrete `Attribute` value and
a `TypedEvaluation[Attribute]` value in the same call:

```go
mi.IfElseT[mi.Attribute](outer, mi.Class("a"), mi.IfElseT(inner, mi.Class("b"), mi.Class("c")))
```

This is purely additive -- `Evaluation`, `(*Builder).If`, and
`(*Builder).IfElse` are completely unchanged and still work exactly as
before for `Div`/`P`/other heterogeneous elements.

**Why `TypedEvaluation` is `Attribute`-only, not also generic over
`Node`**: an earlier version of this design supported both, giving every
`TypedEvaluation[T]` unconditional `Render` and `Apply` methods
regardless of `T`. That meant every instance satisfied both `Node` and
`Attribute` simultaneously, and `createElement`'s type switch always
picked whichever case it checked first -- silently dropping the other.
Confirmed directly with a failing test: a `TypedEvaluation[Node]` passed
to `Div` rendered as an empty element, its child completely missing, no
error anywhere. Constraining `T` to `Attribute` rules this out at compile
time. There's no real gap left to fill for conditional `Node`s in
heterogeneous elements -- the existing, untouched `Evaluation`/`If`/
`IfElse` already handles that case correctly.

**A bug found and fixed along the way, not shipped as a known
limitation**: `TypedEvaluation` originally decided whether to apply
`falseValue` by comparing it against `T`'s zero value. That's correct
when `T` infers as the `Attribute` interface itself (zero value `nil`),
and silently wrong when `T` infers as a concrete struct type instead --
which happens naturally whenever a value is passed to `IfT` without an
explicit `Attribute(...)` conversion, exactly how a user implementing
their own `Attribute` type would write it. A concrete struct's zero value
is a real, non-nil value of that struct, not `nil`, so the original code
was silently applying it on `IfT`'s untaken branch. Fixed with an
explicit `hasFalse` field, set by `IfT` (`false`) vs `IfElseT` (`true`),
which never depends on what `T` happens to be. See
[Part 13, the type-safety guide](minty-13-type-safety.md), for the full
account of this and related findings.

---

## Performance

All numbers below are from `go test -bench=. -benchmem`, `cpu: Intel(R)
Xeon(R) Processor @ 2.80GHz`. Run them yourself:
`go test -bench=. -benchmem -run=^$ .` (core package) and
`go test -bench=. -benchmem -run=^$ ./mintydyn/...` (mintydyn).

### IfT/IfElseT vs. the old, untyped path

The new, type-safe path is not just safer -- it's measurably faster than
the old, untyped `Evaluation`/`If` path it sits alongside, because it
skips `createElement`'s `Evaluation`-unwrapping loop entirely (it
satisfies `Attribute` directly, so the existing `switch arg.(type) {
case Attribute: ... }` handles it in one step):

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| `BenchmarkDirectAttribute` (baseline, no conditional) | ~644 | 480 | 8 |
| `BenchmarkExistingEvaluation_If` (old, untyped path) | ~819 | 544 | 10 |
| `BenchmarkTypedEvaluation_IfT` | ~698 | 528 | 9 |
| `BenchmarkTypedEvaluation_IfElseT` | ~757 | 560 | 10 |
| `BenchmarkTypedEvaluation_NestedTwoLevels` | ~883 | 640 | 12 |
| `BenchmarkTypedEvaluation_NestedFourLevels` | ~1154 | 800 | 16 |

`IfT`/`IfElseT` run roughly 10-15% faster than the old `Evaluation`/`If`
path for the same single-conditional case, observed consistently across
repeated runs, with cost scaling close to linearly as nesting depth
increases.

### Against mogsie's own original `generic-if` exploration

Built mogsie's exact `generic-if` branch code standalone (unmodified) and
benchmarked the identical workload against `IfT`/`IfElseT`:

| Shape | mogsie's `generic-if` | This release | Difference |
|---|---:|---:|---:|
| Single conditional (`IfT`) | 3776 ns/op | 698 ns/op | ~5.4x faster |
| Single conditional (`IfElseT`) | 4159 ns/op | 757 ns/op | ~5.5x faster |
| Nested, 2 levels | 4576 ns/op | 883 ns/op | ~5.2x faster |
| Nested, 4 levels | 4680 ns/op | 1154 ns/op | ~4.1x faster |

Allocation counts were identical between both implementations at every
depth -- the entire gap is CPU time, from `generic-if`'s own
unwrapping loop running up to four type assertions per argument
regardless of whether any of them were needed, a cost `IfT`/`IfElseT`
avoids entirely by satisfying `Attribute` directly.

### Switch-based dispatch vs. a linear scan

`mintydyn`'s `isKnownOperator` (used by `Validate`'s operator-validity
check) dispatches via a length-first nested switch. Measured against a
single flat switch and a naive linear `[]string` scan doing the exact
same job:

| Implementation | Mixed workload | Non-matching ("miss") case |
|---|---:|---:|
| Nested switch (length-first) | ~29 ns/op | ~0.33 ns/op |
| Flat switch | ~28 ns/op | ~0.32 ns/op |
| Linear scan | ~88 ns/op | ~4.5 ns/op |

The honest finding here: nested vs. flat made **no measurable
difference** at this scale (9 short strings) -- Go's compiler already
dispatches a flat string switch efficiently on its own. What *is* a real,
large, measured difference is switch (either form) vs. linear scan:
roughly 3x on a mixed workload, and roughly 14x specifically on a miss,
where a linear scan must exhaust the whole list before concluding "not
found," while either switch form resolves that in near-constant time.
The nested form is kept for its documentation value in showing the
technique, not because it measurably outperforms the simpler flat switch
here -- worth knowing if you're deciding whether to reach for this
pattern in your own code at a similar scale.

### Builder.Debug overhead

`Builder.Debug` wraps each argument's processing in element construction
with panic recovery -- see
[Builder.Debug and ElementConstructionError](#builderdebug-and-elementconstructionerror)
below. Measured cost on the (common) non-panicking path:

| Benchmark | ns/op |
|---|---:|
| `BenchmarkCreateElement_Fast` (3 args, `Debug: false`) | ~1249 |
| `BenchmarkCreateElement_Debug` (3 args, `Debug: true`) | ~1267 |
| `BenchmarkCreateElement_Fast_ManyArgs` (8 args, `Debug: false`) | ~2708 |
| `BenchmarkCreateElement_Debug_ManyArgs` (8 args, `Debug: true`) | ~2716 |

Overhead is small and roughly proportional to argument count -- typically
in the low single-digit percent range for a handful of arguments, though
this varies somewhat run to run; measure your own workload if it matters
for your use case rather than trusting a single number here. `Debug` is
off by default (the zero value), so nothing pays this cost unless
explicitly opted in.

---

## Composite conditions in mintydyn

`StateCondition` and `TriggerCondition` both gained `AllOf`/`AnyOf`
fields, expressing AND/OR logic across multiple fields, nesting
arbitrarily:

```go
mintydyn.TriggerCondition{
    AllOf: []mintydyn.TriggerCondition{
        {ComponentID: "country", Condition: "equals", Value: "US"},
        {ComponentID: "age", Condition: "greaterThan", Value: 18},
    },
}
```

Every existing `TriggerCondition{ComponentID: ...}` (or `StateCondition{
Field: ...}`) value is a valid leaf under the new shape -- purely
additive, no existing config's meaning changes. The two evaluator
functions that previously each had their own near-identical switch
statement (`evaluateCondition`'s five operators were a verbatim subset of
`evaluateTriggerCondition`'s nine) now share one implementation.

See [Part 13, the type-safety guide](minty-13-type-safety.md) for a full
account of six real footguns this shape makes possible (an empty `allOf`
evaluating as vacuously true, for example) and what mitigates them.

---

## Validate and ValidateWithHandler

```go
if err := builder.Validate(); err != nil {
    log.Fatal(err) // or t.Fatal in a test
}
```

A Go-callable pre-flight check for the composite-condition footguns
described in Part 13 -- catch them in a test or at startup, in Go,
rather than only as a browser `console.warn` that requires dev tools open
at the right moment to ever see.

`ValidateWithHandler` additionally calls a closure once per issue found,
receiving the offending value and (where applicable) the full set of
valid alternatives:

```go
builder.ValidateWithHandler(func(issue mintydyn.Issue) {
    if issue.Kind == mintydyn.IssueUnrecognizedOperator {
        log.Printf("%s: %q isn't valid -- did you mean one of %v?",
            issue.Context, issue.Value, issue.Valid)
    }
})
```

---

## Builder.Debug and ElementConstructionError

```go
b := &mi.Builder{Debug: true}
```

When `Debug` is true, a panic during element construction (e.g. inside a
custom `Attribute`'s own `Apply` method) is recovered, wrapped with
context -- which argument index, which concrete type, which element tag,
and the full stack trace captured at the original panic site -- and
re-panicked as an `*ElementConstructionError`, rather than surfacing as a
bare, contextless runtime panic. It still fails; it now explains itself
first, including exactly where to look. `Unwrap` is implemented, so
`errors.Is`/`errors.As` still find the original error if the panic value
was itself one.

Off by default -- see [Performance](#performance) above for the measured
cost when it's on.

---

## Deterministic attribute rendering order

Contributed by [@mogsie](https://github.com/mogsie). Rendered HTML
attributes are now sorted alphabetically, rather than following Go's
randomized map iteration order. Previously, the same element could render
with a different attribute order across separate runs -- harmless for a
browser, but a real problem for anything comparing rendered output
byte-for-byte (snapshot tests, diffs, caching by content hash).

## Clearer failure on a common mistake

Also contributed by [@mogsie](https://github.com/mogsie). Passing an `H`
(a `func(*Builder) Node` template, as returned by `If`/`IfElse`) directly
as a child argument, instead of calling it with a builder first, now
panics with a specific, actionable message instead of silently rendering
as meaningless text (Go's default `%v` formatting of a function value).

Note this only fires for values whose exact, named type is `H` -- an
anonymous `func(*Builder) Node` literal, not explicitly typed as `H`, has
a different dynamic type in Go's eyes and won't trigger this specific
check, even though it's structurally identical. In practice this covers
the realistic mistake, since `If`/`IfElse` genuinely return `H`.

---

## Credits

Thanks to [@mogsie](https://github.com/mogsie) for filing
[issue #14](https://github.com/ha1tch/minty/issues/14), for the
`consistent-attribute-order` and `handle-misconstructed-tree`
contributions taken directly in this release, and for the `generic-if`
exploration that -- even though its own approach wasn't taken as-is --
shaped the design of `IfT`/`IfElseT` and directly prompted the deeper
investigation that found and fixed a real bug (the concrete-type
zero-value issue described above and in Part 13) present in both that
exploration and this release's own first draft.

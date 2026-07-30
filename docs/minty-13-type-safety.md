# Minty Documentation - Part 13
## Type Safety: What It Covers, What It Can't, and What Minty Does About the Gap

> **Part of the Minty System**: This document is a candid accounting of
> where minty's type safety genuinely holds, where Go's type system
> fundamentally cannot help (regardless of how minty is designed), and
> what minty offers as a practical mitigation for the gaps that remain.
> Every example here is a real finding from this codebase's own 0.2.0
> development, not a hypothetical -- most were caught by deliberately
> adversarial testing, and several were genuine bugs, now fixed, that
> shipped unnoticed until that testing found them.

---

### Table of Contents
1. [What type safety actually covers here](#what-type-safety-actually-covers-here)
2. [What it cannot cover, and why](#what-it-cannot-cover-and-why)
3. [What minty offers to mitigate the gap](#what-minty-offers-to-mitigate-the-gap)
4. [A note on Go itself, not just minty](#a-note-on-go-itself-not-just-minty)
5. [Practical guidance if you're implementing your own Attribute or Node](#practical-guidance-if-youre-implementing-your-own-attribute-or-node)

---

## What type safety actually covers here

These are real, verified compile-time guarantees -- each one confirmed
by deliberately writing code that should fail to compile, and checking
that it does.

**`Input` (and every other self-closing element) genuinely cannot accept
a `Node`.** `Input(attrs ...Attribute)`'s signature makes this a compile
error, not a runtime check. This is the entire reason `IfT`/`IfElseT`
exist scoped to `Attribute` specifically -- the constraint they respect
is a real one, not an arbitrary restriction.

**`IfElseT`'s two branches must be the same concrete type, unless
explicitly widened to a common interface.** Verified directly:

```go
type typeA struct{}
func (typeA) Apply(el *mi.Element) {}
type typeB struct{}
func (typeB) Apply(el *mi.Element) {}

b.Input(mi.IfElseT(true, typeA{}, typeB{}))
// ./main.go:14:40: in call to mi.IfElseT, type typeB of typeB{}
// does not match inferred type typeA for T
```

Go's own generic type inference rejects this at compile time, with a
clear, specific error naming both mismatched types. Widening to
`IfElseT[mi.Attribute](...)` explicitly is required, and that's a
deliberate signal at the call site that you're choosing a wider,
less-specific type, not an accident.

**Nesting `TypedEvaluation` values works correctly to arbitrary depth,
via ordinary interface satisfaction, not special-cased unwrapping
logic.** Because `TypedEvaluation[Attribute]` itself satisfies
`Attribute`, an `IfElseT` nested inside another `IfElseT`'s branch
resolves correctly through Go's own method dispatch -- confirmed
directly with three levels of nesting, including a false-branch path
through two nested conditionals, with zero special-casing anywhere in
`createElement`. Compare this to the alternative approach explored in
[issue #14](https://github.com/ha1tch/minty/issues/14), which needed an
explicit, four-case type-switch specifically to handle nesting -- the
`Attribute`-only constraint here isn't just narrower, it's what makes the
nesting case *not need* special handling at all.

**A `TypedEvaluation` cannot ambiguously satisfy both `Node` and
`Attribute`.** An earlier version of this design did allow this (generic
over `any`, with both `Render` and `Apply` methods regardless of the
type parameter), and it was a real, confirmed bug -- see
[What it cannot cover](#what-it-cannot-cover-and-why) below for why that
one specifically couldn't be caught by the type system as originally
designed, and how constraining `T` to `Attribute` closed it.

---

## What it cannot cover, and why

Every one of these is a genuine, found instance of Go's type system
having no way to express what would be needed -- not a minty design
choice that could have been made more strictly, and not a matter of
trying harder.

### Value-level constraints on otherwise well-typed data

`mintydyn`'s `TriggerCondition`/`StateCondition` composite trees
(`AllOf`/`AnyOf`, `docs/minty-12-whats-new-0.2.0.md` describes the
feature) are ordinary, well-typed Go structs serialized to JSON for a
JavaScript evaluator to interpret at runtime. Six real footguns were
found by deliberately adversarial testing, and every one of them is a
constraint on a *value*, not a *type*:

- An `AllOf` slice with zero entries is a perfectly valid `[]TriggerCondition{}`
  -- but evaluates as vacuously true (`Array.prototype.every` on an empty
  array), the opposite of what an empty condition list probably means to
  whoever configured it.
- An `AnyOf` slice with zero entries is equally valid, and evaluates as
  vacuously false.
- An unrecognized operator string like `"eqauls"` (a typo for `"equals"`)
  is a perfectly valid `string` -- Go has no way to constrain a field to
  one of nine specific string values at the type level without a
  hand-rolled enum type, and even then, JSON unmarshaling into a plain
  `string` field accepts anything.
- `AllOf` and `AnyOf` both being set on the same node is two independently
  valid fields both having values -- Go has no "these fields are mutually
  exclusive" constraint.
- A composite node (`AllOf`/`AnyOf` set) that *also* has leaf fields set
  (`ComponentID`, `Condition`, etc.) is the same problem: nothing in the
  type system says these shouldn't coexist.
- A rule whose trigger condition references no `ComponentID` anywhere in
  its tree is syntactically complete and type-correct -- it just can
  never be registered against any trigger and can never fire.

There is no dependent-type or refinement-type mechanism in Go to express
"this slice must be non-empty," "this string must be one of these nine
values," or "these two fields are mutually exclusive" as a *type*
constraint. These are all constraints on the *value* a well-typed field
happens to hold, which Go's type system has no vocabulary for at all --
not "doesn't do it well," but doesn't have the concept.

### The concrete-type zero-value gotcha

This one is more subtle, and it's a real bug that shipped -- not a
hypothetical. `TypedEvaluation`'s first version decided whether to apply
`falseValue` (`IfT`'s untaken branch) by comparing it against `T`'s zero
value. That's correct exactly when `T` infers as the `Attribute`
interface itself, whose zero value is genuinely `nil` -- and silently
wrong when `T` infers as a concrete struct type instead, which happens
naturally whenever a value is passed to `IfT` without an explicit
`Attribute(...)` conversion (exactly how implementing your own custom
`Attribute` type and using it directly would naturally read). A concrete
struct's zero value is a real, non-nil value of that struct, not `nil`
-- so the original code was silently applying a zero-valued attribute
on the branch that was never supposed to render anything at all, with no
error or panic to reveal the mistake, just wrong output.

Confirmed present in
[mogsie's own original `generic-if` exploration](https://github.com/ha1tch/minty/issues/14)
too, by building that exact code and running the identical adversarial
test against it -- this wasn't caught in January by either the original
author or the initial review, only by this release's deliberately
adversarial testing pass.

Why this can't be fixed at the type level: Go generics have no way to
express "was a value of type `T` actually provided, as opposed to `T`'s
zero value" as a distinct *type*. Both cases produce a value of type `T`;
nothing in the type system distinguishes "the caller explicitly passed
the zero value" from "the caller passed nothing and this is what's left."
The fix (an explicit `hasFalse bool`, tracked independently of what `T`
happens to be) is a *runtime* bookkeeping mechanism, not a type-safety
one -- there wasn't a stricter type constraint available that would have
prevented this.

### Typed nil pointers wrapped in a non-nil interface

A different, related, but distinct case, and -- unlike the one above --
not a bug, deliberately not "fixed" further. Passing an explicit `nil`
pointer as `IfElseT`'s `falseValue` (not `IfT`'s implicit, never-set
case, which is the bug above) correctly calls `Apply` on it, and if that
pointer's own `Apply` method dereferences a nil receiver, it panics --
matching ordinary Go semantics for calling a method on any nil concrete
receiver wrapped in a non-nil interface value. This is the classic
"typed nil" gotcha, and it exists everywhere in Go, not specifically in
minty: an interface value is non-nil the moment it holds a concrete
type, even if that concrete value is itself a nil pointer.

Minty deliberately does not add nil-guarding logic here. Silently
swallowing a value the caller explicitly, deliberately provided would be
its own kind of surprise -- arguably a worse one, since it would hide a
real bug in the caller's own code (why would you deliberately pass a nil
pointer as a real value?) behind silently-missing output instead of a
panic pointing at the actual mistake.

### Runtime correctness of a user's own Attribute/Node implementation

The most general case, and true of any interface-based system in any
language with structural or nominal interface typing: `Attribute` and
`Node` are just method signatures. Go's type system verifies that a
method with the right signature *exists*; it has zero visibility into
what that method's body actually *does*. A custom `Attribute`
implementation that panics, that has a bug in its own logic, that
mutates state it shouldn't -- none of this is something any type system
checks, because "is this method's implementation correct" isn't a typing
question at all.

---

## What minty offers to mitigate the gap

None of the above can be fixed by better types. What minty does instead,
for each category:

**For the value-level, JSON-configured footguns** (`mintydyn`'s
composite conditions): `Validate()` and `ValidateWithHandler()` --
a Go-callable pre-flight check, walking the same condition trees the
generated JavaScript evaluates, catching all six footguns as a Go
`error` you can assert on in a test or check at startup, rather than
only ever surfacing as a browser `console.warn` that requires someone to
have dev tools open, at runtime, potentially in production, to ever see:

```go
if err := builder.Validate(); err != nil {
    t.Fatal(err) // catches it in CI, not in a user's browser
}
```

`ValidateWithHandler` goes further, handing you the offending value and
the complete set of valid alternatives for issues where that's
meaningful (currently: unrecognized operators), so custom tooling can
build its own reporting without re-deriving "what would have been valid
here" itself. See
[Part 12](minty-12-whats-new-0.2.0.md#validate-and-validatewithhandler)
for the full API.

This isn't prevention -- every one of these six footguns still resolves
to a defined, fail-safe result at runtime even if never validated
(nothing crashes). It's making the mistake visible early and precisely,
which is the practical ceiling for value-level constraints Go's type
system has no way to express.

**For runtime panics inside a user's own `Attribute`/`Node`
implementation**: `Builder.Debug`. It cannot prevent a panic -- nothing
can, from outside the panicking code -- but it recovers it, adds
context genuinely useful for finding the mistake (which argument index,
which concrete type, which element tag, and the full stack trace
captured at the original panic site, not just its final message), and
re-panics as an `*ElementConstructionError` rather than a bare, contextless
runtime panic:

```go
b := &mi.Builder{Debug: true}
// a panic inside any Attribute's Apply, or Node's Render, now explains
// itself: which argument, which tag, full stack -- before still failing
```

`errors.Is`/`errors.As` still work if the original panic value was
itself an error, via `Unwrap`. Off by default, since it does have a
measured (if small) cost -- see
[Part 12's performance section](minty-12-whats-new-0.2.0.md#performance).

**For the ambiguous dual-interface case**: this is the one gap that
*was* closeable at the type level, and it's worth naming as the
exception to everything else in this document. Constraining
`TypedEvaluation`'s type parameter to `T Attribute` (rather than `T any`)
rules the ambiguity out at compile time -- there's no runtime mitigation
needed at all, because the problematic case can no longer be constructed.
This is the actual ideal outcome whenever it's available: not a better
runtime check, but making the mistake impossible to write in the first
place. It just isn't available for the other cases in this document,
which is precisely why they needed a different kind of answer.

---

## A note on Go itself, not just minty

Every limitation in this document is a property of Go's type system in
general, not something specific to how minty is built. Any Go library
accepting free-form data (JSON, YAML, user-provided config) faces the
same value-vs-type gap; any Go library with an extensible interface
(anyone can implement `Attribute`) faces the same "type-checks but
correctness isn't verified" gap; any Go code using generics with an
interface-typed constraint faces the same concrete-type-zero-value
subtlety if it isn't careful. None of this is a minty shortcoming to
apologize for -- it's the actual, honest boundary of what a
statically-typed-but-not-dependently-typed language can guarantee, and
knowing exactly where that boundary sits is more useful than either
overclaiming type safety covers everything, or underclaiming it doesn't
matter.

---

## Practical guidance if you're implementing your own Attribute or Node

- Don't rely on your own type's zero value meaning anything in
  particular. If `IfT`/`IfElseT`'s "untaken branch" behavior matters to
  your type, remember minty already handles this correctly via
  `hasFalse` -- you don't need to defend against your own zero value
  being applied. But if you're writing other code that inspects a
  `TypedEvaluation`-like value's contents directly, don't assume "zero
  value" means "not provided."
- If your `Apply` or `Render` implementation can panic, consider whether
  callers using `Builder.Debug` will get useful information from that
  panic's message alone -- `ElementConstructionError` will add the
  argument index, type, and stack trace regardless, but a clear panic
  message from your own code is still the first line of defense.
- If you're building something that consumes `mintydyn` condition trees
  from configuration (JSON, a database, user input), call `Validate()`
  (or `ValidateWithHandler` if you want programmatic access to each
  issue) before trusting that configuration, the same way you'd validate
  any other structured input from outside your program's own control.

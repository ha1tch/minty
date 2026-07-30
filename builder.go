package minty

import (
	"fmt"
	"runtime/debug"
	"strconv"
)

// Builder provides methods for creating HTML elements with the Minty pattern.
//
// Debug, when true, wraps each argument's processing in element
// construction (createElement) with panic recovery that adds diagnostic
// context -- which argument (by index and concrete type), which element
// tag, and a full stack trace captured at the original panic site --
// before re-panicking with an *ElementConstructionError. It does not
// change what fails, only how much a failure explains about itself: a
// panic still stops execution (this is a programming error, not
// something to silently paper over), but with enough information to find
// the actual mistake instead of a bare "nil pointer dereference" with no
// indication of which attribute, on which element, caused it.
//
// Off by default (the zero value), since it adds a defer per element
// construction call -- see BenchmarkCreateElement_Debug vs
// BenchmarkCreateElement_Fast for the measured cost. Intended for
// development and test builds, not necessarily production, though the
// overhead may be acceptable there too depending on render volume --
// measure for your own workload rather than assuming either way.
type Builder struct {
	Debug bool
}

// ElementConstructionError wraps a panic recovered during element
// construction (Builder.Debug only) with the context needed to actually
// find the mistake: which argument, which element, and the original
// panic's own stack trace, not just its final message.
type ElementConstructionError struct {
	Tag      string
	ArgIndex int
	ArgType  string
	Original interface{}
	Stack    []byte
}

func (e *ElementConstructionError) Error() string {
	return fmt.Sprintf(
		"minty: panic while processing argument %d (%s) for <%s>: %v\n\nstack at panic:\n%s",
		e.ArgIndex, e.ArgType, e.Tag, e.Original, e.Stack,
	)
}

// Unwrap supports errors.Is/errors.As when the original panic value was
// itself an error (e.g. a panic(fmt.Errorf(...)), or many of the runtime
// package's own panic values, which implement error).
func (e *ElementConstructionError) Unwrap() error {
	if err, ok := e.Original.(error); ok {
		return err
	}
	return nil
}

// createElement creates an element with the given tag and processes mixed arguments.
func (b *Builder) createElement(tag string, selfClosing bool, args ...interface{}) Node {
	element := &Element{
		Tag:         tag,
		SelfClosing: selfClosing,
		Attributes:  make(map[string]string),
		Children:    []Node{},
	}

	if b.Debug {
		for i, arg := range args {
			processArgWithRecovery(element, tag, selfClosing, i, arg)
		}
		return element
	}

	for _, arg := range args {
		processArg(element, selfClosing, arg)
	}

	return element
}

// processArgWithRecovery wraps a single processArg call with panic
// recovery, adding context before re-panicking. Only called from the
// Debug-mode path in createElement -- the fast path calls processArg
// directly, with no defer at all.
func processArgWithRecovery(element *Element, tag string, selfClosing bool, index int, arg interface{}) {
	defer func() {
		if r := recover(); r != nil {
			panic(&ElementConstructionError{
				Tag:      tag,
				ArgIndex: index,
				ArgType:  fmt.Sprintf("%T", arg),
				Original: r,
				Stack:    debug.Stack(),
			})
		}
	}()
	processArg(element, selfClosing, arg)
}

// processArg handles a single createElement argument -- unwrapping
// Evaluation, then dispatching by type. Shared by both the fast path and
// the Debug-mode path so the actual dispatch logic exists exactly once.
func processArg(element *Element, selfClosing bool, arg interface{}) {
	for range 100 { // break out of infinite loops
		if eval, ok := arg.(Evaluation); ok {
			if eval.condition {
				arg = eval.trueValue
			} else {
				arg = eval.falseValue
			}
			continue
		}
		break
	}

	if arg == nil {
		return
	}

	switch v := arg.(type) {
	case H:
		panic("cannot use H type directly in element creation; call with (b *Builder)")
	case Attribute:
		v.Apply(element)
	case Node:
		if !selfClosing {
			element.Children = append(element.Children, v)
		}
	case string:
		if !selfClosing {
			element.Children = append(element.Children, &TextNode{Content: v})
		}
	case int:
		if !selfClosing {
			element.Children = append(element.Children, &TextNode{Content: strconv.Itoa(v)})
		}
	case fmt.Stringer:
		if !selfClosing {
			element.Children = append(element.Children, &TextNode{Content: v.String()})
		}
	default:
		if !selfClosing {
			element.Children = append(element.Children, &TextNode{Content: fmt.Sprintf("%v", v)})
		}
	}
}

// Logic for creating HTML elements
type Evaluation struct {
	condition  bool
	trueValue  any
	falseValue any
}

// If returns the Node if the condition is true, otherwise returns nil.
func (b *Builder) If(condition bool, item any) Evaluation {
	return Evaluation{
		condition: condition,
		trueValue: item,
	}
}

// IfElse returns the trueNode if condition is true, otherwise returns falseNode.
func (b *Builder) IfElse(condition bool, trueNode, falseNode any) Evaluation {
	return Evaluation{
		condition:  condition,
		trueValue:  trueNode,
		falseValue: falseNode,
	}
}

// TypedEvaluation is a type-safe, additive alternative to Evaluation for
// conditional Attributes (github.com/mogsie, issue #14). It exists alongside
// Evaluation/If/IfElse without changing either -- IfT/IfElseT are new
// functions, not replacements.
//
// The problem: createElement's heterogeneous element builders (Div, P, etc.)
// accept ...interface{}, so Evaluation (typed as `any`) flows through fine.
// Self-closing elements (Input, Img, etc.) are deliberately homogeneous
// (...Attribute only), and Evaluation doesn't satisfy Attribute, so
// b.Input(b.If(cond, mi.Class("x"))) doesn't compile.
//
// TypedEvaluation[T Attribute] fixes this by implementing Apply directly, so
// TypedEvaluation[Attribute] itself satisfies Attribute -- no wrapper
// unwrapping needed. Nesting (IfElseT of an IfElseT) works for free: T can
// itself be a TypedEvaluation[Attribute], and since that already satisfies
// Attribute, the outer TypedEvaluation's own Apply delegates correctly
// through arbitrary nesting depth via ordinary Go interface satisfaction --
// no per-depth type-switch cases needed anywhere. Verified directly (three
// levels of nesting, including a false-branch path through two nested
// conditionals) before relying on this claim.
//
// Deliberately Attribute-only, not generic over Node too, despite an earlier
// version of this design supporting both: giving one type unconditional
// Render and Apply methods regardless of T means every instance satisfies
// both Node and Attribute simultaneously, and createElement's switch always
// picks whichever case it checks first -- silently dropping the other,
// confirmed directly by a failing test (a TypedEvaluation[Node] passed to
// Div rendered as an empty element, its child completely missing, no error
// at any point). Constraining T to Attribute rules this out at compile
// time rather than working around it at runtime; conditional Nodes for
// heterogeneous elements already work fine via the existing, untouched
// Evaluation/If/IfElse, so there's no real gap to fill there.
type TypedEvaluation[T Attribute] struct {
	condition  bool
	trueValue  T
	falseValue T
	hasFalse   bool
}

// Apply implements Attribute, so TypedEvaluation[Attribute] (and any T that
// is itself an Attribute, including a nested TypedEvaluation) can be passed
// anywhere an Attribute is expected -- including to self-closing elements
// like Input, which only accept ...Attribute.
func (e TypedEvaluation[T]) Apply(el *Element) {
	if e.condition {
		e.trueValue.Apply(el)
	} else if e.hasFalse {
		e.falseValue.Apply(el)
	}
}

// IfT returns item if condition is true, otherwise applies nothing --
// type-safe sibling of (*Builder).If, usable anywhere T (typically Attribute
// or Node) is expected, including with self-closing elements.
//
// Correctness note, not a hypothetical: whether "otherwise applies nothing"
// is achieved by checking hasFalse explicitly, rather than checking
// falseValue against T's zero value, matters in practice, not just in
// theory. Confirmed directly: when T is inferred as a concrete struct type
// (a user's own custom Attribute implementation, passed without an
// explicit Attribute(...) conversion) rather than the Attribute interface
// itself, T's zero value is a real, non-nil value of that struct -- not
// nil -- so a zero-value comparison would have silently applied a
// zero-valued attribute on the untaken branch instead of applying nothing,
// with no error or panic to reveal it. hasFalse sidesteps this: it's set
// explicitly by IfT/IfElseT and never depends on what T happens to be.
func IfT[T Attribute](condition bool, item T) TypedEvaluation[T] {
	return TypedEvaluation[T]{condition: condition, trueValue: item}
}

// IfElseT returns trueValue if condition is true, otherwise falseValue --
// type-safe sibling of (*Builder).IfElse, usable anywhere T (typically
// Attribute or Node) is expected, including with self-closing elements.
func IfElseT[T Attribute](condition bool, trueValue, falseValue T) TypedEvaluation[T] {
	return TypedEvaluation[T]{condition: condition, trueValue: trueValue, falseValue: falseValue, hasFalse: true}
}

// Document structure elements

// Html creates an <html> element.
func (b *Builder) Html(children ...interface{}) Node {
	return b.createElement("html", false, children...)
}

// Head creates a <head> element.
func (b *Builder) Head(children ...interface{}) Node {
	return b.createElement("head", false, children...)
}

// Body creates a <body> element.
func (b *Builder) Body(children ...interface{}) Node {
	return b.createElement("body", false, children...)
}

// Title creates a <title> element.
func (b *Builder) Title(text string) Node {
	return b.createElement("title", false, text)
}

// Meta creates a <meta> element (self-closing).
func (b *Builder) Meta(attrs ...Attribute) Node {
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("meta", true, args...)
}

// Link creates a <link> element (self-closing).
func (b *Builder) Link(attrs ...Attribute) Node {
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("link", true, args...)
}

// Script creates a <script> element.
func (b *Builder) Script(children ...interface{}) Node {
	return b.createElement("script", false, children...)
}

// Style creates a <style> element.
func (b *Builder) Style(children ...interface{}) Node {
	return b.createElement("style", false, children...)
}

// Base creates a <base> element (self-closing).
func (b *Builder) Base(attrs ...Attribute) Node {
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("base", true, args...)
}

// Noscript creates a <noscript> element.
func (b *Builder) Noscript(children ...interface{}) Node {
	return b.createElement("noscript", false, children...)
}

// Sectioning elements

// Main creates a <main> element.
func (b *Builder) Main(children ...interface{}) Node {
	return b.createElement("main", false, children...)
}

// Header creates a <header> element.
func (b *Builder) Header(children ...interface{}) Node {
	return b.createElement("header", false, children...)
}

// Footer creates a <footer> element.
func (b *Builder) Footer(children ...interface{}) Node {
	return b.createElement("footer", false, children...)
}

// Nav creates a <nav> element.
func (b *Builder) Nav(children ...interface{}) Node {
	return b.createElement("nav", false, children...)
}

// Section creates a <section> element.
func (b *Builder) Section(children ...interface{}) Node {
	return b.createElement("section", false, children...)
}

// Article creates an <article> element.
func (b *Builder) Article(children ...interface{}) Node {
	return b.createElement("article", false, children...)
}

// Aside creates an <aside> element.
func (b *Builder) Aside(children ...interface{}) Node {
	return b.createElement("aside", false, children...)
}

// Heading elements (complete hierarchy)

// H1 creates an <h1> element.
func (b *Builder) H1(children ...interface{}) Node {
	return b.createElement("h1", false, children...)
}

// H2 creates an <h2> element.
func (b *Builder) H2(children ...interface{}) Node {
	return b.createElement("h2", false, children...)
}

// H3 creates an <h3> element.
func (b *Builder) H3(children ...interface{}) Node {
	return b.createElement("h3", false, children...)
}

// H4 creates an <h4> element.
func (b *Builder) H4(children ...interface{}) Node {
	return b.createElement("h4", false, children...)
}

// H5 creates an <h5> element.
func (b *Builder) H5(children ...interface{}) Node {
	return b.createElement("h5", false, children...)
}

// H6 creates an <h6> element.
func (b *Builder) H6(children ...interface{}) Node {
	return b.createElement("h6", false, children...)
}

// Hgroup creates an <hgroup> element.
func (b *Builder) Hgroup(children ...interface{}) Node {
	return b.createElement("hgroup", false, children...)
}

// Text content elements

// P creates a <p> element.
func (b *Builder) P(children ...interface{}) Node {
	return b.createElement("p", false, children...)
}

// Div creates a <div> element.
func (b *Builder) Div(children ...interface{}) Node {
	return b.createElement("div", false, children...)
}

// Span creates a <span> element.
func (b *Builder) Span(children ...interface{}) Node {
	return b.createElement("span", false, children...)
}

// Pre creates a <pre> element.
func (b *Builder) Pre(children ...interface{}) Node {
	return b.createElement("pre", false, children...)
}

// Blockquote creates a <blockquote> element.
func (b *Builder) Blockquote(children ...interface{}) Node {
	return b.createElement("blockquote", false, children...)
}

// Address creates an <address> element.
func (b *Builder) Address(children ...interface{}) Node {
	return b.createElement("address", false, children...)
}

// Hr creates an <hr> element (self-closing).
func (b *Builder) Hr(attrs ...Attribute) Node {
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("hr", true, args...)
}

// Br creates a <br> element (self-closing).
func (b *Builder) Br() Node {
	return b.createElement("br", true)
}

// Inline text semantics

// Strong creates a <strong> element.
func (b *Builder) Strong(children ...interface{}) Node {
	return b.createElement("strong", false, children...)
}

// Em creates an <em> element.
func (b *Builder) Em(children ...interface{}) Node {
	return b.createElement("em", false, children...)
}

// B creates a <b> element.
func (b *Builder) B(children ...interface{}) Node {
	return b.createElement("b", false, children...)
}

// I creates an <i> element.
func (b *Builder) I(children ...interface{}) Node {
	return b.createElement("i", false, children...)
}

// U creates a <u> element.
func (b *Builder) U(children ...interface{}) Node {
	return b.createElement("u", false, children...)
}

// S creates an <s> element.
func (b *Builder) S(children ...interface{}) Node {
	return b.createElement("s", false, children...)
}

// Small creates a <small> element.
func (b *Builder) Small(children ...interface{}) Node {
	return b.createElement("small", false, children...)
}

// Mark creates a <mark> element.
func (b *Builder) Mark(children ...interface{}) Node {
	return b.createElement("mark", false, children...)
}

// Del creates a <del> element.
func (b *Builder) Del(children ...interface{}) Node {
	return b.createElement("del", false, children...)
}

// Ins creates an <ins> element.
func (b *Builder) Ins(children ...interface{}) Node {
	return b.createElement("ins", false, children...)
}

// Sub creates a <sub> element.
func (b *Builder) Sub(children ...interface{}) Node {
	return b.createElement("sub", false, children...)
}

// Sup creates a <sup> element.
func (b *Builder) Sup(children ...interface{}) Node {
	return b.createElement("sup", false, children...)
}

// Code creates a <code> element.
func (b *Builder) Code(children ...interface{}) Node {
	return b.createElement("code", false, children...)
}

// Kbd creates a <kbd> element.
func (b *Builder) Kbd(children ...interface{}) Node {
	return b.createElement("kbd", false, children...)
}

// Samp creates a <samp> element.
func (b *Builder) Samp(children ...interface{}) Node {
	return b.createElement("samp", false, children...)
}

// Var creates a <var> element.
func (b *Builder) Var(children ...interface{}) Node {
	return b.createElement("var", false, children...)
}

// Cite creates a <cite> element.
func (b *Builder) Cite(children ...interface{}) Node {
	return b.createElement("cite", false, children...)
}

// Abbr creates an <abbr> element.
func (b *Builder) Abbr(children ...interface{}) Node {
	return b.createElement("abbr", false, children...)
}

// Dfn creates a <dfn> element.
func (b *Builder) Dfn(children ...interface{}) Node {
	return b.createElement("dfn", false, children...)
}

// Q creates a <q> element.
func (b *Builder) Q(children ...interface{}) Node {
	return b.createElement("q", false, children...)
}

// Time creates a <time> element.
func (b *Builder) Time(children ...interface{}) Node {
	return b.createElement("time", false, children...)
}

// Data creates a <data> element.
func (b *Builder) Data(children ...interface{}) Node {
	return b.createElement("data", false, children...)
}

// Wbr creates a <wbr> element (self-closing).
func (b *Builder) Wbr() Node {
	return b.createElement("wbr", true)
}

// Bdi creates a <bdi> element.
func (b *Builder) Bdi(children ...interface{}) Node {
	return b.createElement("bdi", false, children...)
}

// Bdo creates a <bdo> element.
func (b *Builder) Bdo(children ...interface{}) Node {
	return b.createElement("bdo", false, children...)
}

// Ruby creates a <ruby> element.
func (b *Builder) Ruby(children ...interface{}) Node {
	return b.createElement("ruby", false, children...)
}

// Rt creates an <rt> element.
func (b *Builder) Rt(children ...interface{}) Node {
	return b.createElement("rt", false, children...)
}

// Rp creates an <rp> element.
func (b *Builder) Rp(children ...interface{}) Node {
	return b.createElement("rp", false, children...)
}

// List elements

// Ul creates a <ul> element.
func (b *Builder) Ul(children ...interface{}) Node {
	return b.createElement("ul", false, children...)
}

// Ol creates an <ol> element.
func (b *Builder) Ol(children ...interface{}) Node {
	return b.createElement("ol", false, children...)
}

// Li creates an <li> element.
func (b *Builder) Li(children ...interface{}) Node {
	return b.createElement("li", false, children...)
}

// Dl creates a <dl> element.
func (b *Builder) Dl(children ...interface{}) Node {
	return b.createElement("dl", false, children...)
}

// Dt creates a <dt> element.
func (b *Builder) Dt(children ...interface{}) Node {
	return b.createElement("dt", false, children...)
}

// Dd creates a <dd> element.
func (b *Builder) Dd(children ...interface{}) Node {
	return b.createElement("dd", false, children...)
}

// Table elements

// Table creates a <table> element.
func (b *Builder) Table(children ...interface{}) Node {
	return b.createElement("table", false, children...)
}

// Thead creates a <thead> element.
func (b *Builder) Thead(children ...interface{}) Node {
	return b.createElement("thead", false, children...)
}

// Tbody creates a <tbody> element.
func (b *Builder) Tbody(children ...interface{}) Node {
	return b.createElement("tbody", false, children...)
}

// Tfoot creates a <tfoot> element.
func (b *Builder) Tfoot(children ...interface{}) Node {
	return b.createElement("tfoot", false, children...)
}

// Tr creates a <tr> element.
func (b *Builder) Tr(children ...interface{}) Node {
	return b.createElement("tr", false, children...)
}

// Th creates a <th> element.
func (b *Builder) Th(children ...interface{}) Node {
	return b.createElement("th", false, children...)
}

// Td creates a <td> element.
func (b *Builder) Td(children ...interface{}) Node {
	return b.createElement("td", false, children...)
}

// Caption creates a <caption> element.
func (b *Builder) Caption(children ...interface{}) Node {
	return b.createElement("caption", false, children...)
}

// Colgroup creates a <colgroup> element.
func (b *Builder) Colgroup(children ...interface{}) Node {
	return b.createElement("colgroup", false, children...)
}

// Col creates a <col> element (self-closing).
func (b *Builder) Col(attrs ...Attribute) Node {
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("col", true, args...)
}

// Form elements

// Form creates a <form> element.
func (b *Builder) Form(args ...interface{}) Node {
	return b.createElement("form", false, args...)
}

// Input creates an <input> element (self-closing).
func (b *Builder) Input(attrs ...Attribute) Node {
	// Convert to []interface{} for createElement
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("input", true, args...)
}

// Button creates a <button> element.
func (b *Builder) Button(args ...interface{}) Node {
	return b.createElement("button", false, args...)
}

// Label creates a <label> element.
func (b *Builder) Label(args ...interface{}) Node {
	return b.createElement("label", false, args...)
}

// Textarea creates a <textarea> element.
func (b *Builder) Textarea(args ...interface{}) Node {
	return b.createElement("textarea", false, args...)
}

// Select creates a <select> element.
func (b *Builder) Select(args ...interface{}) Node {
	return b.createElement("select", false, args...)
}

// Option creates an <option> element.
func (b *Builder) Option(args ...interface{}) Node {
	return b.createElement("option", false, args...)
}

// Optgroup creates an <optgroup> element.
func (b *Builder) Optgroup(args ...interface{}) Node {
	return b.createElement("optgroup", false, args...)
}

// Fieldset creates a <fieldset> element.
func (b *Builder) Fieldset(children ...interface{}) Node {
	return b.createElement("fieldset", false, children...)
}

// Legend creates a <legend> element.
func (b *Builder) Legend(children ...interface{}) Node {
	return b.createElement("legend", false, children...)
}

// Datalist creates a <datalist> element.
func (b *Builder) Datalist(children ...interface{}) Node {
	return b.createElement("datalist", false, children...)
}

// Output creates an <output> element.
func (b *Builder) Output(children ...interface{}) Node {
	return b.createElement("output", false, children...)
}

// Progress creates a <progress> element.
func (b *Builder) Progress(children ...interface{}) Node {
	return b.createElement("progress", false, children...)
}

// Meter creates a <meter> element.
func (b *Builder) Meter(children ...interface{}) Node {
	return b.createElement("meter", false, children...)
}

// Media elements

// Img creates an <img> element (self-closing).
func (b *Builder) Img(attrs ...Attribute) Node {
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("img", true, args...)
}

// Video creates a <video> element.
func (b *Builder) Video(children ...interface{}) Node {
	return b.createElement("video", false, children...)
}

// Audio creates an <audio> element.
func (b *Builder) Audio(children ...interface{}) Node {
	return b.createElement("audio", false, children...)
}

// Source creates a <source> element (self-closing).
func (b *Builder) Source(attrs ...Attribute) Node {
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("source", true, args...)
}

// Track creates a <track> element (self-closing).
func (b *Builder) Track(attrs ...Attribute) Node {
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("track", true, args...)
}

// Picture creates a <picture> element.
func (b *Builder) Picture(children ...interface{}) Node {
	return b.createElement("picture", false, children...)
}

// Canvas creates a <canvas> element.
func (b *Builder) Canvas(children ...interface{}) Node {
	return b.createElement("canvas", false, children...)
}

// Map creates a <map> element.
func (b *Builder) Map(children ...interface{}) Node {
	return b.createElement("map", false, children...)
}

// Area creates an <area> element (self-closing).
func (b *Builder) Area(attrs ...Attribute) Node {
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("area", true, args...)
}

// Svg creates an <svg> element.
func (b *Builder) Svg(children ...interface{}) Node {
	return b.createElement("svg", false, children...)
}

// Polygon creates an SVG <polygon> element.
func (b *Builder) Polygon(children ...interface{}) Node {
	return b.createElement("polygon", true, children...)
}

// Path creates an SVG <path> element.
func (b *Builder) Path(children ...interface{}) Node {
	return b.createElement("path", true, children...)
}

// Circle creates an SVG <circle> element.
func (b *Builder) Circle(children ...interface{}) Node {
	return b.createElement("circle", true, children...)
}

// Rect creates an SVG <rect> element.
func (b *Builder) Rect(children ...interface{}) Node {
	return b.createElement("rect", true, children...)
}

// Line creates an SVG <line> element.
func (b *Builder) Line(children ...interface{}) Node {
	return b.createElement("line", true, children...)
}

// G creates an SVG <g> (group) element.
func (b *Builder) G(children ...interface{}) Node {
	return b.createElement("g", false, children...)
}

// Math creates a <math> element.
func (b *Builder) Math(children ...interface{}) Node {
	return b.createElement("math", false, children...)
}

// Interactive elements

// Details creates a <details> element.
func (b *Builder) Details(children ...interface{}) Node {
	return b.createElement("details", false, children...)
}

// Summary creates a <summary> element.
func (b *Builder) Summary(children ...interface{}) Node {
	return b.createElement("summary", false, children...)
}

// Dialog creates a <dialog> element.
func (b *Builder) Dialog(children ...interface{}) Node {
	return b.createElement("dialog", false, children...)
}

// Menu creates a <menu> element.
func (b *Builder) Menu(children ...interface{}) Node {
	return b.createElement("menu", false, children...)
}

// Links and navigation

// A creates an <a> element.
func (b *Builder) A(args ...interface{}) Node {
	return b.createElement("a", false, args...)
}

// Embedded content

// Embed creates an <embed> element (self-closing).
func (b *Builder) Embed(attrs ...Attribute) Node {
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("embed", true, args...)
}

// Object creates an <object> element.
func (b *Builder) Object(children ...interface{}) Node {
	return b.createElement("object", false, children...)
}

// Param creates a <param> element (self-closing).
func (b *Builder) Param(attrs ...Attribute) Node {
	args := make([]interface{}, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return b.createElement("param", true, args...)
}

// Iframe creates an <iframe> element.
func (b *Builder) Iframe(children ...interface{}) Node {
	return b.createElement("iframe", false, children...)
}

// Web Components

// Template creates a <template> element.
func (b *Builder) Template(children ...interface{}) Node {
	return b.createElement("template", false, children...)
}

// Slot creates a <slot> element.
func (b *Builder) Slot(children ...interface{}) Node {
	return b.createElement("slot", false, children...)
}

// Additional SVG Elements

// Polyline creates an SVG <polyline> element.
func (b *Builder) Polyline(children ...interface{}) Node {
	return b.createElement("polyline", true, children...)
}

// Defs creates an SVG <defs> element.
func (b *Builder) Defs(children ...interface{}) Node {
	return b.createElement("defs", false, children...)
}

// Use creates an SVG <use> element.
func (b *Builder) Use(children ...interface{}) Node {
	return b.createElement("use", true, children...)
}

// SvgText creates an SVG <text> element.
func (b *Builder) SvgText(children ...interface{}) Node {
	return b.createElement("text", false, children...)
}

// Global builder instance using standard Minty alias pattern
var B = &Builder{}

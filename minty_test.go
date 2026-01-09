package minty

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// Smoke test - basic functionality works
func TestBasicRendering(t *testing.T) {
	template := func(b *Builder) Node {
		return b.Html(
			b.Head(b.Title("Test")),
			b.Body(b.H1("Hello World")),
		)
	}

	var buf bytes.Buffer
	if err := Render(template, &buf); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "<title>Test</title>") {
		t.Error("Missing title element")
	}
	if !strings.Contains(html, "<h1>Hello World</h1>") {
		t.Error("Missing h1 element")
	}
	if !strings.Contains(html, "<html>") || !strings.Contains(html, "</html>") {
		t.Error("Missing html wrapper")
	}
}

// Security regression test - HTML escaping
func TestHTMLEscaping(t *testing.T) {
	dangerous := "<script>alert('xss')</script>"

	template := func(b *Builder) Node {
		return b.P(dangerous)
	}

	var buf bytes.Buffer
	Render(template, &buf)
	html := buf.String()

	if strings.Contains(html, "<script>") {
		t.Error("HTML not properly escaped - security vulnerability!")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("Expected escaped HTML")
	}
}

// Interface regression test - attributes work
func TestAttributes(t *testing.T) {
	template := func(b *Builder) Node {
		return b.Div(Class("test-class"), ID("test-id"), "content")
	}

	var buf bytes.Buffer
	Render(template, &buf)
	html := buf.String()

	if !strings.Contains(html, `class="test-class"`) {
		t.Error("Class attribute not applied")
	}
	if !strings.Contains(html, `id="test-id"`) {
		t.Error("ID attribute not applied")
	}
	if !strings.Contains(html, "content") {
		t.Error("Text content not rendered")
	}
}

// Logic test - If works
func TestIf(t *testing.T) {
	template := func(b *Builder) Node {
		return b.Div(
			b.If(true, b.P("Condition is true")),
			b.If(false, b.P("Condition is false")),
		)
	}

	var buf bytes.Buffer
	Render(template, &buf)
	html := buf.String()

	if !strings.Contains(html, "<div><p>Condition is true</p></div>") {
		t.Errorf("If(true) should render content, got: %s", html)
	}
	if strings.Contains(html, "<div><p>Condition is false</p></div>") {
		t.Errorf("If(false) should not render content, got: %s", html)
	}
}

// Builder regression test - global instance works
func TestGlobalBuilder(t *testing.T) {
	template := func(b *Builder) Node {
		return b.P("test")
	}

	// Should work with global B instance
	node := template(B)
	if node == nil {
		t.Error("Global builder B not working")
	}

	var buf bytes.Buffer
	if err := node.Render(&buf); err != nil {
		t.Errorf("Global builder render failed: %v", err)
	}

	if !strings.Contains(buf.String(), "<p>test</p>") {
		t.Error("Global builder didn't produce expected output")
	}
}

// Link element test - navigation functionality
func TestLinkElements(t *testing.T) {
	template := func(b *Builder) Node {
		return b.Ul(
			b.Li(b.A(Href("/about"), "About")),
			b.Li(b.A(Href("/contact"), Target("_blank"), "Contact")),
		)
	}

	var buf bytes.Buffer
	Render(template, &buf)
	html := buf.String()

	if !strings.Contains(html, `href="/about"`) {
		t.Error("Href attribute not applied")
	}
	if !strings.Contains(html, `target="_blank"`) {
		t.Error("Target attribute not applied")
	}
	if !strings.Contains(html, "<ul>") || !strings.Contains(html, "</ul>") {
		t.Error("List structure not rendered")
	}
}

// Mixed content test - strings, nodes, and attributes
func TestMixedContent(t *testing.T) {
	template := func(b *Builder) Node {
		return b.Div(
			Class("container"),
			b.H1("Title"),
			"Some text",
			b.P("Paragraph content"),
		)
	}

	var buf bytes.Buffer
	Render(template, &buf)
	html := buf.String()

	if !strings.Contains(html, `class="container"`) {
		t.Error("Class attribute missing")
	}
	if !strings.Contains(html, "<h1>Title</h1>") {
		t.Error("H1 element missing")
	}
	if !strings.Contains(html, "Some text") {
		t.Error("Text content missing")
	}
	if !strings.Contains(html, "<p>Paragraph content</p>") {
		t.Error("Paragraph missing")
	}
}

// RenderToString utility test
func TestRenderToString(t *testing.T) {
	template := func(b *Builder) Node {
		return b.P("Hello World")
	}

	html := RenderToString(template)
	expected := "<p>Hello World</p>"

	if html != expected {
		t.Errorf("Expected %q, got %q", expected, html)
	}
}

// Week 2 Tests - Form Elements

func TestFormElements(t *testing.T) {
	template := func(b *Builder) Node {
		return b.Form(Action("/submit"), Method("POST"),
			b.Label(For("email"), "Email:"),
			b.Input(Type("email"), Name("email"), Required()),
			b.Button(Type("submit"), "Submit"),
		)
	}

	html := RenderToString(template)

	// Check form attributes
	if !strings.Contains(html, `action="/submit"`) {
		t.Error("Form action attribute missing")
	}
	if !strings.Contains(html, `method="POST"`) {
		t.Error("Form method attribute missing")
	}
	
	// Check input attributes
	if !strings.Contains(html, `type="email"`) {
		t.Error("Input type attribute missing")
	}
	if !strings.Contains(html, `name="email"`) {
		t.Error("Input name attribute missing")
	}
	if !strings.Contains(html, `required="required"`) {
		t.Error("Required boolean attribute missing")
	}
	
	// Check label
	if !strings.Contains(html, `for="email"`) {
		t.Error("Label for attribute missing")
	}
}

func TestTextareaAndSelect(t *testing.T) {
	template := func(b *Builder) Node {
		return b.Div(
			b.Textarea(Name("message"), Rows(5), Cols(40), "Default text"),
			b.Select(Name("country"), 
				b.Option(Value("us"), "United States"),
				b.Option(Value("ca"), Selected(), "Canada"),
			),
		)
	}

	html := RenderToString(template)
	
	if !strings.Contains(html, `<textarea`) {
		t.Error("Textarea element missing")
	}
	if !strings.Contains(html, `rows="5"`) {
		t.Error("Textarea rows attribute missing")
	}
	if !strings.Contains(html, `Default text`) {
		t.Error("Textarea content missing")
	}
	if !strings.Contains(html, `<select`) {
		t.Error("Select element missing")  
	}
	if !strings.Contains(html, `selected="selected"`) {
		t.Error("Selected option missing")
	}
}

// Week 2 Tests - Layout System

func TestBasicLayout(t *testing.T) {
	content := func(b *Builder) Node {
		return b.H1("Test Content")
	}
	
	template := Layout("Test Page", content)
	html := RenderToString(template)
	
	if !strings.Contains(html, "<html>") {
		t.Error("HTML wrapper missing from layout")
	}
	if !strings.Contains(html, "<title>Test Page</title>") {
		t.Error("Title missing from layout")
	}
	if !strings.Contains(html, `name="viewport"`) {
		t.Error("Viewport meta tag missing")
	}
	if !strings.Contains(html, "<h1>Test Content</h1>") {
		t.Error("Content not rendered in layout")
	}
}

func TestNavigation(t *testing.T) {
	navLinks := []NavLink{
		{URL: "/", Text: "Home"},
		{URL: "/about", Text: "About"},
	}
	
	template := Navigation(navLinks)
	html := RenderToString(template)
	
	if !strings.Contains(html, `href="/"`) {
		t.Error("Home link missing")
	}
	if !strings.Contains(html, `href="/about"`) {
		t.Error("About link missing")
	}
	if !strings.Contains(html, "Home") {
		t.Error("Home link text missing")
	}
	if !strings.Contains(html, " | ") {
		t.Error("Navigation separator missing")
	}
}

// Week 2 Tests - Validation

func TestValidation(t *testing.T) {
	// Test required validation
	if err := ValidateRequired("", "Name"); err == nil {
		t.Error("Required validation should fail for empty string")
	}
	if err := ValidateRequired("   ", "Name"); err == nil {
		t.Error("Required validation should fail for whitespace")
	}
	if err := ValidateRequired("valid", "Name"); err != nil {
		t.Error("Required validation should pass for non-empty string")
	}
	
	// Test email validation
	if err := ValidateEmail("invalid", "Email"); err == nil {
		t.Error("Email validation should fail for invalid email")
	}
	if err := ValidateEmail("test@example.com", "Email"); err != nil {
		t.Error("Email validation should pass for valid email")
	}
	if err := ValidateEmail("", "Email"); err != nil {
		t.Error("Email validation should pass for empty string (use Required separately)")
	}
}

func TestValidationResult(t *testing.T) {
	result := &ValidationResult{IsValid: true}
	
	result.AddError("email", "Invalid email")
	result.AddError("name", "Required field")
	
	if result.IsValid {
		t.Error("ValidationResult should be invalid after adding errors")
	}
	if len(result.Errors) != 2 {
		t.Error("ValidationResult should contain 2 errors")
	}
	if result.GetError("email") != "Invalid email" {
		t.Error("GetError should return correct error message")
	}
	if !result.HasError("name") {
		t.Error("HasError should return true for existing error")
	}
	if result.HasError("nonexistent") {
		t.Error("HasError should return false for non-existing error")
	}
}

// Week 2 Tests - Control Flow

func TestControlFlow(t *testing.T) {
	trueTemplate := func(b *Builder) Node {
		return b.P("True content")
	}
	falseTemplate := func(b *Builder) Node {
		return b.P("False content")
	}
	
	// Test If
	ifTrue := If(true, trueTemplate)
	html := RenderToString(ifTrue)
	if !strings.Contains(html, "True content") {
		t.Error("If(true) should render content")
	}
	
	ifFalse := If(false, trueTemplate)
	html = RenderToString(ifFalse)
	if strings.Contains(html, "True content") {
		t.Error("If(false) should not render content")
	}
	
	// Test IfElse
	ifElseTrue := IfElse(true, trueTemplate, falseTemplate)
	html = RenderToString(ifElseTrue)
	if !strings.Contains(html, "True content") {
		t.Error("IfElse(true) should render true template")
	}
	
	ifElseFalse := IfElse(false, trueTemplate, falseTemplate)
	html = RenderToString(ifElseFalse)
	if !strings.Contains(html, "False content") {
		t.Error("IfElse(false) should render false template")
	}
}

// Week 3 Tests - HTMX Integration

func TestHTMXAttributes(t *testing.T) {
	template := func(b *Builder) Node {
		return b.Div(
			b.Input(
				Type("text"),
				Name("search"),
				HtmxGet("/api/search"),
				HtmxTarget("#results"),
				HtmxTrigger("keyup changed delay:300ms"),
			),
			b.Button(
				HtmxPost("/api/submit"),
				HtmxConfirm("Are you sure?"),
				HtmxSwap("innerHTML"),
				"Submit",
			),
		)
	}

	html := RenderToString(template)

	// Check HTMX attributes
	if !strings.Contains(html, `hx-get="/api/search"`) {
		t.Error("HtmxGet attribute missing")
	}
	if !strings.Contains(html, `hx-target="#results"`) {
		t.Error("HtmxTarget attribute missing")
	}
	if !strings.Contains(html, `hx-trigger="keyup changed delay:300ms"`) {
		t.Error("HtmxTrigger attribute missing")
	}
	if !strings.Contains(html, `hx-post="/api/submit"`) {
		t.Error("HtmxPost attribute missing")
	}
	if !strings.Contains(html, `hx-confirm="Are you sure?"`) {
		t.Error("HtmxConfirm attribute missing")
	}
	if !strings.Contains(html, `hx-swap="innerHTML"`) {
		t.Error("HtmxSwap attribute missing")
	}
}

func TestHTMXHelperComponents(t *testing.T) {
	// Test LiveSearch
	liveSearch := LiveSearch("/search", "#results", "Search...")
	html := RenderToString(liveSearch)
	
	if !strings.Contains(html, `hx-get="/search"`) {
		t.Error("LiveSearch should include hx-get")
	}
	if !strings.Contains(html, `hx-target="#results"`) {
		t.Error("LiveSearch should include hx-target")
	}
	if !strings.Contains(html, `placeholder="Search..."`) {
		t.Error("LiveSearch should include placeholder")
	}

	// Test HTMXForm
	htmxForm := HTMXForm("POST", "/submit", "#target", 
		B.Input(Type("text"), Name("data")))
	html = RenderToString(htmxForm)
	
	if !strings.Contains(html, `method="POST"`) {
		t.Error("HTMXForm should include method")
	}
	if !strings.Contains(html, `hx-post="/submit"`) {
		t.Error("HTMXForm should include hx-post")
	}
	if !strings.Contains(html, `hx-target="#target"`) {
		t.Error("HTMXForm should include hx-target")
	}
}

func TestHTMXLoadingIndicators(t *testing.T) {
	indicator := HTMXLoadingIndicator("Loading...")
	html := RenderToString(indicator)
	
	if !strings.Contains(html, `class="htmx-indicator"`) {
		t.Error("Loading indicator should have htmx-indicator class")
	}
	if !strings.Contains(html, "Loading...") {
		t.Error("Loading indicator should contain text")
	}
	
	spinner := HTMXLoadingSpinner()
	html = RenderToString(spinner)
	
	if !strings.Contains(html, `class="htmx-indicator spinner"`) {
		t.Error("Loading spinner should have correct classes")
	}
	if !strings.Contains(html, "animation: spin") {
		t.Error("Loading spinner should include spin animation CSS")
	}
}

func TestHTMXSwapStrategies(t *testing.T) {
	strategies := HTMXSwapStrategies
	
	expected := map[string]string{
		strategies.InnerHTML:   "innerHTML",
		strategies.OuterHTML:   "outerHTML",
		strategies.BeforeBegin: "beforebegin",
		strategies.AfterBegin:  "afterbegin",
		strategies.BeforeEnd:   "beforeend",
		strategies.AfterEnd:    "afterend",
		strategies.Delete:      "delete",
		strategies.None:        "none",
	}
	
	for strategy, expectedValue := range expected {
		if strategy != expectedValue {
			t.Errorf("Expected swap strategy %s, got %s", expectedValue, strategy)
		}
	}
}

func TestHTMXTriggers(t *testing.T) {
	triggers := HTMXTriggers
	
	if triggers.KeyUpDelayed != "keyup changed delay:300ms" {
		t.Error("KeyUpDelayed trigger should have correct value")
	}
	if triggers.ChangeDelayed != "change delay:500ms" {
		t.Error("ChangeDelayed trigger should have correct value")
	}
	if triggers.Click != "click" {
		t.Error("Click trigger should be 'click'")
	}
}

func TestInfiniteScroll(t *testing.T) {
	infiniteScroll := InfiniteScroll("/api/load-more", "#content")
	html := RenderToString(infiniteScroll)
	
	if !strings.Contains(html, `hx-get="/api/load-more"`) {
		t.Error("InfiniteScroll should include hx-get")
	}
	if !strings.Contains(html, `hx-trigger="revealed"`) {
		t.Error("InfiniteScroll should use revealed trigger")
	}
	if !strings.Contains(html, `hx-target="#content"`) {
		t.Error("InfiniteScroll should include target")
	}
	if !strings.Contains(html, `hx-swap="beforeend"`) {
		t.Error("InfiniteScroll should use beforeend swap")
	}
}

// Week 4 Tests - Enhanced Control Flow

func TestEachFunction(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	
	renderer := func(item string) H {
		return func(b *Builder) Node {
			return b.Li(item)
		}
	}
	
	nodes := Each(items, renderer)
	
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}
	
	// Test with fragment
	template := func(b *Builder) Node {
		return b.Ul(NewFragment(Each(items, renderer)...))
	}
	
	html := RenderToString(template)
	
	if !strings.Contains(html, "<li>apple</li>") {
		t.Error("Each should render first item")
	}
	if !strings.Contains(html, "<li>banana</li>") {
		t.Error("Each should render second item")
	}
	if !strings.Contains(html, "<li>cherry</li>") {
		t.Error("Each should render third item")
	}
}

func TestEachWithEmptySlice(t *testing.T) {
	var items []string
	
	renderer := func(item string) H {
		return func(b *Builder) Node {
			return b.Li(item)
		}
	}
	
	nodes := Each(items, renderer)
	
	if len(nodes) != 0 {
		t.Error("Each should return empty slice for empty input")
	}
}

func TestEachWithIndex(t *testing.T) {
	items := []string{"first", "second"}
	
	renderer := func(i int, item string) H {
		return func(b *Builder) Node {
			return b.Li(fmt.Sprintf("%d: %s", i, item))
		}
	}
	
	nodes := EachWithIndex(items, renderer)
	
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodes))
	}
	
	template := func(b *Builder) Node {
		return NewFragment(nodes...)
	}
	
	html := RenderToString(template)
	
	if !strings.Contains(html, "0: first") {
		t.Error("EachWithIndex should include index in output")
	}
	if !strings.Contains(html, "1: second") {
		t.Error("EachWithIndex should include correct index")
	}
}

func TestFilterFunction(t *testing.T) {
	type User struct {
		Name   string
		Active bool
	}
	
	users := []User{
		{"Alice", true},
		{"Bob", false},
		{"Charlie", true},
	}
	
	predicate := func(u User) bool {
		return u.Active
	}
	
	renderer := func(u User) H {
		return func(b *Builder) Node {
			return b.P(u.Name)
		}
	}
	
	nodes := Filter(users, predicate, renderer)
	
	if len(nodes) != 2 {
		t.Errorf("Expected 2 active users, got %d", len(nodes))
	}
	
	template := func(b *Builder) Node {
		return NewFragment(nodes...)
	}
	
	html := RenderToString(template)
	
	if !strings.Contains(html, "Alice") {
		t.Error("Filter should include Alice (active)")
	}
	if strings.Contains(html, "Bob") {
		t.Error("Filter should exclude Bob (inactive)")
	}
	if !strings.Contains(html, "Charlie") {
		t.Error("Filter should include Charlie (active)")
	}
}

func TestMapFunction(t *testing.T) {
	numbers := []int{1, 2, 3}
	
	renderer := func(n int) H {
		return func(b *Builder) Node {
			return b.Span(fmt.Sprintf("Number: %d", n))
		}
	}
	
	template := Map(numbers, renderer)
	html := RenderToString(template)
	
	if !strings.Contains(html, "Number: 1") {
		t.Error("Map should render first number")
	}
	if !strings.Contains(html, "Number: 2") {
		t.Error("Map should render second number")
	}
	if !strings.Contains(html, "Number: 3") {
		t.Error("Map should render third number")
	}
}

func TestRangeFunction(t *testing.T) {
	renderer := func(i int) H {
		return func(b *Builder) Node {
			return b.Span(fmt.Sprintf("Item %d", i))
		}
	}
	
	nodes := Range(0, 3, renderer)
	
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}
	
	template := func(b *Builder) Node {
		return NewFragment(nodes...)
	}
	
	html := RenderToString(template)
	
	if !strings.Contains(html, "Item 0") {
		t.Error("Range should start from 0")
	}
	if !strings.Contains(html, "Item 2") {
		t.Error("Range should include 2")
	}
	if strings.Contains(html, "Item 3") {
		t.Error("Range should exclude end value")
	}
}

func TestWhenFunction(t *testing.T) {
	status := "active"
	
	cases := []WhenCase[string]{
		{"active", func(b *Builder) Node { return b.Span("User is active") }},
		{"inactive", func(b *Builder) Node { return b.Span("User is inactive") }},
		{"pending", func(b *Builder) Node { return b.Span("User is pending") }},
	}
	
	defaultTemplate := func(b *Builder) Node {
		return b.Span("Unknown status")
	}
	
	template := When(status, cases, defaultTemplate)
	html := RenderToString(template)
	
	if !strings.Contains(html, "User is active") {
		t.Error("When should match active status")
	}
	
	// Test default case
	unknownStatus := "deleted"
	defaultTemplate2 := When(unknownStatus, cases, defaultTemplate)
	html2 := RenderToString(defaultTemplate2)
	
	if !strings.Contains(html2, "Unknown status") {
		t.Error("When should use default template for unmatched value")
	}
}

func TestJoinFunction(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	
	renderer := func(item string) H {
		return func(b *Builder) Node {
			return b.Span(item)
		}
	}
	
	separator := func(b *Builder) Node {
		return b.Text(", ")
	}
	
	template := Join(items, renderer, separator)
	html := RenderToString(template)
	
	expected := "<span>apple</span>, <span>banana</span>, <span>cherry</span>"
	if html != expected {
		t.Errorf("Expected %q, got %q", expected, html)
	}
}

func TestRepeatFunction(t *testing.T) {
	template := func(b *Builder) Node {
		return b.Span("★")
	}
	
	nodes := Repeat(3, template)
	
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}
	
	fragmentTemplate := func(b *Builder) Node {
		return NewFragment(nodes...)
	}
	
	html := RenderToString(fragmentTemplate)
	expected := "<span>★</span><span>★</span><span>★</span>"
	
	if html != expected {
		t.Errorf("Expected %q, got %q", expected, html)
	}
}

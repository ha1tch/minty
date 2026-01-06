// Package minty provides ultra-concise HTML generation for Go web applications.
// Import with: import mi "github.com/ha1tch/minty"
package minty

import (
	"fmt"
	"html"
	"io"
	"slices"
	"strings"
)

// Node represents any HTML content that can be rendered.
type Node interface {
	Render(w io.Writer) error
}

// H represents a template function that generates HTML.
type H func(*Builder) Node

// Element represents an HTML tag with attributes and children.
type Element struct {
	Tag         string
	Attributes  map[string]string
	Children    []Node
	SelfClosing bool
}

// Render outputs the element as HTML.
func (e *Element) Render(w io.Writer) error {
	// Write opening tag
	if _, err := w.Write([]byte("<" + e.Tag)); err != nil {
		return err
	}

	// Sort attributes for consistent output
	keys := make([]string, 0, len(e.Attributes))
	for key := range e.Attributes {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	// Write attributes
	for _, key := range keys {
		if _, err := fmt.Fprintf(w, ` %s="%s"`, key, html.EscapeString(e.Attributes[key])); err != nil {
			return err
		}
	}

	if e.SelfClosing {
		_, err := w.Write([]byte(" />"))
		return err
	}

	if _, err := w.Write([]byte(">")); err != nil {
		return err
	}

	// Render children
	for _, child := range e.Children {
		if err := child.Render(w); err != nil {
			return err
		}
	}

	// Write closing tag
	_, err := fmt.Fprintf(w, "</%s>", e.Tag)
	return err
}

// TextNode represents escaped text content.
type TextNode struct {
	Content string
}

// Render outputs the text with HTML escaping.
func (t *TextNode) Render(w io.Writer) error {
	escaped := html.EscapeString(t.Content)
	_, err := w.Write([]byte(escaped))
	return err
}

// RawNode represents unescaped HTML content (use with caution).
type RawNode struct {
	Content string
}

// Render outputs unescaped content.
func (r *RawNode) Render(w io.Writer) error {
	_, err := w.Write([]byte(r.Content))
	return err
}

// Fragment represents a collection of nodes without a wrapper element.
type Fragment struct {
	Children []Node
}

// Render outputs all child nodes.
func (f *Fragment) Render(w io.Writer) error {
	for _, child := range f.Children {
		if err := child.Render(w); err != nil {
			return err
		}
	}
	return nil
}

// Attribute represents an HTML attribute that can be applied to elements.
type Attribute interface {
	Apply(*Element)
}

// StringAttribute represents an attribute with a string value.
type StringAttribute struct {
	Name  string
	Value string
}

// Apply adds the string attribute to an element.
func (sa StringAttribute) Apply(element *Element) {
	if element.Attributes == nil {
		element.Attributes = make(map[string]string)
	}
	element.Attributes[sa.Name] = sa.Value
}

// BooleanAttribute represents a boolean HTML attribute.
type BooleanAttribute struct {
	Name string
}

// Apply adds the boolean attribute to an element.
func (ba BooleanAttribute) Apply(element *Element) {
	if element.Attributes == nil {
		element.Attributes = make(map[string]string)
	}
	element.Attributes[ba.Name] = ba.Name
}

// Raw creates a node with unescaped HTML content.
func Raw(content string) Node {
	return &RawNode{Content: content}
}

// NewFragment creates a fragment containing the given nodes.
func NewFragment(nodes ...Node) Node {
	return &Fragment{Children: nodes}
}

// Text creates a text node (for explicit text handling).
func (b *Builder) Text(content string) Node {
	return &TextNode{Content: content}
}

// Render renders a template to the provided writer.
func Render(template H, w io.Writer) error {
	node := template(B)
	return node.Render(w)
}

// RenderToString renders a template and returns the HTML as a string.
func RenderToString(template H) string {
	var buf strings.Builder
	if err := Render(template, &buf); err != nil {
		return ""
	}
	return buf.String()
}

// Txt creates a text node (standalone function, alias for b.Text).
func Txt(content string) Node {
	return &TextNode{Content: content}
}

// RawHTML is an alias for Raw, creates unescaped HTML content.
func RawHTML(content string) Node {
	return Raw(content)
}

// Document creates a complete HTML document structure.
// It takes a title string, head nodes slice, and body node.
func Document(title string, headNodes []Node, body Node) H {
	return func(b *Builder) Node {
		// Build head content with title and additional nodes
		headChildren := []interface{}{
			b.Title(title),
			b.Meta(Charset("UTF-8")),
			b.Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
		}
		for _, node := range headNodes {
			headChildren = append(headChildren, node)
		}

		return NewFragment(
			Raw("<!DOCTYPE html>"),
			b.Html(Lang("en"),
				b.Head(headChildren...),
				body,
			),
		)
	}
}

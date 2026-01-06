package minty

import "fmt"

// Core attribute helper functions for common HTML attributes.
// These functions return Attribute instances that can be applied to elements.

// Universal attributes

// Attr creates an arbitrary attribute with the given name and value.
// Use this for attributes not covered by specific helper functions,
// such as SVG attributes or custom attributes.
func Attr(name, value string) Attribute {
	return StringAttribute{Name: name, Value: value}
}

// Class creates a class attribute.
func Class(value string) Attribute {
	return StringAttribute{Name: "class", Value: value}
}

// ID creates an id attribute.
func ID(value string) Attribute {
	return StringAttribute{Name: "id", Value: value}
}

// Style creates a style attribute.
func Style(value string) Attribute {
	return StringAttribute{Name: "style", Value: value}
}

// Title creates a title attribute.
func Title(value string) Attribute {
	return StringAttribute{Name: "title", Value: value}
}

// Dir creates a dir attribute for text direction.
func Dir(value string) Attribute {
	return StringAttribute{Name: "dir", Value: value}
}

// TabIndex creates a tabindex attribute.
func TabIndex(value int) Attribute {
	return StringAttribute{Name: "tabindex", Value: fmt.Sprintf("%d", value)}
}

// AccessKey creates an accesskey attribute.
func AccessKey(value string) Attribute {
	return StringAttribute{Name: "accesskey", Value: value}
}

// ContentEditable creates a contenteditable attribute.
func ContentEditable(value bool) Attribute {
	if value {
		return StringAttribute{Name: "contenteditable", Value: "true"}
	}
	return StringAttribute{Name: "contenteditable", Value: "false"}
}

// Hidden creates a hidden boolean attribute.
func Hidden() Attribute {
	return BooleanAttribute{Name: "hidden"}
}

// Spellcheck creates a spellcheck attribute.
func Spellcheck(value bool) Attribute {
	if value {
		return StringAttribute{Name: "spellcheck", Value: "true"}
	}
	return StringAttribute{Name: "spellcheck", Value: "false"}
}

// Translate creates a translate attribute.
func Translate(value bool) Attribute {
	if value {
		return StringAttribute{Name: "translate", Value: "yes"}
	}
	return StringAttribute{Name: "translate", Value: "no"}
}

// Link and navigation attributes

// Href creates an href attribute for links.
func Href(url string) Attribute {
	return StringAttribute{Name: "href", Value: url}
}

// Target creates a target attribute for links.
func Target(value string) Attribute {
	return StringAttribute{Name: "target", Value: value}
}

// Rel creates a rel attribute for link relationships.
func Rel(value string) Attribute {
	return StringAttribute{Name: "rel", Value: value}
}

// Download creates a download attribute.
func Download(filename string) Attribute {
	if filename == "" {
		return BooleanAttribute{Name: "download"}
	}
	return StringAttribute{Name: "download", Value: filename}
}

// Hreflang creates an hreflang attribute.
func Hreflang(value string) Attribute {
	return StringAttribute{Name: "hreflang", Value: value}
}

// Referrerpolicy creates a referrerpolicy attribute.
func Referrerpolicy(value string) Attribute {
	return StringAttribute{Name: "referrerpolicy", Value: value}
}

// Data creates a data-* attribute.
func Data(name, value string) Attribute {
	return StringAttribute{Name: "data-" + name, Value: value}
}

// DataAttr is an alias for Data, creates a data-* attribute.
func DataAttr(name, value string) Attribute {
	return Data(name, value)
}

// Meta attributes

// Name creates a name attribute.
func Name(value string) Attribute {
	return StringAttribute{Name: "name", Value: value}
}

// Content creates a content attribute.
func Content(value string) Attribute {
	return StringAttribute{Name: "content", Value: value}
}

// Lang creates a lang attribute.
func Lang(value string) Attribute {
	return StringAttribute{Name: "lang", Value: value}
}

// Charset creates a charset attribute.
func Charset(value string) Attribute {
	return StringAttribute{Name: "charset", Value: value}
}

// HttpEquiv creates an http-equiv attribute.
func HttpEquiv(value string) Attribute {
	return StringAttribute{Name: "http-equiv", Value: value}
}

// Form attributes

// Action creates an action attribute for forms.
func Action(url string) Attribute {
	return StringAttribute{Name: "action", Value: url}
}

// Method creates a method attribute for forms.
func Method(value string) Attribute {
	return StringAttribute{Name: "method", Value: value}
}

// Type creates a type attribute for input elements.
func Type(value string) Attribute {
	return StringAttribute{Name: "type", Value: value}
}

// Value creates a value attribute.
func Value(value string) Attribute {
	return StringAttribute{Name: "value", Value: value}
}

// For creates a for attribute for labels.
func For(value string) Attribute {
	return StringAttribute{Name: "for", Value: value}
}

// Placeholder creates a placeholder attribute.
func Placeholder(value string) Attribute {
	return StringAttribute{Name: "placeholder", Value: value}
}

// Accept creates an accept attribute.
func Accept(value string) Attribute {
	return StringAttribute{Name: "accept", Value: value}
}

// Autocomplete creates an autocomplete attribute.
func Autocomplete(value string) Attribute {
	return StringAttribute{Name: "autocomplete", Value: value}
}

// Enctype creates an enctype attribute.
func Enctype(value string) Attribute {
	return StringAttribute{Name: "enctype", Value: value}
}

// Novalidate creates a novalidate boolean attribute.
func Novalidate() Attribute {
	return BooleanAttribute{Name: "novalidate"}
}

// Boolean attributes

// Required creates a required boolean attribute.
func Required() Attribute {
	return BooleanAttribute{Name: "required"}
}

// Disabled creates a disabled boolean attribute.
func Disabled() Attribute {
	return BooleanAttribute{Name: "disabled"}
}

// Checked creates a checked boolean attribute.
func Checked() Attribute {
	return BooleanAttribute{Name: "checked"}
}

// Multiple creates a multiple boolean attribute.
func Multiple() Attribute {
	return BooleanAttribute{Name: "multiple"}
}

// Readonly creates a readonly boolean attribute.
func Readonly() Attribute {
	return BooleanAttribute{Name: "readonly"}
}

// Autofocus creates an autofocus boolean attribute.
func Autofocus() Attribute {
	return BooleanAttribute{Name: "autofocus"}
}

// Autoplay creates an autoplay boolean attribute.
func Autoplay() Attribute {
	return BooleanAttribute{Name: "autoplay"}
}

// Controls creates a controls boolean attribute.
func Controls() Attribute {
	return BooleanAttribute{Name: "controls"}
}

// Loop creates a loop boolean attribute.
func Loop() Attribute {
	return BooleanAttribute{Name: "loop"}
}

// Muted creates a muted boolean attribute.
func Muted() Attribute {
	return BooleanAttribute{Name: "muted"}
}

// Open creates an open boolean attribute.
func Open() Attribute {
	return BooleanAttribute{Name: "open"}
}

// Selected creates a selected boolean attribute.
func Selected() Attribute {
	return BooleanAttribute{Name: "selected"}
}

// Defer creates a defer boolean attribute.
func Defer() Attribute {
	return BooleanAttribute{Name: "defer"}
}

// Async creates an async boolean attribute.
func Async() Attribute {
	return BooleanAttribute{Name: "async"}
}

// Additional useful form attributes

// Rows creates a rows attribute for textarea.
func Rows(value int) Attribute {
	return StringAttribute{Name: "rows", Value: fmt.Sprintf("%d", value)}
}

// Cols creates a cols attribute for textarea.
func Cols(value int) Attribute {
	return StringAttribute{Name: "cols", Value: fmt.Sprintf("%d", value)}
}

// MaxLength creates a maxlength attribute.
func MaxLength(value int) Attribute {
	return StringAttribute{Name: "maxlength", Value: fmt.Sprintf("%d", value)}
}

// MinLength creates a minlength attribute.
func MinLength(value int) Attribute {
	return StringAttribute{Name: "minlength", Value: fmt.Sprintf("%d", value)}
}

// Size creates a size attribute.
func Size(value int) Attribute {
	return StringAttribute{Name: "size", Value: fmt.Sprintf("%d", value)}
}

// Min creates a min attribute.
func Min(value interface{}) Attribute {
	return StringAttribute{Name: "min", Value: fmt.Sprintf("%v", value)}
}

// Max creates a max attribute.
func Max(value interface{}) Attribute {
	return StringAttribute{Name: "max", Value: fmt.Sprintf("%v", value)}
}

// Step creates a step attribute.
func Step(value interface{}) Attribute {
	return StringAttribute{Name: "step", Value: fmt.Sprintf("%v", value)}
}

// Pattern creates a pattern attribute.
func Pattern(value string) Attribute {
	return StringAttribute{Name: "pattern", Value: value}
}

// Form creates a form attribute.
func Form(value string) Attribute {
	return StringAttribute{Name: "form", Value: value}
}

// Formaction creates a formaction attribute.
func Formaction(value string) Attribute {
	return StringAttribute{Name: "formaction", Value: value}
}

// Formenctype creates a formenctype attribute.
func Formenctype(value string) Attribute {
	return StringAttribute{Name: "formenctype", Value: value}
}

// Formmethod creates a formmethod attribute.
func Formmethod(value string) Attribute {
	return StringAttribute{Name: "formmethod", Value: value}
}

// Formnovalidate creates a formnovalidate boolean attribute.
func Formnovalidate() Attribute {
	return BooleanAttribute{Name: "formnovalidate"}
}

// Formtarget creates a formtarget attribute.
func Formtarget(value string) Attribute {
	return StringAttribute{Name: "formtarget", Value: value}
}

// List creates a list attribute.
func List(value string) Attribute {
	return StringAttribute{Name: "list", Value: value}
}

// Media attributes

// Src creates a src attribute for images and media elements.
func Src(url string) Attribute {
	return StringAttribute{Name: "src", Value: url}
}

// Alt creates an alt attribute for images.
func Alt(text string) Attribute {
	return StringAttribute{Name: "alt", Value: text}
}

// Width creates a width attribute.
func Width(value interface{}) Attribute {
	return StringAttribute{Name: "width", Value: fmt.Sprintf("%v", value)}
}

// Height creates a height attribute.
func Height(value interface{}) Attribute {
	return StringAttribute{Name: "height", Value: fmt.Sprintf("%v", value)}
}

// Crossorigin creates a crossorigin attribute.
func Crossorigin(value string) Attribute {
	return StringAttribute{Name: "crossorigin", Value: value}
}

// Poster creates a poster attribute for video.
func Poster(url string) Attribute {
	return StringAttribute{Name: "poster", Value: url}
}

// Preload creates a preload attribute.
func Preload(value string) Attribute {
	return StringAttribute{Name: "preload", Value: value}
}

// Srcset creates a srcset attribute.
func Srcset(value string) Attribute {
	return StringAttribute{Name: "srcset", Value: value}
}

// Sizes creates a sizes attribute.
func Sizes(value string) Attribute {
	return StringAttribute{Name: "sizes", Value: value}
}

// Media creates a media attribute.
func Media(value string) Attribute {
	return StringAttribute{Name: "media", Value: value}
}

// Table attributes

// Colspan creates a colspan attribute.
func Colspan(value int) Attribute {
	return StringAttribute{Name: "colspan", Value: fmt.Sprintf("%d", value)}
}

// Rowspan creates a rowspan attribute.
func Rowspan(value int) Attribute {
	return StringAttribute{Name: "rowspan", Value: fmt.Sprintf("%d", value)}
}

// Headers creates a headers attribute.
func Headers(value string) Attribute {
	return StringAttribute{Name: "headers", Value: value}
}

// Scope creates a scope attribute.
func Scope(value string) Attribute {
	return StringAttribute{Name: "scope", Value: value}
}

// Abbr creates an abbr attribute.
func AbbrAttr(value string) Attribute {
	return StringAttribute{Name: "abbr", Value: value}
}

// Interactive element attributes

// Datetime creates a datetime attribute.
func Datetime(value string) Attribute {
	return StringAttribute{Name: "datetime", Value: value}
}

// Cite creates a cite attribute.
func CiteAttr(value string) Attribute {
	return StringAttribute{Name: "cite", Value: value}
}

// High creates a high attribute for meter.
func High(value float64) Attribute {
	return StringAttribute{Name: "high", Value: fmt.Sprintf("%g", value)}
}

// Low creates a low attribute for meter.
func Low(value float64) Attribute {
	return StringAttribute{Name: "low", Value: fmt.Sprintf("%g", value)}
}

// Optimum creates an optimum attribute for meter.
func Optimum(value float64) Attribute {
	return StringAttribute{Name: "optimum", Value: fmt.Sprintf("%g", value)}
}

// HTMX Core Attributes

// HtmxGet creates an hx-get attribute for HTMX GET requests.
func HtmxGet(url string) Attribute {
	return StringAttribute{Name: "hx-get", Value: url}
}

// HtmxPost creates an hx-post attribute for HTMX POST requests.
func HtmxPost(url string) Attribute {
	return StringAttribute{Name: "hx-post", Value: url}
}

// HtmxPut creates an hx-put attribute for HTMX PUT requests.
func HtmxPut(url string) Attribute {
	return StringAttribute{Name: "hx-put", Value: url}
}

// HtmxDelete creates an hx-delete attribute for HTMX DELETE requests.
func HtmxDelete(url string) Attribute {
	return StringAttribute{Name: "hx-delete", Value: url}
}

// HtmxPatch creates an hx-patch attribute for HTMX PATCH requests.
func HtmxPatch(url string) Attribute {
	return StringAttribute{Name: "hx-patch", Value: url}
}

// HTMX Targeting and Swapping

// HtmxTarget creates an hx-target attribute to specify where to place response.
func HtmxTarget(selector string) Attribute {
	return StringAttribute{Name: "hx-target", Value: selector}
}

// HtmxSwap creates an hx-swap attribute to specify how to swap content.
func HtmxSwap(strategy string) Attribute {
	return StringAttribute{Name: "hx-swap", Value: strategy}
}

// HtmxSwapOOB creates an hx-swap-oob attribute for out-of-band swaps.
func HtmxSwapOOB(value string) Attribute {
	return StringAttribute{Name: "hx-swap-oob", Value: value}
}

// HTMX Triggering

// HtmxTrigger creates an hx-trigger attribute to specify what triggers the request.
func HtmxTrigger(trigger string) Attribute {
	return StringAttribute{Name: "hx-trigger", Value: trigger}
}

// HtmxOn creates an hx-on:* attribute to specify event handlers. The hx-on*
// attributes allow you to embed scripts inline to respond to events directly on
// an element; similar to the onevent properties found in HTML, such as onClick.
func HtmxOn(event, script string) Attribute {
	return StringAttribute{Name: "hx-on:" + event, Value: script}
}

// HTMX Indicators and Feedback

// HtmxIndicator creates an hx-indicator attribute to show loading indicators.
func HtmxIndicator(selector string) Attribute {
	return StringAttribute{Name: "hx-indicator", Value: selector}
}

// HtmxLoadingClass creates an htmx-indicator class for loading states.
func HtmxLoadingClass() Attribute {
	return StringAttribute{Name: "class", Value: "htmx-indicator"}
}

// HTMX Confirmation and Prompts

// HtmxConfirm creates an hx-confirm attribute for confirmation dialogs.
func HtmxConfirm(message string) Attribute {
	return StringAttribute{Name: "hx-confirm", Value: message}
}

// HtmxPrompt creates an hx-prompt attribute for user input prompts.
func HtmxPrompt(message string) Attribute {
	return StringAttribute{Name: "hx-prompt", Value: message}
}

// HTMX Headers and Parameters

// HtmxHeaders creates an hx-headers attribute for custom headers.
func HtmxHeaders(headers string) Attribute {
	return StringAttribute{Name: "hx-headers", Value: headers}
}

// HtmxVals creates an hx-vals attribute for additional parameters.
func HtmxVals(values string) Attribute {
	return StringAttribute{Name: "hx-vals", Value: values}
}

// HtmxInclude creates an hx-include attribute to include additional form data.
func HtmxInclude(selector string) Attribute {
	return StringAttribute{Name: "hx-include", Value: selector}
}

// HTMX History and Navigation

// HtmxPushURL creates an hx-push-url attribute for browser history.
func HtmxPushURL(url string) Attribute {
	return StringAttribute{Name: "hx-push-url", Value: url}
}

// HtmxReplaceURL creates an hx-replace-url attribute to replace current URL.
func HtmxReplaceURL(url string) Attribute {
	return StringAttribute{Name: "hx-replace-url", Value: url}
}

// HTMX Synchronization

// HtmxSync creates an hx-sync attribute for request synchronization.
func HtmxSync(value string) Attribute {
	return StringAttribute{Name: "hx-sync", Value: value}
}

// HTMX Extensions

// HtmxExt creates an hx-ext attribute for HTMX extensions.
func HtmxExt(extensions string) Attribute {
	return StringAttribute{Name: "hx-ext", Value: extensions}
}

// HTMX Boost

// HtmxBoost creates an hx-boost attribute for progressive enhancement.
func HtmxBoost() Attribute {
	return BooleanAttribute{Name: "hx-boost"}
}

// HtmxPreserve creates an hx-preserve attribute to preserve elements during swaps.
func HtmxPreserve() Attribute {
	return BooleanAttribute{Name: "hx-preserve"}
}

// ARIA attributes for accessibility

// AriaLabel creates an aria-label attribute.
func AriaLabel(value string) Attribute {
	return StringAttribute{Name: "aria-label", Value: value}
}

// AriaLabelledby creates an aria-labelledby attribute.
func AriaLabelledby(value string) Attribute {
	return StringAttribute{Name: "aria-labelledby", Value: value}
}

// AriaDescribedby creates an aria-describedby attribute.
func AriaDescribedby(value string) Attribute {
	return StringAttribute{Name: "aria-describedby", Value: value}
}

// AriaHidden creates an aria-hidden attribute.
func AriaHidden(value bool) Attribute {
	return StringAttribute{Name: "aria-hidden", Value: fmt.Sprintf("%t", value)}
}

// AriaExpanded creates an aria-expanded attribute.
func AriaExpanded(value bool) Attribute {
	return StringAttribute{Name: "aria-expanded", Value: fmt.Sprintf("%t", value)}
}

// AriaSelected creates an aria-selected attribute.
func AriaSelected(value bool) Attribute {
	return StringAttribute{Name: "aria-selected", Value: fmt.Sprintf("%t", value)}
}

// AriaChecked creates an aria-checked attribute.
func AriaChecked(value bool) Attribute {
	return StringAttribute{Name: "aria-checked", Value: fmt.Sprintf("%t", value)}
}

// AriaDisabled creates an aria-disabled attribute.
func AriaDisabled(value bool) Attribute {
	return StringAttribute{Name: "aria-disabled", Value: fmt.Sprintf("%t", value)}
}

// Role creates a role attribute.
func Role(value string) Attribute {
	return StringAttribute{Name: "role", Value: value}
}

// HTMX shorthand aliases (Hx* for brevity)

// HxGet is an alias for HtmxGet
func HxGet(url string) Attribute { return HtmxGet(url) }

// HxPost is an alias for HtmxPost
func HxPost(url string) Attribute { return HtmxPost(url) }

// HxPut is an alias for HtmxPut
func HxPut(url string) Attribute { return HtmxPut(url) }

// HxDelete is an alias for HtmxDelete
func HxDelete(url string) Attribute { return HtmxDelete(url) }

// HxPatch is an alias for HtmxPatch
func HxPatch(url string) Attribute { return HtmxPatch(url) }

// HxTarget is an alias for HtmxTarget
func HxTarget(selector string) Attribute { return HtmxTarget(selector) }

// HxSwap is an alias for HtmxSwap
func HxSwap(strategy string) Attribute { return HtmxSwap(strategy) }

// HxTrigger is an alias for HtmxTrigger
func HxTrigger(trigger string) Attribute { return HtmxTrigger(trigger) }

// HxOn is an alias for HtmxOn
func HxOn(event, script string) Attribute { return HtmxOn(event, script) }

// HxIndicator is an alias for HtmxIndicator
func HxIndicator(selector string) Attribute { return HtmxIndicator(selector) }

// HxConfirm is an alias for HtmxConfirm
func HxConfirm(message string) Attribute { return HtmxConfirm(message) }

// HxVals is an alias for HtmxVals
func HxVals(values string) Attribute { return HtmxVals(values) }

// HxInclude is an alias for HtmxInclude
func HxInclude(selector string) Attribute { return HtmxInclude(selector) }

// =====================================================
// SVG ATTRIBUTES
// =====================================================

// ViewBox creates a viewBox attribute for SVG elements.
func ViewBox(value string) Attribute {
	return StringAttribute{Name: "viewBox", Value: value}
}

// Fill creates a fill attribute for SVG elements.
func Fill(value string) Attribute {
	return StringAttribute{Name: "fill", Value: value}
}

// Stroke creates a stroke attribute for SVG elements.
func Stroke(value string) Attribute {
	return StringAttribute{Name: "stroke", Value: value}
}

// StrokeWidth creates a stroke-width attribute for SVG elements.
func StrokeWidth(value string) Attribute {
	return StringAttribute{Name: "stroke-width", Value: value}
}

// Points creates a points attribute for SVG polygon/polyline elements.
func Points(value string) Attribute {
	return StringAttribute{Name: "points", Value: value}
}

// D creates a d attribute for SVG path elements.
func D(value string) Attribute {
	return StringAttribute{Name: "d", Value: value}
}

// Cx creates a cx attribute for SVG circle elements.
func Cx(value string) Attribute {
	return StringAttribute{Name: "cx", Value: value}
}

// Cy creates a cy attribute for SVG circle elements.
func Cy(value string) Attribute {
	return StringAttribute{Name: "cy", Value: value}
}

// R creates an r attribute for SVG circle elements.
func R(value string) Attribute {
	return StringAttribute{Name: "r", Value: value}
}

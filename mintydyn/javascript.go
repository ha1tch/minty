package mintydyn

import (
	"fmt"
	"strings"
)

// =============================================================================
// JAVASCRIPT GENERATION
// =============================================================================

// sanitizeID converts an ID to a valid JavaScript identifier.
// Replaces hyphens and other invalid characters with underscores.
func sanitizeID(id string) string {
	// Replace hyphens with underscores (most common case)
	result := strings.ReplaceAll(id, "-", "_")
	// Replace any other non-alphanumeric characters
	var sb strings.Builder
	for i, r := range result {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9' && i > 0) || r == '_' {
			sb.WriteRune(r)
		} else if i == 0 && r >= '0' && r <= '9' {
			sb.WriteRune('_')
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	return sb.String()
}

// generateJavaScript creates all client-side code for the component.
func (db *DynamicBuilder[S, D, R]) generateJavaScript(pattern DetectedPattern) string {
	var js strings.Builder

	js.WriteString("<script>\n")

	// Generate base component class
	js.WriteString(db.generateBaseClass())

	// Generate pattern-specific managers
	if pattern.HasStates {
		js.WriteString(db.generateStatesManager())
	}
	if pattern.HasData {
		js.WriteString(db.generateDataManager())
	}
	if pattern.HasRules {
		js.WriteString(db.generateRulesManager())
	}

	// Generate coordination logic
	js.WriteString(db.generateCoordinationLogic(pattern))

	// Generate initialization
	js.WriteString(db.generateInitialization())

	js.WriteString("\n</script>")

	result := js.String()

	// Apply minification if enabled
	if db.options.MinifyJS {
		result = MinifyJS(result)
	}

	return result
}

// =============================================================================
// BASE COMPONENT CLASS
// =============================================================================

func (db *DynamicBuilder[S, D, R]) generateBaseClass() string {
	jsID := sanitizeID(db.id)
	return fmt.Sprintf(`
// Dynamic Component: %s
class DynamicComponent_%s {
    constructor() {
        this.id = '%s';
        this.container = document.getElementById(this.id);
        this.config = this.loadConfig();
        this.managers = {};
        this.externals = {};  // Registry for external objects (Google Maps, D3, etc.)
        this.hooks = this.config.hooks || {};
        this.state = {
            currentState: null,
            filters: {},
            dependencies: new Map(),
            initialized: false
        };
        
        this.initWithDependencies();
    }
    
    loadConfig() {
        const configScript = document.getElementById(this.id + '-config');
        return configScript ? JSON.parse(configScript.textContent) : {};
    }
    
    // Async initialization that waits for external scripts
    async initWithDependencies() {
        try {
            // Load required external scripts first
            await this.loadExternalScripts();
            
            // Initialize external object registry
            this.initExternalRegistry();
            
            // Run beforeInit hook
            if (this.hooks.beforeInit) {
                const result = await this.runHook('beforeInit', {});
                if (result === false) {
                    console.warn('DynamicComponent %s: beforeInit hook cancelled initialization');
                    return;
                }
            }
            
            // Core initialization
            this.init();
            
            // Run afterInit hook
            if (this.hooks.afterInit) {
                await this.runHook('afterInit', {});
            }
            
            this.state.initialized = true;
            this.trigger('component:ready');
            
        } catch (error) {
            console.error('DynamicComponent %s: initialization failed:', error);
            this.trigger('component:error', { error });
        }
    }
    
    // Load external scripts (Google Maps, D3, etc.)
    async loadExternalScripts() {
        const scripts = this.config.externalScripts || [];
        const required = scripts.filter(s => s.required);
        const optional = scripts.filter(s => !s.required);
        
        // Load required scripts first (blocks init)
        await Promise.all(required.map(script => this.loadScript(script)));
        
        // Load optional scripts in background
        optional.forEach(script => this.loadScript(script).catch(err => {
            console.warn('Optional script failed to load:', script.src, err);
        }));
    }
    
    loadScript(script) {
        return new Promise((resolve, reject) => {
            // Check if already loaded
            if (document.querySelector('script[src="' + script.src + '"]')) {
                if (script.onLoad) {
                    try { this.runHookCode(script.onLoad, {}); } catch(e) { console.warn(e); }
                }
                resolve();
                return;
            }
            
            const el = document.createElement('script');
            el.src = script.src;
            if (script.async) el.async = true;
            if (script.defer) el.defer = true;
            
            el.onload = () => {
                if (script.onLoad) {
                    try { this.runHookCode(script.onLoad, {}); } catch(e) { console.warn(e); }
                }
                resolve();
            };
            el.onerror = () => reject(new Error('Failed to load: ' + script.src));
            
            document.head.appendChild(el);
        });
    }
    
    // Initialize placeholder registry for external objects
    initExternalRegistry() {
        const registry = this.config.externalRegistry || [];
        registry.forEach(name => {
            this.externals[name] = null;  // Placeholder
        });
    }
    
    init() {
        if (!this.container) {
            console.error('DynamicComponent %s: container not found');
            return;
        }
        
        this.initializeManagers();
        this.setupCoordination();
        this.bindEvents();
    }
    
    initializeManagers() {
        const pattern = this.config.pattern;
        
        if (pattern.hasStates) {
            this.managers.states = new StatesManager_%s(this);
        }
        if (pattern.hasData) {
            this.managers.data = new DataManager_%s(this);
        }
        if (pattern.hasRules) {
            this.managers.rules = new RulesManager_%s(this);
        }
    }
    
    setupCoordination() {
        // Implemented by generateCoordinationLogic
    }
    
    bindEvents() {
        this.container.addEventListener('click', this.handleClick.bind(this));
        this.container.addEventListener('change', this.handleChange.bind(this));
        this.container.addEventListener('input', this.handleInput.bind(this));
    }
    
    handleClick(event) {
        // Use closest() to handle clicks on nested elements (e.g., SVG icons inside buttons)
        const actionElement = event.target.closest('[data-client-action]');
        if (actionElement) {
            const action = actionElement.dataset.clientAction;
            this.executeAction(action, actionElement, event);
        }
    }
    
    handleChange(event) {
        if (event.target.dataset.filterField) {
            this.trigger('filter:change', {
                field: event.target.dataset.filterField,
                value: this.getInputValue(event.target),
                element: event.target
            });
        }
        
        if (event.target.dataset.dependencyTrigger) {
            this.trigger('dependency:trigger', {
                triggerId: event.target.dataset.dependencyTrigger,
                value: this.getInputValue(event.target),
                element: event.target
            });
        }
    }
    
    handleInput(event) {
        clearTimeout(this.inputTimeout);
        this.inputTimeout = setTimeout(() => {
            this.handleChange(event);
        }, 300);
    }
    
    executeAction(action, element, event) {
        switch (action) {
            case 'switch-state':
                const stateId = element.dataset.stateTarget;
                this.switchToState(stateId);
                break;
            default:
                console.warn('Unknown action:', action);
        }
    }
    
    // State switching with hooks
    async switchToState(stateId) {
        const prevState = this.state.currentState;
        
        // beforeStateChange hook - can cancel
        if (this.hooks.beforeStateChange) {
            const result = await this.runHook('beforeStateChange', { from: prevState, to: stateId });
            if (result === false) {
                return false;
            }
        }
        
        // Per-state hook
        const stateHooks = this.hooks.stateHooks || {};
        if (stateHooks[stateId]) {
            await this.runHookCode(stateHooks[stateId], { from: prevState, to: stateId });
        }
        
        // Actual state switch
        if (this.managers.states) {
            this.managers.states.switchTo(stateId);
        }
        
        // afterStateChange hook
        if (this.hooks.afterStateChange) {
            await this.runHook('afterStateChange', { from: prevState, to: stateId });
        }
        
        return true;
    }
    
    getInputValue(element) {
        switch (element.type) {
            case 'checkbox':
            case 'radio':
                return element.checked;
            case 'number':
            case 'range':
                return Number(element.value);
            default:
                return element.value;
        }
    }
    
    // Hook execution
    runHook(hookName, context) {
        const hookCode = this.hooks[hookName];
        if (hookCode) {
            return this.runHookCode(hookCode, context);
        }
        return undefined;
    }
    
    runHookCode(code, context) {
        try {
            const fn = new Function('context', code);
            return fn.call(this, context);
        } catch (error) {
            console.error('Hook execution error:', error);
            return undefined;
        }
    }
    
    // External object management
    registerExternal(name, obj) {
        this.externals[name] = obj;
        this.trigger('external:registered', { name, obj });
        return obj;
    }
    
    getExternal(name) {
        return this.externals[name];
    }
    
    hasExternal(name) {
        return this.externals[name] != null;
    }
    
    // Event system
    trigger(eventName, data = {}) {
        const event = new CustomEvent('dyn:' + eventName, {
            detail: { ...data, component: this },
            bubbles: true
        });
        this.container.dispatchEvent(event);
    }
    
    on(eventName, callback) {
        this.container.addEventListener('dyn:' + eventName, callback);
        return () => this.container.removeEventListener('dyn:' + eventName, callback);
    }
    
    // Cleanup
    destroy() {
        // Run onDestroy hook
        if (this.hooks.onDestroy) {
            this.runHook('onDestroy', {});
        }
        
        // Cleanup externals
        Object.keys(this.externals).forEach(name => {
            const ext = this.externals[name];
            if (ext && typeof ext.destroy === 'function') {
                try { ext.destroy(); } catch(e) { console.warn('Cleanup error:', name, e); }
            }
        });
        
        this.externals = {};
        this.managers = {};
        this.state.initialized = false;
        
        // Remove from window
        delete window['DynComponent_%s'];
        
        this.trigger('component:destroyed');
    }
}

// Shared condition-operator evaluator, used by both StatesManager
// (evaluateCondition) and RulesManager (evaluateTriggerCondition) --
// they're separate classes, so this is a standalone function rather than
// a method, defined once here since generateBaseClass always runs
// regardless of which managers a given component ends up with.
// De-duplicates what were two near-identical switch statements (the
// original evaluateCondition's five operators were a verbatim subset of
// evaluateTriggerCondition's nine) into one shared implementation.
function evaluateConditionOperator_%s(value, operator, expected) {
    switch (operator) {
        case 'equals': return value == expected;
        case 'notEquals': return value != expected;
        case 'contains': return String(value).includes(String(expected));
        case 'greaterThan': return Number(value) > Number(expected);
        case 'lessThan': return Number(value) < Number(expected);
        case 'checked': return value === true;
        case 'unchecked': return value === false;
        case 'empty': return !value || value === '';
        case 'notEmpty': return value && value !== '';
        // List-membership operators added in 0.2.1: 'expected' must be
        // an array; each element is compared with '==' (loose equality,
        // matching every other operator above -- '5' matches 5, not
        // just '5' matches '5'). A malformed (non-array) 'expected' for
        // 'in' evaluates false; the same malformed 'expected' for
        // 'notIn' evaluates true (negating "not in the list" when there
        // is no valid list at all) -- both deliberately, not
        // accidentally: this matches the exact fail-safe behavior of
        // the evaluator this was designed to be a drop-in replacement
        // for, not a "corrected" version of it.
        case 'in':
            return Array.isArray(expected) && expected.some(item => item == value);
        case 'notIn':
            return !(Array.isArray(expected) && expected.some(item => item == value));
        default:
            console.warn('mintydyn: unrecognized condition operator', JSON.stringify(operator) + ' -- treating as false. Likely a typo (valid: equals, notEquals, contains, greaterThan, lessThan, checked, unchecked, empty, notEmpty, in, notIn).');
            return false;
    }
}

// Warns (does not throw -- these are all configuration mistakes that
// still resolve to a defined, fail-safe result; a hard error would be a
// worse experience for a mistake that's easy to make and easy to miss)
// about composite-condition-tree shapes confirmed, by direct testing, to
// be silent footguns otherwise:
//   - allOf and anyOf both set on the same node: allOf wins silently,
//     anyOf is ignored entirely.
//   - a composite (allOf/anyOf) node that also carries leaf fields
//     (field/component/componentId/operator/condition): the leaf fields
//     are ignored entirely, silently.
//   - an empty allOf array: JS's Array.prototype.every on an empty array
//     is vacuously true, so this evaluates as "condition met" -- the
//     opposite of what an empty condition list probably means to whoever
//     configured it.
//   - an empty anyOf array: vacuously false (Array.prototype.some), less
//     surprising than the allOf case but included for the same reason.
// Called from all three composite-tree entry points (StatesManager's
// evaluateCondition, RulesManager's evaluateTriggerCondition and
// evaluateTriggerConditionNode) rather than duplicated in each.
function warnIfMalformedConditionNode_%s(node, context) {
    if (node.allOf && node.anyOf) {
        console.warn('mintydyn (' + context + '): a condition node has both allOf and anyOf set -- allOf wins, anyOf is ignored entirely.', node);
    }
    const hasLeafFields = node.field || node.component || node.componentId || node.operator || node.condition;
    if ((node.allOf || node.anyOf) && hasLeafFields) {
        console.warn('mintydyn (' + context + '): a condition node has both a composite (allOf/anyOf) and leaf fields set -- the leaf fields are ignored entirely.', node);
    }
    if (node.allOf && node.allOf.length === 0) {
        console.warn('mintydyn (' + context + '): an allOf array is empty -- this evaluates as condition MET (vacuous AND), which is likely not what was intended.', node);
    }
    if (node.anyOf && node.anyOf.length === 0) {
        console.warn('mintydyn (' + context + '): an anyOf array is empty -- this evaluates as condition NOT met (vacuous OR).', node);
    }
}
`, db.id, jsID, db.id, db.id, db.id, db.id, jsID, jsID, jsID, jsID, jsID, jsID)
}

// =============================================================================
// STATES MANAGER
// =============================================================================

func (db *DynamicBuilder[S, D, R]) generateStatesManager() string {
	jsID := sanitizeID(db.id)
	return fmt.Sprintf(`
// States Manager
class StatesManager_%s {
    constructor(component) {
        this.component = component;
        this.states = component.config.states || [];
        this.themeClasses = component.config.themeClasses || {};
        this.currentState = null;
        this.stateElements = new Map();
        this.triggers = new Map();
        
        this.init();
    }
    
    init() {
        this.findStateElements();
        this.findStateTriggers();
        this.setInitialState();
    }
    
    findStateElements() {
        this.states.forEach(state => {
            const element = document.getElementById('state-' + state.id);
            if (element) {
                this.stateElements.set(state.id, element);
            }
        });
    }
    
    findStateTriggers() {
        const triggers = this.component.container.querySelectorAll('[data-state-target]');
        triggers.forEach(trigger => {
            const stateId = trigger.dataset.stateTarget;
            this.triggers.set(stateId, trigger);
        });
    }
    
    setInitialState() {
        const activeState = this.states.find(state => state.active);
        if (activeState) {
            this.switchTo(activeState.id, false);
        } else if (this.states.length > 0) {
            this.switchTo(this.states[0].id, false);
        }
    }
    
    switchTo(stateId, notify = true) {
        const state = this.states.find(s => s.id === stateId);
        if (!state) {
            console.warn('State not found:', stateId);
            return false;
        }
        
        if (state.disabled) {
            return false;
        }
        
        // Check condition if present
        if (state.condition && !this.evaluateCondition(state.condition)) {
            return false;
        }
        
        const prevState = this.currentState;
        
        // Hide current state
        if (this.currentState) {
            this.hideState(this.currentState);
        }
        
        // Show new state
        this.showState(stateId);
        this.currentState = stateId;
        this.component.state.currentState = stateId;
        
        if (notify) {
            this.component.trigger('state:change', {
                from: prevState,
                to: stateId,
                state: state
            });
        }
        
        return true;
    }
    
    showState(stateId) {
        const element = this.stateElements.get(stateId);
        const trigger = this.triggers.get(stateId);
        
        if (element) {
            // Remove hidden classes, add active classes
            this.removeClasses(element, this.themeClasses.contentHidden);
            this.addClasses(element, this.themeClasses.contentActive);
            element.setAttribute('aria-hidden', 'false');
        }
        
        if (trigger) {
            // Add active classes to trigger
            this.addClasses(trigger, this.themeClasses.triggerActive);
            trigger.setAttribute('aria-selected', 'true');
        }
    }
    
    hideState(stateId) {
        const element = this.stateElements.get(stateId);
        const trigger = this.triggers.get(stateId);
        
        if (element) {
            // Remove active classes, add hidden classes
            this.removeClasses(element, this.themeClasses.contentActive);
            this.addClasses(element, this.themeClasses.contentHidden);
            element.setAttribute('aria-hidden', 'true');
        }
        
        if (trigger) {
            // Remove active classes from trigger
            this.removeClasses(trigger, this.themeClasses.triggerActive);
            trigger.setAttribute('aria-selected', 'false');
        }
    }
    
    // Helper to add multiple classes (space-separated string)
    addClasses(element, classString) {
        if (!classString) return;
        classString.split(' ').filter(c => c).forEach(c => element.classList.add(c));
    }
    
    // Helper to remove multiple classes (space-separated string)
    removeClasses(element, classString) {
        if (!classString) return;
        classString.split(' ').filter(c => c).forEach(c => element.classList.remove(c));
    }
    
    // Handles both a plain leaf condition and a composite (allOf/anyOf)
    // condition tree, nesting arbitrarily. A leaf's value is always read
    // fresh from the DOM (unchanged from the original, non-composite
    // behaviour) since StateCondition never received a passed-in value to
    // begin with.
    evaluateCondition(condition) {
        warnIfMalformedConditionNode_%s(condition, 'StatesManager.evaluateCondition');
        if (condition.allOf) {
            return condition.allOf.every(c => this.evaluateCondition(c));
        }
        if (condition.anyOf) {
            return condition.anyOf.some(c => this.evaluateCondition(c));
        }

        const element = document.getElementById(condition.component || condition.field);
        if (!element) return false;

        const value = this.component.getInputValue(element);
        return evaluateConditionOperator_%s(value, condition.operator, condition.value);
    }
    
    getActiveState() {
        return this.currentState;
    }
    
    isStateActive(stateId) {
        return this.currentState === stateId;
    }
    
    getState(stateId) {
        return this.states.find(s => s.id === stateId);
    }
}
`, jsID, jsID, jsID)
}

// =============================================================================
// DATA MANAGER
// =============================================================================

func (db *DynamicBuilder[S, D, R]) generateDataManager() string {
	jsID := sanitizeID(db.id)
	return fmt.Sprintf(`
// Data Manager
class DataManager_%s {
    constructor(component) {
        this.component = component;
        this.schema = component.config.schema || { fields: [] };
        this.filterOptions = component.config.filterOptions || {};
        this.filters = new Map();
        this.currentPage = 1;
        this.itemsPerPage = this.filterOptions.itemsPerPage || 10;
        
        // Server-rendered mode uses pre-rendered DOM elements
        this.serverRendered = this.filterOptions.serverRendered || false;
        this.rowSelector = this.filterOptions.rowSelector || '.dyn-data-row';
        this.counterSelector = this.filterOptions.counterSelector || '';
        
        if (this.serverRendered) {
            this.rows = document.querySelectorAll(this.rowSelector);
            this.data = []; // Not used in server-rendered mode
            this.filteredData = [];
        } else {
            this.data = component.config.data || [];
            this.filteredData = [...this.data];
            this.rows = null;
        }
        
        this.init();
    }
    
    init() {
        this.setupFilters();
        if (this.serverRendered) {
            this.applyServerFilters();
        } else {
            this.renderResults();
        }
        this.bindFilterEvents();
    }
    
    setupFilters() {
        if (this.schema.fields && this.schema.fields.length > 0) {
            this.schema.fields.forEach(field => {
                this.filters.set(field.name, {
                    type: field.type,
                    value: field.defaultValue || this.getDefaultFilterValue(field.type),
                    active: false
                });
            });
        }
    }
    
    getDefaultFilterValue(type) {
        switch (type) {
            case 'text': return '';
            case 'boolean': return false;
            case 'range': return { min: null, max: null };
            case 'multiselect': return [];
            case 'select': return '';
            default: return null;
        }
    }
    
    bindFilterEvents() {
        this.component.on('filter:change', (event) => {
            this.updateFilter(event.detail.field, event.detail.value);
        });
    }
    
    updateFilter(field, value, notify = true) {
        if (this.filters.has(field)) {
            const filter = this.filters.get(field);
            filter.value = value;
            filter.active = this.isFilterValueActive(filter.type, value);
            
            if (this.serverRendered) {
                this.applyServerFilters();
            } else {
                this.applyFilters();
                this.currentPage = 1;
                this.renderResults();
            }
            
            if (notify) {
                const count = this.serverRendered ? this.visibleCount : this.filteredData.length;
                this.component.trigger('data:filtered', {
                    field: field,
                    value: value,
                    resultCount: count
                });
            }
        }
    }
    
    // Server-rendered filtering: show/hide existing DOM elements
    applyServerFilters() {
        if (!this.rows) return;
        
        let visibleCount = 0;
        this.rows.forEach(row => {
            const matches = this.rowMatchesFilters(row);
            row.style.display = matches ? '' : 'none';
            if (matches) visibleCount++;
        });
        
        this.visibleCount = visibleCount;
        this.updateCounter(visibleCount);
    }
    
    rowMatchesFilters(row) {
        for (const [field, filter] of this.filters.entries()) {
            if (!filter.active) continue;
            
            // Get value from data attribute (data-fieldname or data-field-name)
            const attrName = field.replace(/([A-Z])/g, '-$1').toLowerCase();
            const rowValue = row.dataset[field] || row.dataset[attrName] || '';
            
            if (!this.valueMatchesFilter(rowValue, filter)) {
                return false;
            }
        }
        return true;
    }
    
    valueMatchesFilter(rowValue, filter) {
        switch (filter.type) {
            case 'text':
                return String(rowValue).toLowerCase().includes(String(filter.value).toLowerCase());
            case 'boolean':
                return (rowValue === 'true' || rowValue === '1') === filter.value;
            case 'select':
                return filter.value === '' || rowValue === filter.value;
            case 'multiselect':
                return filter.value.length === 0 || filter.value.includes(rowValue);
            case 'range':
                const num = Number(rowValue);
                const min = filter.value.min != null ? Number(filter.value.min) : -Infinity;
                const max = filter.value.max != null ? Number(filter.value.max) : Infinity;
                return num >= min && num <= max;
            default:
                return rowValue === filter.value;
        }
    }
    
    updateCounter(count) {
        if (this.counterSelector) {
            const counter = document.querySelector(this.counterSelector);
            if (counter) {
                counter.textContent = 'Showing ' + count + ' items';
            }
        }
    }
    
    isFilterValueActive(type, value) {
        switch (type) {
            case 'text': return value && value.length > 0;
            case 'boolean': return value === true;
            case 'range': return value.min != null || value.max != null;
            case 'multiselect': return Array.isArray(value) && value.length > 0;
            case 'select': return value && value !== '';
            default: return value != null;
        }
    }
    
    // Client-side filtering for JSON data
    applyFilters() {
        this.filteredData = this.data.filter(item => {
            return Array.from(this.filters.entries()).every(([field, filter]) => {
                if (!filter.active) return true;
                return this.matchesFilter(item, field, filter);
            });
        });
    }
    
    matchesFilter(item, field, filter) {
        const itemValue = item[field];
        
        switch (filter.type) {
            case 'text':
                return String(itemValue || '').toLowerCase().includes(String(filter.value).toLowerCase());
            case 'boolean':
                return itemValue === filter.value;
            case 'range':
                const num = Number(itemValue);
                const min = filter.value.min != null ? Number(filter.value.min) : -Infinity;
                const max = filter.value.max != null ? Number(filter.value.max) : Infinity;
                return num >= min && num <= max;
            case 'multiselect':
                return filter.value.includes(itemValue);
            case 'select':
                return itemValue === filter.value;
            default:
                return itemValue === filter.value;
        }
    }
    
    renderResults() {
        const resultsContainer = document.getElementById(this.component.id + '-results');
        const summaryContainer = document.getElementById(this.component.id + '-summary');
        
        if (!resultsContainer) return;
        
        // Update summary
        if (summaryContainer) {
            summaryContainer.textContent = this.filteredData.length + ' results';
        }
        
        // Empty state
        if (this.filteredData.length === 0) {
            resultsContainer.innerHTML = '<div class="dyn-no-results text-gray-500 dark:text-gray-400 text-center py-8">No results found</div>';
            return;
        }
        
        // Paginate if needed
        let displayData = this.filteredData;
        if (this.filterOptions.enablePagination) {
            const start = (this.currentPage - 1) * this.itemsPerPage;
            const end = start + this.itemsPerPage;
            displayData = this.filteredData.slice(start, end);
        }
        
        // Render items - uses template from server or default
        resultsContainer.innerHTML = displayData.map(item => this.renderItem(item)).join('');
        
        // Update pagination
        if (this.filterOptions.enablePagination) {
            this.renderPagination();
        }
    }
    
    renderItem(item) {
        // Check if JSON view is requested via data-view-mode attribute
        const viewMode = this.component.container.dataset.viewMode;
        if (viewMode === 'json') {
            return '<div class="json-view">' + JSON.stringify(item, null, 2) + '</div>';
        }
        // Use template if provided, otherwise fall back to JSON
        if (this.filterOptions.itemTemplate) {
            let result = this.filterOptions.itemTemplate;
            Object.keys(item).forEach(field => {
                result = result.split('${' + field + '}').join(item[field] !== undefined ? item[field] : '');
            });
            return result;
        }
        // Default rendering - JSON dump
        return '<div class="dyn-result-item">' + JSON.stringify(item) + '</div>';
    }
    
    renderPagination() {
        const paginationContainer = document.getElementById(this.component.id + '-pagination');
        if (!paginationContainer) return;
        
        const totalPages = Math.ceil(this.filteredData.length / this.itemsPerPage);
        const themeClasses = this.component.config.themeClasses || {};
        const btnClass = themeClasses.paginationButton || 'dyn-page-btn';
        const activeClass = themeClasses.paginationButtonActive || 'active';
        let html = '';
        
        for (let i = 1; i <= totalPages; i++) {
            const classes = i === this.currentPage ? btnClass + ' ' + activeClass : btnClass;
            html += '<button class="' + classes + '" data-page="' + i + '">' + i + '</button>';
        }
        
        paginationContainer.innerHTML = html;
        
        // Bind page click events
        paginationContainer.querySelectorAll('button[data-page]').forEach(btn => {
            btn.addEventListener('click', () => {
                this.currentPage = parseInt(btn.dataset.page);
                this.renderResults();
            });
        });
    }
    
    getData() {
        if (this.serverRendered) {
            return Array.from(this.rows).filter(r => r.style.display !== 'none');
        }
        return this.filteredData;
    }
    
    getAllData() {
        if (this.serverRendered) {
            return Array.from(this.rows);
        }
        return this.data;
    }
    
    getVisibleCount() {
        if (this.serverRendered) {
            return this.visibleCount;
        }
        return this.filteredData.length;
    }
    
    clearFilters() {
        this.filters.forEach((filter, field) => {
            filter.active = false;
            filter.value = this.getDefaultFilterValue(filter.type);
        });
        if (this.serverRendered) {
            this.applyServerFilters();
        } else {
            this.applyFilters();
            this.currentPage = 1;
            this.renderResults();
        }
    }
    
    setData(newData) {
        if (this.serverRendered) {
            console.warn('setData not supported in server-rendered mode');
            return;
        }
        this.data = newData;
        this.applyFilters();
        this.currentPage = 1;
        this.renderResults();
    }
    
    refreshRows() {
        // Re-query rows (useful if DOM changed)
        if (this.serverRendered) {
            this.rows = document.querySelectorAll(this.rowSelector);
            this.applyServerFilters();
        }
    }
}
`, jsID)
}

// =============================================================================
// RULES MANAGER
// =============================================================================

func (db *DynamicBuilder[S, D, R]) generateRulesManager() string {
	jsID := sanitizeID(db.id)
	return fmt.Sprintf(`
// Rules Manager
class RulesManager_%s {
    constructor(component) {
        this.component = component;
        this.rules = component.config.rules || [];
        this.activeRules = new Map();
        this.ruleHistory = [];
        
        this.init();
    }
    
    init() {
        this.processRules();
        this.bindRuleEvents();
        this.evaluateInitialState();
    }
    
    processRules() {
        // Sort by priority
        this.rules.sort((a, b) => (b.priority || 0) - (a.priority || 0));
        
        // Index by every trigger component referenced anywhere in the
        // rule's condition tree -- a composite (allOf/anyOf) rule must
        // re-evaluate whenever any of its referenced components change,
        // not just a single top-level one. For a plain, non-composite
        // trigger this collects exactly the one componentId, matching the
        // original behaviour unchanged.
        this.rules.forEach(rule => {
            const triggerIds = [];
            this.collectTriggerComponentIds(rule.trigger, triggerIds);
            if (triggerIds.length === 0) {
                console.warn('mintydyn: rule', JSON.stringify(rule.id), 'references no componentId anywhere in its trigger condition -- it will never be registered against any trigger and can never fire.', rule);
            }
            triggerIds.forEach(triggerId => {
                if (!this.activeRules.has(triggerId)) {
                    this.activeRules.set(triggerId, []);
                }
                this.activeRules.get(triggerId).push(rule);
            });
        });
    }
    
    // Recursively walks a TriggerCondition tree, collecting every
    // componentId referenced at any leaf into out.
    collectTriggerComponentIds(node, out) {
        if (node.allOf) {
            node.allOf.forEach(c => this.collectTriggerComponentIds(c, out));
            return;
        }
        if (node.anyOf) {
            node.anyOf.forEach(c => this.collectTriggerComponentIds(c, out));
            return;
        }
        if (node.componentId) {
            out.push(node.componentId);
        }
    }
    
    bindRuleEvents() {
        this.component.on('dependency:trigger', (event) => {
            this.evaluateRules(event.detail.triggerId, event.detail.value);
        });
    }
    
    evaluateInitialState() {
        this.activeRules.forEach((rules, triggerId) => {
            // Find element(s) by data-dependency-trigger attribute
            const elements = this.component.container.querySelectorAll('[data-dependency-trigger="' + triggerId + '"]');
            if (elements.length === 0) return;
            
            // For radio buttons, find the checked one
            let element = elements[0];
            if (element.type === 'radio') {
                for (const el of elements) {
                    if (el.checked) {
                        element = el;
                        break;
                    }
                }
            }
            
            const value = this.component.getInputValue(element);
            this.evaluateRules(triggerId, value);
        });
    }
    
    evaluateRules(triggerId, value) {
        const rules = this.activeRules.get(triggerId) || [];
        
        rules.forEach(rule => {
            const conditionMet = this.evaluateTriggerCondition(rule.trigger, value);
            
            // For show/hide rules, toggle based on condition
            rule.actions.forEach(action => {
                if (action.action === 'show') {
                    if (conditionMet) {
                        this.executeAction(action);
                    } else {
                        // Hide when condition not met
                        this.executeAction({ ...action, action: 'hide' });
                    }
                } else if (action.action === 'hide') {
                    if (conditionMet) {
                        this.executeAction(action);
                    } else {
                        // Show when condition not met
                        this.executeAction({ ...action, action: 'show' });
                    }
                } else if (conditionMet) {
                    // Other actions only execute when condition is met
                    this.executeAction(action);
                }
            });
            
            if (conditionMet) {
                this.ruleHistory.push({
                    ruleId: rule.id,
                    triggerId: triggerId,
                    value: value,
                    timestamp: Date.now()
                });
                
                this.component.trigger('rule:executed', {
                    rule: rule,
                    triggerId: triggerId,
                    value: value
                });
            }
        });
    }
    
    // The top-level leaf case is unchanged from the original: it uses the
    // value passed in from the event that fired, never reading the DOM
    // itself. A composite (allOf/anyOf) delegates to
    // evaluateTriggerConditionNode for every sub-condition instead, since
    // those aren't necessarily the element that fired -- their values are
    // read fresh from the DOM by componentId.
    evaluateTriggerCondition(trigger, value) {
        warnIfMalformedConditionNode_%s(trigger, 'RulesManager.evaluateTriggerCondition');
        if (trigger.allOf) {
            return trigger.allOf.every(c => this.evaluateTriggerConditionNode(c));
        }
        if (trigger.anyOf) {
            return trigger.anyOf.some(c => this.evaluateTriggerConditionNode(c));
        }
        return evaluateConditionOperator_%s(value, trigger.condition, trigger.value);
    }
    
    // Evaluates a single node within a composite TriggerCondition tree,
    // recursing through further nesting and reading each leaf's current
    // value fresh from the DOM by componentId -- unlike the top-level
    // leaf case, a composite sub-condition has no passed-in value to use,
    // since it isn't necessarily the component that fired the event being
    // evaluated.
    evaluateTriggerConditionNode(node) {
        warnIfMalformedConditionNode_%s(node, 'RulesManager.evaluateTriggerConditionNode');
        if (node.allOf) {
            return node.allOf.every(c => this.evaluateTriggerConditionNode(c));
        }
        if (node.anyOf) {
            return node.anyOf.some(c => this.evaluateTriggerConditionNode(c));
        }
        const element = document.getElementById(node.componentId);
        if (!element) return false;
        const value = this.component.getInputValue(element);
        return evaluateConditionOperator_%s(value, node.condition, node.value);
    }
    
    executeRuleActions(rule) {
        rule.actions.forEach(action => {
            this.executeAction(action);
        });
    }
    
    executeAction(action) {
        const target = document.getElementById(action.targetId);
        if (!target) {
            console.warn('Rule target not found:', action.targetId);
            return;
        }
        
        switch (action.action) {
            case 'show':
                target.classList.remove('hidden', 'd-none');
                target.style.display = '';
                target.setAttribute('aria-hidden', 'false');
                break;
            case 'hide':
                target.classList.add('hidden');
                target.style.display = 'none';
                target.setAttribute('aria-hidden', 'true');
                break;
            case 'enable':
                target.disabled = false;
                target.classList.remove('disabled');
                break;
            case 'disable':
                target.disabled = true;
                target.classList.add('disabled');
                break;
            case 'addClass':
                if (action.value) target.classList.add(String(action.value));
                break;
            case 'removeClass':
                if (action.value) target.classList.remove(String(action.value));
                break;
            case 'setValue':
                if (target.type === 'checkbox' || target.type === 'radio') {
                    target.checked = Boolean(action.value);
                } else {
                    target.value = String(action.value || '');
                }
                break;
            case 'setText':
                target.textContent = String(action.value || '');
                break;
            case 'setHTML':
                target.innerHTML = String(action.value || '');
                break;
            case 'focus':
                target.focus();
                break;
            case 'blur':
                target.blur();
                break;
            default:
                console.warn('Unknown rule action:', action.action);
        }
    }
    
    getRuleHistory() {
        return this.ruleHistory;
    }
    
    clearRuleHistory() {
        this.ruleHistory = [];
    }
}
`, jsID, jsID, jsID, jsID, jsID)
}

// =============================================================================
// COORDINATION LOGIC
// =============================================================================

func (db *DynamicBuilder[S, D, R]) generateCoordinationLogic(pattern DetectedPattern) string {
	jsID := sanitizeID(db.id)
	var js strings.Builder

	js.WriteString(fmt.Sprintf(`
// Coordination Logic
DynamicComponent_%s.prototype.setupCoordination = function() {
    const pattern = this.config.pattern;
`, jsID))

	switch pattern.PrimaryPattern {
	case PatternStatefulData, PatternFilterableStates:
		js.WriteString(`
    // States + Data coordination
    if (this.managers.states && this.managers.data) {
        this.on('state:change', (event) => {
            const stateContext = event.detail.to;
            this.filterDataForState(stateContext);
        });
    }
`)

	case PatternDependentStates:
		js.WriteString(`
    // States + Rules coordination
    if (this.managers.states && this.managers.rules) {
        this.on('rule:executed', (event) => {
            const rule = event.detail.rule;
            this.handleStateAffectingRule(rule);
        });
    }
`)

	case PatternDependentData:
		js.WriteString(`
    // Data + Rules coordination
    if (this.managers.data && this.managers.rules) {
        this.on('rule:executed', (event) => {
            const rule = event.detail.rule;
            this.handleFilterAffectingRule(rule);
        });
    }
`)

	case PatternComplete:
		js.WriteString(`
    // Complete coordination
    this.setupCompleteCoordination();
`)
	}

	js.WriteString(`
};
`)

	// Add helper methods based on pattern
	if pattern.HasStates && pattern.HasData {
		js.WriteString(fmt.Sprintf(`
DynamicComponent_%s.prototype.filterDataForState = function(stateContext) {
    const stateFilters = this.getFiltersForState(stateContext);
    Object.entries(stateFilters).forEach(([field, value]) => {
        this.managers.data.updateFilter(field, value, false);
    });
    this.managers.data.renderResults();
};

DynamicComponent_%s.prototype.getFiltersForState = function(stateId) {
    // Override in specific implementations or define via hooks
    return {};
};
`, jsID, jsID))
	}

	if pattern.PrimaryPattern == PatternComplete {
		js.WriteString(fmt.Sprintf(`
DynamicComponent_%s.prototype.setupCompleteCoordination = function() {
    // State changes affect data
    this.on('state:change', (event) => {
        this.filterDataForState(event.detail.to);
    });
    
    // Rules affect both states and data
    this.on('rule:executed', (event) => {
        const rule = event.detail.rule;
        rule.actions.forEach(action => {
            if (action.targetId.startsWith('state-')) {
                this.handleStateAffectingRule(rule);
            } else {
                this.handleFilterAffectingRule(rule);
            }
        });
    });
    
    // Empty results trigger
    this.on('data:filtered', (event) => {
        if (event.detail.resultCount === 0) {
            this.handleEmptyResults();
        }
    });
};

DynamicComponent_%s.prototype.handleStateAffectingRule = function(rule) {
    // Rules can affect state visibility
};

DynamicComponent_%s.prototype.handleFilterAffectingRule = function(rule) {
    // Rules can affect filter availability
};

DynamicComponent_%s.prototype.handleEmptyResults = function() {
    // Show no-results state if available
    if (this.managers.states) {
        const noResultsState = this.managers.states.states.find(s => s.id === 'no-results');
        if (noResultsState) {
            this.managers.states.switchTo('no-results');
        }
    }
};
`, jsID, jsID, jsID, jsID))
	}

	return js.String()
}

// =============================================================================
// INITIALIZATION
// =============================================================================

func (db *DynamicBuilder[S, D, R]) generateInitialization() string {
	jsID := sanitizeID(db.id)
	return fmt.Sprintf(`
// Auto-initialization
document.addEventListener('DOMContentLoaded', function() {
    if (document.getElementById('%s')) {
        window.DynComponent_%s = new DynamicComponent_%s();
    }
});
`, db.id, jsID, jsID)
}

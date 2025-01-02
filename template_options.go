package templatex

import "html/template"

// Option is a function type that takes a pointer to an Engine as its argument.
// It represents a functional option pattern for configuring the Engine instance.
// Options are applied using variadic functions to modify Engine properties
// in a type-safe and flexible way.
type Option func(*Engine)

// WithFuncs sets custom template functions that will be available in all templates.
// It accepts a template.FuncMap containing the mapping of function names to their
// implementations. If the provided FuncMap is not empty, these functions will be
// added to the Engine's function map, making them accessible within templates.
// Existing functions with the same names will be overwritten.
func WithFuncs(fns template.FuncMap) Option {
	return func(e *Engine) {
		if len(fns) > 0 {
			for name, fn := range fns {
				e.funcMap[name] = fn
			}
		}
	}
}

// WithFunc sets a single template function that will be available in all templates.
// It accepts a name string for the function and the implementation function itself.
// The provided function will be added to the Engine's function map, making it
// accessible within templates. An existing function with the same name will be
// overwritten.
func WithFunc(name string, fn interface{}) Option {
	return func(e *Engine) {
		e.funcMap[name] = fn
	}
}

// WithExtensions sets the file extensions that will be used for template files.
// It accepts a variadic number of string arguments representing file extensions
// (e.g., ".tmpl", ".html") and replaces the default ".gohtml" extension.
// If no extensions are provided, the current extension settings remain unchanged.
// Multiple extensions can be specified to support different template file types.
func WithExtensions(exts ...string) Option {
	return func(e *Engine) {
		if len(exts) > 0 {
			e.exts = exts
		}
	}
}

// WithLayouts sets the layout templates that will be used as base templates for all pages.
// It accepts a variadic number of string arguments representing layout template file paths
// (e.g., "layouts/base.gohtml", "layouts/main.gohtml"). These layouts are used as common
// templates that wrap content templates. Setting layouts explicitly can optimize template
// processing by pre-defining the layout chain computation. If no layouts are provided,
// the current layout settings remain unchanged.
func WithLayouts(layouts ...string) Option {
	return func(e *Engine) {
		if len(layouts) > 0 {
			e.commonLayouts = layouts
		}
	}
}

// WithHardCache sets the hard caching behavior of the template engine.
// When hard caching is enabled, rendered templates are cached permanently and only
// re-rendered if the cache is manually cleared. This can significantly improve
// performance for templates with static content, but should be used with caution
// for dynamic content. When disabled (default), cache key includes template content,
// layouts and data hash, so templates are only re-rendered when data changes.
func WithHardCache(enabled bool) Option {
	return func(e *Engine) {
		e.hardCache = enabled
	}
}

package templatex

import "html/template"

// Option represents a functional option for the Engine
type Option func(*Engine)

// WithFuncs sets the template functions
func WithFuncs(fns template.FuncMap) Option {
	return func(e *Engine) {
		if len(fns) > 0 {
			for name, fn := range fns {
				e.funcMap[name] = fn
			}
		}
	}
}

// WithExtensions sets the template extensions.
// It replaces the default ".gohtml" extension.
// Example: ".tmpl", ".html"
func WithExtensions(exts ...string) Option {
	return func(e *Engine) {
		if len(exts) > 0 {
			e.exts = exts
		}
	}
}

// WithLayouts sets the layout templates
// Example: "layouts/base.gohtml", "layouts/main.gohtml"
// Usefull for optimizing the layout chain computation.
func WithLayouts(layouts ...string) Option {
	return func(e *Engine) {
		if len(layouts) > 0 {
			e.commonLayouts = layouts
		}
	}
}

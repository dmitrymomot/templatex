package templatex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// layoutChain represents a pre-computed chain of templates
type layoutChain struct {
	templates []*template.Template
}

// Engine holds parsed templates and manages their rendering
type Engine struct {
	templates     *template.Template
	layouts       map[string]*template.Template
	mu            sync.RWMutex
	cache         sync.Map // template cache
	layoutCache   sync.Map // layout chain cache
	funcMap       template.FuncMap
	exts          []string
	commonLayouts []string
}

// New initializes a Template Engine with optimized caching and pre-compiled layouts
func New(root string, opts ...Option) (*Engine, error) {
	if root == "" {
		return nil, ErrNoTemplateDirectory
	}

	// Check if directory exists
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, errors.Join(ErrNoTemplateDirectory, fmt.Errorf("template directory does not exist: %s", root))
	}

	// Initialize engine
	e := &Engine{
		layouts: make(map[string]*template.Template),
		funcMap: defaultFuncs(),
		exts:    []string{".gohtml"},
	}

	// Apply options
	for _, opt := range opts {
		opt(e)
	}

	// Parse templates
	tmpl := template.New("").Option("missingkey=zero").Funcs(e.funcMap)
	if err := filepath.Walk(root, e.walkFunc(tmpl, root, e.exts)); err != nil {
		return nil, errors.Join(ErrTemplateParsingFailed, err)
	}

	if tmpl.Templates() == nil {
		return nil, ErrNoTemplatesParsed
	}

	e.templates = tmpl

	// Pre-compile common layouts
	e.precompileCommonLayouts()

	return e, nil
}

// walkFunc is now a method of Engine to access its internal state
func (e *Engine) walkFunc(tmpl *template.Template, root string, exts []string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Check file extension
		validExt := false
		for _, ext := range exts {
			if filepath.Ext(path) == ext {
				validExt = true
				break
			}
		}
		if !validExt {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		tmplName := strings.TrimSuffix(relPath, filepath.Ext(relPath))

		if bytes.Contains(content, []byte("{{define")) {
			_, err = tmpl.ParseFiles(path)
		} else {
			_, err = tmpl.New(tmplName).Parse(string(content))
		}

		return err
	}
}

// precompileCommonLayouts pre-compiles frequently used layouts
func (e *Engine) precompileCommonLayouts() {
	for _, layout := range e.commonLayouts {
		if t := e.templates.Lookup(layout); t != nil {
			e.layouts[layout] = t
		}
	}
}

// getLayoutChain returns a cached layout chain or creates a new one
func (e *Engine) getLayoutChain(layouts ...string) (*layoutChain, error) {
	if len(layouts) == 0 {
		return &layoutChain{}, nil
	}

	cacheKey := strings.Join(layouts, ":")
	if cached, ok := e.layoutCache.Load(cacheKey); ok {
		return cached.(*layoutChain), nil
	}

	chain := &layoutChain{
		templates: make([]*template.Template, len(layouts)),
	}

	for i, layout := range layouts {
		if t := e.templates.Lookup(layout); t != nil {
			chain.templates[i] = t
		} else {
			return nil, fmt.Errorf("layout not found: %s", layout)
		}
	}

	e.layoutCache.Store(cacheKey, chain)
	return chain, nil
}

// Render implements optimized template rendering
// Render implements optimized template rendering
func (e *Engine) Render(ctx context.Context, out io.Writer, name string, binding interface{}, layouts ...string) error {
	if e == nil || e.templates == nil {
		return ErrTemplateEngineNotInitialized
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("%s-%v", name, layouts)
	if cached, ok := e.cache.Load(cacheKey); ok {
		if cachedContent, ok := cached.(string); ok {
			_, err := io.WriteString(out, cachedContent)
			return err
		}
	}

	// Get buffer from pool
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Get the base template
	e.mu.RLock()
	baseTmpl := e.templates.Lookup(name)
	e.mu.RUnlock()

	if baseTmpl == nil {
		return errors.Join(ErrTemplateNotFound, fmt.Errorf("template: %s", name))
	}

	// Create a new template with context-specific functions
	localFuncs := template.FuncMap{
		"T":      getTranslator(ctx),
		"ctxVal": ctxValue(ctx),
	}

	// Execute the base template
	if err := executeTemplateWithFuncs(baseTmpl, buf, binding, localFuncs); err != nil {
		return errors.Join(ErrTemplateExecutionFailed, err)
	}

	// Get layout chain
	chain, err := e.getLayoutChain(layouts...)
	if err != nil {
		return err
	}

	// Process layout chain
	content := buf.String()
	for _, layoutTmpl := range chain.templates {
		buf.Reset()

		embedFunc := template.FuncMap{
			"embed": func() template.HTML {
				return template.HTML(content)
			},
		}

		if err := executeTemplateWithFuncs(layoutTmpl, buf, binding, embedFunc); err != nil {
			return errors.Join(ErrTemplateExecutionFailed, err)
		}

		content = buf.String()
	}

	// Store the final rendered content in cache
	e.cache.Store(cacheKey, content)

	// Write final output
	_, err = io.WriteString(out, content)
	return err
}

// executeTemplateWithFuncs safely executes a template with additional functions
func executeTemplateWithFuncs(tmpl *template.Template, buf *bytes.Buffer, data interface{}, fns template.FuncMap) error {
	// Create a new template
	newTmpl, err := tmpl.Clone()
	if err != nil {
		return err
	}

	// Add the functions
	newTmpl = newTmpl.Funcs(fns)

	// Execute the template
	return newTmpl.Execute(buf, data)
}

// RenderString and RenderHTML implementations remain similar but use the optimized Render method
func (e *Engine) RenderString(ctx context.Context, name string, binding interface{}, layouts ...string) (string, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	if err := e.Render(ctx, buf, name, binding, layouts...); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (e *Engine) RenderHTML(ctx context.Context, name string, binding interface{}, layouts ...string) (template.HTML, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	if err := e.Render(ctx, buf, name, binding, layouts...); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

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
	templates   *template.Template
	layouts     map[string]*template.Template
	mu          sync.RWMutex
	cache       sync.Map // template cache
	layoutCache sync.Map // layout chain cache
	funcMap     template.FuncMap
}

// New initializes a Template Engine with optimized caching and pre-compiled layouts
func New(root string, fns template.FuncMap, exts ...string) (*Engine, error) {
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
		funcMap: make(template.FuncMap),
	}

	// Combine function maps
	baseFuncs := defaultFuncs()
	for name, fn := range baseFuncs {
		e.funcMap[name] = fn
	}
	if len(fns) > 0 {
		for name, fn := range fns {
			e.funcMap[name] = fn
		}
	}

	// Parse templates
	tmpl := template.New("").Option("missingkey=zero").Funcs(e.funcMap)

	if len(exts) == 0 {
		exts = []string{".gohtml"}
	}

	if err := filepath.Walk(root, e.walkFunc(tmpl, root, exts)); err != nil {
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
	commonLayouts := []string{"base_layout.html", "app_layout.html"}
	for _, layout := range commonLayouts {
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
func (e *Engine) Render(ctx context.Context, out io.Writer, name string, binding interface{}, layouts ...string) error {
	if e == nil || e.templates == nil {
		return ErrTemplateEngineNotInitialized
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("%s-%v", name, layouts)
	if cached, ok := e.cache.Load(cacheKey); ok {
		if tmpl, ok := cached.(*template.Template); ok {
			return tmpl.Execute(out, binding)
		}
	}

	// Get buffer from pool
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Execute base template
	e.mu.RLock()
	tmpl := e.templates.Lookup(name)
	e.mu.RUnlock()

	if tmpl == nil {
		return errors.Join(ErrTemplateNotFound, fmt.Errorf("template: %s", name))
	}

	// Add context-specific functions
	tmpl = tmpl.Funcs(template.FuncMap{
		"T":      getTranslator(ctx),
		"ctxVal": ctxValue(ctx),
	})

	if err := tmpl.Execute(buf, binding); err != nil {
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

		layoutTmpl = layoutTmpl.Funcs(template.FuncMap{
			"embed": func() template.HTML {
				return template.HTML(content)
			},
		})

		if err := layoutTmpl.Execute(buf, binding); err != nil {
			return errors.Join(ErrTemplateExecutionFailed, err)
		}

		content = buf.String()
	}

	// Store in cache
	e.cache.Store(cacheKey, tmpl)

	// Write final output
	_, err = io.WriteString(out, content)
	return err
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

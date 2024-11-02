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

type Renderer interface {
	Render(ctx context.Context, out io.Writer, name string, binding interface{}, layouts ...string) error
}

type RendererHTML interface {
	Renderer
	RenderHTML(ctx context.Context, name string, binding interface{}, layouts ...string) (template.HTML, error)
}

type RendererString interface {
	Renderer
	RenderString(ctx context.Context, name string, binding interface{}, layouts ...string) (string, error)
}

// Engine holds parsed templates and manages their rendering.
type Engine struct {
	templates *template.Template
	mu        sync.RWMutex
}

// New initializes a TemplateManager by parsing templates from a glob pattern and adding custom function maps.
func New(root string, fns template.FuncMap) (*Engine, error) {
	if root == "" {
		return nil, ErrNoTemplateDirectory
	}

	// Parse the templates and add a custom function map
	tmpl := template.New("").
		Option("missingkey=zero").
		Funcs(defaultFuncs()).
		Funcs(template.FuncMap{
			"embed": func() template.HTML { return "" },                  // placeholder function
			"T":     func(key string, args ...any) string { return key }, // placeholder function with variadic args
		})

	// Add a custom function map
	if len(fns) > 0 {
		tmpl = tmpl.Funcs(fns)
	}

	// Parse the templates
	err := filepath.Walk(root, walkFunc(tmpl, root))
	if err != nil {
		return nil, errors.Join(ErrTemplateParsingFailed, err)
	}

	return &Engine{templates: tmpl}, nil
}

// walkFunc is a helper function that parses a template file and adds it to the template manager.
// It is used with filepath.Walk to parse all template files in a directory.
func walkFunc(tmpl *template.Template, root string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".html" {
			// Get the relative path and normalize it
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

			// Parse the file and name it with the relative path
			if _, err = tmpl.New(relPath).ParseFiles(path); err != nil {
				return err
			}
		}
		return nil
	}
}

// Render renders a page using the specified layout chain.
func (tm *Engine) Render(ctx context.Context, out io.Writer, name string, binding interface{}, layouts ...string) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var err error
	var buf bytes.Buffer

	// Get the locale from the context
	tm.templates.Funcs(template.FuncMap{"T": getTranslator(ctx)})

	// Render the base content (e.g., contacts.html) into a buffer.
	err = tm.templates.ExecuteTemplate(&buf, name, binding)
	if err != nil {
		return errors.Join(ErrTemplateExecutionFailed, err)
	}

	// Prepare the embed function to provide the previous buffer's content.
	embedContent := buf.String()

	// Iterate through each layout, wrapping the previous output.
	for _, layout := range layouts {
		buf.Reset() // Clear buffer for the next layer.

		// Create a template with the embed function updated for each layer.
		tmpl := tm.templates.Lookup(layout)
		if tmpl == nil {
			return errors.Join(ErrTemplateNotFound, fmt.Errorf("layout: %s", layout))
		}

		// Add a custom embed function that returns the current content.
		tmpl = tmpl.Funcs(template.FuncMap{
			"embed": func() template.HTML {
				return template.HTML(embedContent)
			},
		})

		// Render the current layout with the updated embed content.
		if err = tmpl.Execute(&buf, binding); err != nil {
			return errors.Join(ErrTemplateExecutionFailed, err)
		}

		// Update embedContent with new embedded content for the next layer.
		embedContent = buf.String()
	}

	// Write the final output to the provided writer.
	if _, err = io.WriteString(out, embedContent); err != nil {
		return errors.Join(ErrTemplateExecutionFailed, err)
	}
	return nil
}

// RenderString renders a page using the specified layout chain and returns the result as a string.
func (tm *Engine) RenderString(ctx context.Context, name string, binding interface{}, layouts ...string) (string, error) {
	var buf bytes.Buffer
	if err := tm.Render(ctx, &buf, name, binding, layouts...); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderHTML renders a page using the specified layout chain and returns the result as a template.HTML.
// This function is useful when embedding the result in a template.
func (tm *Engine) RenderHTML(ctx context.Context, name string, binding interface{}, layouts ...string) (template.HTML, error) {
	var buf bytes.Buffer
	if err := tm.Render(ctx, &buf, name, binding, layouts...); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

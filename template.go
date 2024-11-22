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
	mu        sync.Mutex
}

// New initializes a TemplateManager by parsing templates from a glob pattern and adding custom function maps.
func New(root string, fns template.FuncMap, exts ...string) (*Engine, error) {
	if root == "" {
		return nil, ErrNoTemplateDirectory
	}

	// Check if directory exists
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, errors.Join(ErrNoTemplateDirectory, fmt.Errorf("template directory does not exist: %s", root))
	}

	// Parse the templates and add a custom function map
	tmpl := template.New("").Option("missingkey=zero").Funcs(defaultFuncs())

	// Add a custom function map
	if len(fns) > 0 {
		tmpl = tmpl.Funcs(fns)
	}

	// Parse the templates
	if len(exts) == 0 {
		exts = []string{".html"}
	}
	if err := filepath.Walk(root, walkFunc(tmpl, root, exts)); err != nil {
		return nil, errors.Join(ErrTemplateParsingFailed, err)
	}

	// Verify that at least one template was parsed
	if tmpl.Templates() == nil {
		return nil, ErrNoTemplatesParsed
	}

	return &Engine{templates: tmpl}, nil
}

// walkFunc is a helper function that parses a template file and adds it to the template manager.
// It is used with filepath.Walk to parse all template files in a directory.
func walkFunc(tmpl *template.Template, root string, exts []string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			isValidExt := false
			for _, ext := range exts {
				if filepath.Ext(path) == ext {
					isValidExt = true
					break
				}
			}
			if !isValidExt {
				return nil
			}

			// Get the relative path and normalize it
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

			// Read the content of the file
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Check if the file contains any {{define}} blocks
			if bytes.Contains(content, []byte("{{define")) {
				// Parse the file directly if it contains define blocks
				_, err = tmpl.ParseFiles(path)
				if err != nil {
					return err
				}
			} else {
				// If no define blocks, create a new template with the file name
				// and parse the content
				_, err = tmpl.New(relPath).Parse(string(content))
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// Render renders a page using the specified layout chain.
func (tm *Engine) Render(ctx context.Context, out io.Writer, name string, binding interface{}, layouts ...string) error {
	if tm == nil || tm.templates == nil {
		return ErrTemplateEngineNotInitialized
	}

	tm.mu.Lock() // Changed from RLock to Lock since we're modifying the template
	// Create a clone of the template to avoid concurrent modifications
	tmpl, err := tm.templates.Clone()
	tm.mu.Unlock()
	if err != nil {
		return errors.Join(ErrTemplateExecutionFailed, err)
	}

	// Add context-specific functions
	tmpl = tmpl.Funcs(template.FuncMap{
		"T":      getTranslator(ctx),
		"ctxVal": ctxValue(ctx),
	})

	var buf bytes.Buffer

	// Render the base content (e.g., contacts.html) into a buffer.
	err = tmpl.ExecuteTemplate(&buf, name, binding)
	if err != nil {
		return errors.Join(ErrTemplateExecutionFailed, err)
	}

	// Prepare the embed function to provide the previous buffer's content.
	embedContent := buf.String()

	// Iterate through each layout, wrapping the previous output.
	for _, layout := range layouts {
		buf.Reset() // Clear buffer for the next layer.

		// Create a template with the embed function updated for each layer.
		tmpl := tmpl.Lookup(layout)
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

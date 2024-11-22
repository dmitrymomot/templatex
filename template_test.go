package templatex_test

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmitrymomot/templatex"
	"github.com/invopop/ctxi18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed example/*.yml
var testTranslations embed.FS

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		fns      template.FuncMap
		exts     []string
		wantErr  bool
		errorMsg string
	}{
		{
			name:     "Empty root directory",
			root:     "",
			wantErr:  true,
			errorMsg: "no template directory provided",
		},
		{
			name:     "Non-existent directory",
			root:     "nonexistent",
			wantErr:  true,
			errorMsg: "template directory does not exist",
		},
		{
			name: "Valid directory with custom functions",
			root: "example/templates/",
			fns: template.FuncMap{
				"upper": strings.ToUpper,
			},
			wantErr: false,
		},
		{
			name:    "Valid directory with custom extensions",
			root:    "example/templates/",
			exts:    []string{".html", ".tpl"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := templatex.New(tt.root, tt.fns, tt.exts...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, engine)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, engine)
			}
		})
	}
}

func TestRender(t *testing.T) {
	// Setup test environment
	engine, err := templatex.New("example/templates/", nil)
	require.NoError(t, err)
	require.NotNil(t, engine)

	// Load translations
	err = ctxi18n.LoadWithDefault(testTranslations, "en")
	require.NoError(t, err)

	ctx, err := ctxi18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)

	tests := []struct {
		name     string
		template string
		data     interface{}
		layouts  []string
		want     string
		wantErr  bool
	}{
		{
			name:     "Simple template",
			template: "greeter.html",
			data: pageData{
				Title:    "Test",
				Username: "John",
				Test:     "Message",
			},
			layouts: []string{"base_layout.html"},
			wantErr: false,
		},
		{
			name:     "Multiple layouts",
			template: "greeter.html",
			data: pageData{
				Title:    "Test",
				Username: "John",
				Test:     "Message",
			},
			layouts: []string{"app_layout.html", "base_layout.html"},
			wantErr: false,
		},
		{
			name:     "Non-existent template",
			template: "nonexistent.html",
			data:     nil,
			wantErr:  true,
		},
		{
			name:     "Non-existent layout",
			template: "greeter.html",
			data:     nil,
			layouts:  []string{"nonexistent.html"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := engine.Render(ctx, &buf, tt.template, tt.data, tt.layouts...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, buf.String())
			}
		})
	}
}

func TestRenderString(t *testing.T) {
	engine, err := templatex.New("example/templates/", nil)
	require.NoError(t, err)

	// Load translations
	err = ctxi18n.LoadWithDefault(testTranslations, "en")
	require.NoError(t, err)

	ctx, err := ctxi18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)

	result, err := engine.RenderString(ctx, "greeter.html", pageData{
		Title:    "Test",
		Username: "John",
		Test:     "Message",
	}, "base_layout.html")

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestRenderHTML(t *testing.T) {
	engine, err := templatex.New("example/templates/", nil)
	require.NoError(t, err)

	// Load translations
	err = ctxi18n.LoadWithDefault(testTranslations, "en")
	require.NoError(t, err)

	ctx, err := ctxi18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)

	result, err := engine.RenderHTML(ctx, "greeter.html", pageData{
		Title:    "Test",
		Username: "John",
		Test:     "Message",
	}, "base_layout.html")

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestTemplateWithCustomFunctions(t *testing.T) {
	customFuncs := template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}

	_, err := templatex.New("example/templates/", customFuncs)
	require.NoError(t, err)

	// Create a test template file with custom functions
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.html")
	content := `{{.Text | upper}}`
	err = os.WriteFile(tempFile, []byte(content), 0644)
	require.NoError(t, err)

	// Create a new engine with the temp directory
	engine, err := templatex.New(tempDir, customFuncs)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = engine.Render(context.Background(), &buf, "test.html", struct{ Text string }{"hello"})
	assert.NoError(t, err)
	assert.Equal(t, "HELLO", buf.String())
}

func TestTemplateWithDifferentExtensions(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files with different extensions
	files := map[string]string{
		"test1.html": "HTML template",
		"test2.tpl":  "TPL template",
		"test3.txt":  "TXT template",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test with specific extensions
	engine, err := templatex.New(tempDir, nil, ".html", ".tpl")
	require.NoError(t, err)

	// Try rendering each template
	for name, expectedContent := range files {
		ext := filepath.Ext(name)
		if ext == ".html" || ext == ".tpl" {
			var buf bytes.Buffer
			err := engine.Render(context.Background(), &buf, name, nil)
			assert.NoError(t, err)
			assert.Equal(t, expectedContent, buf.String())
		}
	}
}

func TestConcurrentRendering(t *testing.T) {
	engine, err := templatex.New("example/templates/", nil)
	require.NoError(t, err)

	// Load translations
	err = ctxi18n.LoadWithDefault(testTranslations, "en")
	require.NoError(t, err)

	ctx, err := ctxi18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)

	data := pageData{
		Title:    "Test",
		Username: "John",
		Test:     "Message",
	}

	// Run multiple goroutines to test concurrent rendering
	concurrency := 10
	done := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func() {
			var buf bytes.Buffer
			err := engine.Render(ctx, &buf, "greeter.html", data, "base_layout.html")
			assert.NoError(t, err)
			assert.NotEmpty(t, buf.String())
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

func TestNilEngine(t *testing.T) {
	var engine *templatex.Engine
	var buf bytes.Buffer
	err := engine.Render(context.Background(), &buf, "test.html", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template engine not initialized")
}

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
			exts:    []string{"", ".tpl"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := templatex.New(tt.root, templatex.WithFuncs(tt.fns), templatex.WithExtensions(tt.exts...))
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
	engine, err := templatex.New("example/templates/", templatex.WithExtensions(".gohtml"))
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
			template: "greeter",
			data: pageData{
				Title:    "Test",
				Username: "John",
				Test:     "Message",
			},
			layouts: []string{"base_layout"},
			wantErr: false,
		},
		{
			name:     "Multiple layouts",
			template: "greeter",
			data: pageData{
				Title:    "Test",
				Username: "John",
				Test:     "Message",
			},
			layouts: []string{"app_layout", "base_layout"},
			wantErr: false,
		},
		{
			name:     "Non-existent template",
			template: "nonexistent",
			data:     nil,
			wantErr:  true,
		},
		{
			name:     "Non-existent layout",
			template: "greeter",
			data:     nil,
			layouts:  []string{"nonexistent"},
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
	engine, err := templatex.New("example/templates/", templatex.WithExtensions(".gohtml"))
	require.NoError(t, err)

	// Load translations
	err = ctxi18n.LoadWithDefault(testTranslations, "en")
	require.NoError(t, err)

	ctx, err := ctxi18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)

	result, err := engine.RenderString(ctx, "greeter", pageData{
		Title:    "Test",
		Username: "John",
		Test:     "Message",
	}, "base_layout")

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestRenderHTML(t *testing.T) {
	engine, err := templatex.New("example/templates/", templatex.WithExtensions(".gohtml"))
	require.NoError(t, err)

	// Load translations
	err = ctxi18n.LoadWithDefault(testTranslations, "en")
	require.NoError(t, err)

	ctx, err := ctxi18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)

	result, err := engine.RenderHTML(ctx, "greeter", pageData{
		Title:    "Test",
		Username: "John",
		Test:     "Message",
	}, "base_layout")

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestTemplateWithCustomFunctions(t *testing.T) {
	customFuncs := template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}

	_, err := templatex.New("example/templates/", templatex.WithExtensions(".gohtml"), templatex.WithFuncs(customFuncs))
	require.NoError(t, err)

	// Create a test template file with custom functions
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.gohtml")
	content := `{{.Text | upper}}`
	err = os.WriteFile(tempFile, []byte(content), 0644)
	require.NoError(t, err)

	// Create a new engine with the temp directory
	engine, err := templatex.New(tempDir, templatex.WithExtensions(".gohtml"), templatex.WithFuncs(customFuncs))
	require.NoError(t, err)

	var buf bytes.Buffer
	err = engine.Render(context.Background(), &buf, "test", struct{ Text string }{"hello"})
	assert.NoError(t, err)
	assert.Equal(t, "HELLO", buf.String())
}

func TestTemplateWithDifferentExtensions(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files with different extensions
	files := map[string]string{
		"test1.gohtml": "HTML template",
		"test2.tpl":    "TPL template",
		"test3.txt":    "TXT template",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test with specific extensions
	engine, err := templatex.New(tempDir, templatex.WithExtensions(".gohtml", ".tpl", ".txt"))
	require.NoError(t, err)

	// Try rendering each template
	for name, expectedContent := range files {
		ext := filepath.Ext(name)
		if ext == "" || ext == ".tpl" {
			var buf bytes.Buffer
			name = strings.TrimSuffix(name, ext)
			err := engine.Render(context.Background(), &buf, name, nil)
			assert.NoError(t, err)
			assert.Equal(t, expectedContent, buf.String())
		}
	}
}

func TestConcurrentRendering(t *testing.T) {
	engine, err := templatex.New("example/templates/", templatex.WithExtensions(".gohtml"))
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
			err := engine.Render(ctx, &buf, "greeter", data, "base_layout")
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
	err := engine.Render(context.Background(), &buf, "test", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template engine not initialized")
}

func TestDefaultFunctions(t *testing.T) {
	engine, err := templatex.New("example/templates/")
	require.NoError(t, err)

	tests := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		{
			name:     "upper function",
			template: `{{ "hello" | upper }}`,
			expected: "HELLO",
		},
		{
			name:     "lower function",
			template: `{{ "HELLO" | lower }}`,
			expected: "hello",
		},
		{
			name:     "title function",
			template: `{{ "hello world" | title }}`,
			expected: "Hello World",
		},
		{
			name:     "tern function / true",
			template: `{{ tern true "yes" "no" }}`,
			expected: "yes",
		},
		{
			name:     "tern function / false",
			template: `{{ tern false "yes" "no" }}`,
			expected: "no",
		},
		{
			name:     "trim function",
			template: `{{ " hello " | trim }}`,
			expected: "hello",
		},
		{
			name:     "replace function",
			template: `{{ replace "hello" "l" "w" }}`,
			expected: "hewwo",
		},
		{
			name:     "split and join function",
			template: `{{ split "a,b,c" "," | join "-" }}`,
			expected: "a-b-c",
		},
		{
			name:     "join function / string slice",
			template: `{{ join "-" . }}`,
			data:     []string{"a", "b", "c"},
			expected: "a-b-c",
		},
		{
			name:     "join function / string",
			template: `{{ join "-" . }}`,
			data:     "abc",
			expected: "abc",
		},
		{
			name:     "join function / interface slice",
			template: `{{ join "-" . }}`,
			data:     []interface{}{"a", "b", "c"},
			expected: "a-b-c",
		},
		{
			name:     "join function / nil",
			template: `{{ join "-" . }}`,
			data:     nil,
			expected: "",
		},
		{
			name:     "contains function",
			template: `{{ contains "hello" "ll" }}`,
			expected: "true",
		},
		{
			name:     "len function",
			template: `{{ len "hello" }}`,
			expected: "5",
		},
		{
			name:     "len function / default value",
			template: `{{ len 1 }}`,
			expected: "0",
		},
		{
			name:     "default function",
			template: `{{ "" | default "empty" }}`,
			expected: "empty",
		},
		{
			name:     "default function / nil",
			template: `{{ . | default "empty" }}`,
			data:     nil,
			expected: "empty",
		},
		{
			name:     "default function / pointer",
			template: `{{ . | default "empty" }}`,
			data: func() *string {
				s := "hello"
				return &s
			}(),
			expected: "hello",
		},
		{
			name:     "embed placeholder",
			template: `{{ embed }}`,
			expected: "",
		},
		{
			name:     "T placeholder",
			template: `{{ T "test.key" }}`,
			expected: "test.key",
		},
		{
			name:     "ctxVal placeholder",
			template: `{{ ctxVal "test" }}`,
			expected: "",
		},
		{
			name:     "debug",
			template: `{{ debug . }}`,
			data:     "hello",
			expected: "&#34;hello&#34;",
		},
		{
			name:     "debug / nil",
			template: `{{ debug . }}`,
			data:     nil,
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := template.New("test").Funcs(engine.GetFuncMap())
			tmpl, err := tmpl.Parse(tt.template)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestTemplateFunctions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		{
			name:     "hasPrefix function",
			template: `{{ hasPrefix "hello world" "hello" }}`,
			expected: "true",
		},
		{
			name:     "hasSuffix function",
			template: `{{ hasSuffix "hello world" "world" }}`,
			expected: "true",
		},
		{
			name:     "repeat function",
			template: `{{ repeat "a" 3 }}`,
			expected: "aaa",
		},
		{
			name:     "len function with map",
			template: `{{ len . }}`,
			data:     map[string]interface{}{"a": 1, "b": 2},
			expected: "2",
		},
		{
			name:     "len function with slice",
			template: `{{ len . }}`,
			data:     []interface{}{1, 2, 3},
			expected: "3",
		},
		{
			name:     "htmlSafe function",
			template: `{{ "<p>hello</p>" | htmlSafe }}`,
			expected: "<p>hello</p>",
		},
		{
			name:     "isset function with nil",
			template: `{{ isset . }}`,
			data:     nil,
			expected: "false",
		},
		{
			name:     "isset function with value",
			template: `{{ isset . }}`,
			data:     "value",
			expected: "true",
		},
		{
			name:     "boolToString function",
			template: `{{ boolToString true }}`,
			expected: "true",
		},
	}

	engine, err := templatex.New("example/templates/")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := template.New("test").Funcs(engine.GetFuncMap())
			tmpl, err := tmpl.Parse(tt.template)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestDefaultValue(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		data         interface{}
		defaultValue interface{}
		expected     string
	}{
		{
			name:     "default with nil pointer",
			template: `{{ . | default "default" }}`,
			data:     (*string)(nil),
			expected: "default",
		},
		{
			name:     "default with empty interface",
			template: `{{ . | default "default" }}`,
			data:     interface{}(nil),
			expected: "default",
		},
		{
			name:     "default with zero int",
			template: `{{ . | default "default" }}`,
			data:     0,
			expected: "default",
		},
		{
			name:     "default with zero uint",
			template: `{{ . | default "default" }}`,
			data:     uint(0),
			expected: "default",
		},
		{
			name:     "default with zero float",
			template: `{{ . | default "default" }}`,
			data:     0.0,
			expected: "default",
		},
		{
			name:     "default with empty slice",
			template: `{{ . | default "default" }}`,
			data:     []string{},
			expected: "default",
		},
		{
			name:     "default with false bool",
			template: `{{ . | default "default" }}`,
			data:     false,
			expected: "default",
		},
	}

	engine, err := templatex.New("example/templates/")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := template.New("test").Funcs(engine.GetFuncMap())
			tmpl, err := tmpl.Parse(tt.template)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestCustomFunctions(t *testing.T) {
	engine, err := templatex.New(
		"example/templates/",
		templatex.WithFunc("customFunc", func() string { return "custom" }),
	)
	require.NoError(t, err)

	tmpl := template.New("test").Funcs(engine.GetFuncMap())
	tmpl, err = tmpl.Parse(`{{ customFunc }}`)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	require.NoError(t, err)
	assert.Equal(t, "custom", buf.String())
}

func TestTemplateWithLayouts(t *testing.T) {
	engine, err := templatex.New(
		"example/templates/",
		templatex.WithLayouts("base_layout", "app_layout"),
	)
	require.NoError(t, err)
	require.NotNil(t, engine)
}

func TestSafeFieldFunction(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		{
			name:     "valid field",
			template: `{{ safeField . "Name" }}`,
			data:     TestStruct{Name: "John"},
			expected: "John",
		},
		{
			name:     "invalid field",
			template: `{{ safeField . "Invalid" }}`,
			data:     TestStruct{Name: "John"},
			expected: "",
		},
		{
			name:     "field with fallback",
			template: `{{ safeField . "Invalid" "fallback" }}`,
			data:     TestStruct{Name: "John"},
			expected: "fallback",
		},
	}

	engine, err := templatex.New("example/templates/")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := template.New("test").Funcs(engine.GetFuncMap())
			tmpl, err := tmpl.Parse(tt.template)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestPrintIfFunctions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "printIf true",
			template: `{{ printIf true "show" }}`,
			expected: "show",
		},
		{
			name:     "printIf false",
			template: `{{ printIf false "show" }}`,
			expected: "",
		},
		{
			name:     "printIfElse true",
			template: `{{ printIfElse true "yes" "no" }}`,
			expected: "yes",
		},
		{
			name:     "printIfElse false",
			template: `{{ printIfElse false "yes" "no" }}`,
			expected: "no",
		},
	}

	engine, err := templatex.New("example/templates/")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := template.New("test").Funcs(engine.GetFuncMap())
			tmpl, err := tmpl.Parse(tt.template)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestTemplateCache(t *testing.T) {
	engine, err := templatex.New("example/templates/",
		templatex.WithExtensions(".gohtml"),
		templatex.WithHardCache(true),
		templatex.WithLayouts("base_layout"),
		templatex.WithLayoutCache(true),
	)
	require.NoError(t, err)
	require.NotNil(t, engine)

	// Test data
	data := pageData{
		Title:    "Test",
		Username: "John",
		Test:     "Message",
	}

	// First render - should cache the result
	var buf1 bytes.Buffer
	err = engine.Render(context.Background(), &buf1, "greeter", data, "base_layout")
	require.NoError(t, err)
	firstResult := buf1.String()
	require.NotEmpty(t, firstResult)

	// Modify the data
	data.Username = "Jane"

	// Second render - should use cached result
	var buf2 bytes.Buffer
	err = engine.Render(context.Background(), &buf2, "greeter", data, "base_layout")
	require.NoError(t, err)
	secondResult := buf2.String()
	require.NotEmpty(t, secondResult)

	// Results should be identical due to caching
	assert.Equal(t, firstResult, secondResult)
	assert.Contains(t, secondResult, "John")    // Should contain original name
	assert.NotContains(t, secondResult, "Jane") // Should not contain modified name
}

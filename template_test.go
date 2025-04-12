package templatex_test

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmitrymomot/templatex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// We'll use "locale" as the context key since that's what the template engine expects
var langKey = "locale"

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

	// We no longer need to load translations from files
	// since we're using a custom translator function

	ctx := context.WithValue(context.Background(), langKey, "en")

	tests := []struct {
		name     string
		template string
		data     any
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
	// No need to load translations anymore

	ctx := context.WithValue(context.Background(), langKey, "en")

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
	// No need to load translations anymore

	ctx := context.WithValue(context.Background(), langKey, "en")

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
	// No need to load translations anymore

	ctx := context.WithValue(context.Background(), langKey, "en")

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
		data     any
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
			data:     []any{"a", "b", "c"},
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
		data     any
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
			data:     map[string]any{"a": 1, "b": 2},
			expected: "2",
		},
		{
			name:     "len function with slice",
			template: `{{ len . }}`,
			data:     []any{1, 2, 3},
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
		data         any
		defaultValue any
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
			data:     any(nil),
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
		data     any
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
	// Create a custom translator that returns username in the greeting
	customTranslator := func(lang, key string, args ...string) string {
		if key == "greeting" {
			return "John"
		}
		return key
	}

	engine, err := templatex.New("example/templates/",
		templatex.WithExtensions(".gohtml"),
		templatex.WithHardCache(true),
		templatex.WithLayouts("base_layout"),
		templatex.WithLayoutCache(true),
		templatex.WithTranslator(customTranslator),
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

// LangKey is a custom type for language keys to avoid SA1029 linter error
type LangKey string

// TransKey is a custom type for translation keys to avoid SA1029 linter error
type TransKey string

// DisplayKey is a custom type for display keys to avoid SA1029 linter error
type DisplayKey string

func TestTranslationInLayout(t *testing.T) {
	// Setup test environment
	
	// Create a custom translator that uses ctxi18n
	customTranslator := func(lang, key string, args ...string) string {
		// Create map of expected translations for testing
		translations := map[LangKey]map[TransKey]string{
			LangKey("en"): {
				TransKey("layout.title"):  "Test Title",
				TransKey("layout.header"): "Test Header",
				TransKey("layout.footer"): "Test Footer",
				TransKey("greeting"):      "Hello, John",
				TransKey("welcome"):       "Welcome to our awesome app!",
			},
			LangKey("es"): {
				TransKey("layout.title"):  "Título de Prueba",
				TransKey("layout.header"): "Encabezado de Prueba",
				TransKey("layout.footer"): "Pie de Página de Prueba",
				TransKey("greeting"):      "Hola, John",
				TransKey("welcome"):       "¡Bienvenido a nuestra increíble aplicación!",
			},
		}
		
		// Return translation based on language and key
		if langTranslations, ok := translations[LangKey(lang)]; ok {
			if translation, ok := langTranslations[TransKey(key)]; ok {
				return translation
			}
		}
		
		// Default fallback
		return key
	}
	
	engine, err := templatex.New("example/templates/", 
		templatex.WithExtensions(".gohtml"),
		templatex.WithTranslator(customTranslator),
	)
	require.NoError(t, err)
	require.NotNil(t, engine)

	// No need to load translations anymore - we use a custom translator

	// Test cases for different languages
	tests := []struct {
		name     string
		locale   string
		expected map[DisplayKey]string
	}{
		{
			name:   "English translations",
			locale: "en",
			expected: map[DisplayKey]string{
				DisplayKey("title"):    "Test Title",
				DisplayKey("header"):   "Test Header",
				DisplayKey("footer"):   "Test Footer",
				DisplayKey("greeting"): "Hello, John",
				DisplayKey("welcome"):  "Welcome to our awesome app!",
			},
		},
		{
			name:   "Spanish translations",
			locale: "es",
			expected: map[DisplayKey]string{
				DisplayKey("title"):    "Título de Prueba",
				DisplayKey("header"):   "Encabezado de Prueba",
				DisplayKey("footer"):   "Pie de Página de Prueba",
				DisplayKey("greeting"): "Hola, John",
				DisplayKey("welcome"):  "¡Bienvenido a nuestra increíble aplicación!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with locale key to store language
			ctx := context.WithValue(context.Background(), langKey, tt.locale)

			// Render the template with the trans_layout
			var buf bytes.Buffer
			err = engine.Render(ctx, &buf, "greeter", pageData{
				Title:    "Test Page",
				Username: "John",
				Test:     "Test Message",
			}, "trans_layout")
			require.NoError(t, err)

			// Verify that translations are present in the output
			result := buf.String()
			for key, expectedText := range tt.expected {
				assert.Contains(t, result, expectedText,
					"Translation for '%s' not found in %s locale", string(key), tt.locale)
			}

			// Verify that translations from other locales are not present
			for _, otherTest := range tests {
				if otherTest.locale != tt.locale {
					for _, unexpectedText := range otherTest.expected {
						assert.NotContains(t, result, unexpectedText,
							"Found translation from %s locale in %s locale output",
							otherTest.locale, tt.locale)
					}
				}
			}
		})
	}
}

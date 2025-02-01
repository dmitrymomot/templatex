# templatex

[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/dmitrymomot/templatex)](https://github.com/dmitrymomot/templatex)
[![Go Reference](https://pkg.go.dev/badge/github.com/dmitrymomot/templatex.svg)](https://pkg.go.dev/github.com/dmitrymomot/templatex)
[![License](https://img.shields.io/github/license/dmitrymomot/templatex)](https://github.com/dmitrymomot/templatex/blob/main/LICENSE)

[![Tests](https://github.com/dmitrymomot/templatex/actions/workflows/tests.yml/badge.svg)](https://github.com/dmitrymomot/templatex/actions/workflows/tests.yml)
[![codecov](https://codecov.io/github/dmitrymomot/templatex/graph/badge.svg?token=V5XK8QQYIQ)](https://codecov.io/github/dmitrymomot/templatex)
[![CodeQL Analysis](https://github.com/dmitrymomot/templatex/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/dmitrymomot/templatex/actions/workflows/codeql-analysis.yml)
[![GolangCI Lint](https://github.com/dmitrymomot/templatex/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/dmitrymomot/templatex/actions/workflows/golangci-lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmitrymomot/templatex)](https://goreportcard.com/report/github.com/dmitrymomot/templatex)

A high-performance Go template engine with support for layouts, caching, internationalization (i18n), and extensive template functions.

## Features

- **High-Performance Template Engine**

    - Thread-safe concurrent rendering with sync.RWMutex
    - Efficient buffer pooling to minimize memory allocations
    - Smart caching system with support for both soft and hard caching
    - Pre-compilation of common layouts for faster rendering

- **Advanced Layout System**

    - Support for multiple nested layouts
    - Layout chain caching for improved performance
    - Flexible template organization
    - Multiple template extension support

- **Internationalization (i18n)**

    - Built-in integration with `ctxi18n`
    - Context-aware translation functions
    - Support for multiple locales
    - Dynamic language switching

- **Rich Template Functions**

    - String manipulation and formatting
    - Conditional rendering helpers
    - Type conversion and safety checks
    - Debugging utilities
    - Custom function support

- **Multiple Output Formats**
    - Direct writing to `http.ResponseWriter`
    - String output with `RenderString`
    - Safe HTML output with `RenderHTML`
    - Support for custom writers

## Installation

```bash
go get github.com/dmitrymomot/templatex
```

## Usage

### Basic Setup

```go
import "github.com/dmitrymomot/templatex"

// Initialize with default settings
engine, err := templatex.New("templates/")
if err != nil {
    panic(err)
}

// Initialize with options
engine, err := templatex.New("templates/",
    templatex.WithLayouts("app_layout", "base_layout"),
    templatex.WithExtensions(".gohtml"),
)
```

### Template Structure

```
templates/
├── base_layout.gohtml
├── app_layout.gohtml
└── pages/
    └── greeter.gohtml
```

### Layout System

```html
<!-- base_layout.gohtml -->
<!DOCTYPE html>
<html>
    <head>
        <title>{{.Title}}</title>
    </head>
    <body>
        {{embed}}
    </body>
</html>

<!-- app_layout.gohtml -->
<div class="container">{{embed}}</div>

<!-- greeter.gohtml -->
<h1>Welcome, {{.Username}}!</h1>
```

### Rendering Templates

```go
// Render to http.ResponseWriter
data := struct {
    Title    string
    Username string
}{
    Title:    "Welcome",
    Username: "John Doe",
}

// With layouts
err := engine.Render(ctx, w, "greeter", data, "app_layout", "base_layout")

// Render to string
str, err := engine.RenderString(ctx, "greeter", data, "app_layout", "base_layout")

// Render as HTML
html, err := engine.RenderHTML(ctx, "greeter", data, "app_layout", "base_layout")
```

### Template Functions

#### String Functions

```gotemplate
{{upper .Text}}                // Convert string to uppercase
{{lower .Text}}                // Convert string to lowercase
{{title .Text}}                // Convert string to title case
{{trim .Text}}                 // Remove whitespace from both ends
{{replace .Text "old" "new"}}  // Replace all occurrences of "old" with "new"
{{split .Text ","}}            // Split string by separator
{{join .Array ","}}            // Join array elements with separator
{{contains .Text "substr"}}    // Check if string contains substring
{{hasPrefix .Text "prefix"}}   // Check if string starts with prefix
{{hasSuffix .Text "suffix"}}   // Check if string ends with suffix
{{truncate .Text 50}}          // Truncate text to specified length
{{stripHTML .Text}}            // Remove HTML tags from text
{{wordCount .Text}}            // Count words in text
{{slug .Text}}                 // Convert text to URL-friendly slug
```

#### Date/Time Functions

```gotemplate
{{now}}                                    // Current time
{{formatDate .Date "2006-01-02"}}         // Format date with layout
{{parseDate "2006-01-02" "2025-01-28"}}   // Parse date string with layout
{{addDays .Date 7}}                        // Add days to date
{{dateEqual .Date1 .Date2}}               // Check if dates are equal
{{dateBefore .Date1 .Date2}}              // Check if date1 is before date2
{{dateAfter .Date1 .Date2}}               // Check if date1 is after date2
{{duration .Duration}}                     // Format duration
{{humanizeTime .Date}}                     // Human-readable time difference
```

#### Number Functions

```gotemplate
{{formatNumber .Value 2}}      // Format number with decimal places
{{roundNumber .Value 2}}       // Round number to decimal places
{{ceil .Value}}                // Round up to nearest integer
{{floor .Value}}               // Round down to nearest integer
{{abs .Value}}                 // Absolute value
{{max .Val1 .Val2 .Val3}}     // Maximum value
{{min .Val1 .Val2 .Val3}}     // Minimum value
{{sum .Array}}                 // Sum of array values
{{avg .Array}}                 // Average of array values
```

#### Array/Slice Functions

```gotemplate
{{first .Array}}              // First element
{{last .Array}}               // Last element
{{rest .Array}}               // All elements except first
{{slice .Array 0 5}}          // Slice array from start to end index
{{reverse .Array}}            // Reverse array
{{sort .Array}}               // Sort array
{{unique .Array}}             // Remove duplicates
{{inSlice .Value .Array}}     // Check if value is in array
{{filter .Func .Array}}       // Filter array by function
{{map .Func .Array}}          // Map function over array
{{count .Array}}              // Count array elements
```

#### URL Functions

```gotemplate
{{urlEncode .Text}}           // URL encode string
{{urlDecode .Text}}           // URL decode string
{{parseURL .URL}}             // Parse URL into components
{{buildQuery .Params}}        // Build URL query string from map
{{isValidURL .URL}}           // Check if URL is valid
```

#### Conditional Functions

```gotemplate
{{tern .Condition .TrueVal .FalseVal}}  // Ternary operator
{{printIf .Condition .Value}}            // Print value if condition is true
{{printIfElse .Condition .Val1 .Val2}}   // Print val1 if true, val2 if false
```

#### HTML Functions

```gotemplate
{{escapeHTML .Text}}          // Escape HTML special characters
{{unescapeHTML .Text}}        // Unescape HTML special characters
```

### Advanced Usage

#### Caching Configuration

```go
// Enable hard caching for maximum performance
engine, err := templatex.New("templates/",
    templatex.WithHardCache(true),
)

// Use soft caching (default) for dynamic content
engine, err := templatex.New("templates/",
    templatex.WithHardCache(false),
)
```

#### Custom Function Registration

```go
// Add a single custom function
engine, err := templatex.New("templates/",
    templatex.WithFunc("myFunc", func(s string) string {
        return strings.ToUpper(s)
    }),
)

// Add multiple custom functions
customFuncs := template.FuncMap{
    "formatDate": func(t time.Time) string {
        return t.Format("2006-01-02")
    },
    "calculateTotal": func(items []float64) float64 {
        var total float64
        for _, item := range items {
            total += item
        }
        return total
    },
}

engine, err := templatex.New("templates/",
    templatex.WithFuncs(customFuncs),
)
```

#### Layout Chain Example

```go
// Template structure
templates/
├── layouts/
│   ├── base.gohtml      // Base HTML structure
│   ├── app.gohtml       // Application wrapper
│   └── section.gohtml   // Section-specific layout
└── pages/
    └── dashboard.gohtml // Content template

// Render with multiple layouts
err := engine.Render(ctx, w, "pages/dashboard", data,
    "layouts/section",
    "layouts/app",
    "layouts/base",
)
```

#### Internationalization Setup

```go
// Load translations from embedded files
//go:embed translations/*.yml
var translations embed.FS

// Initialize with default language
if err := ctxi18n.LoadWithDefault(translations, "en"); err != nil {
    panic(err)
}

// Add middleware for language detection
r.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        lang := r.Header.Get("Accept-Language")
        ctx := ctxi18n.WithLocale(r.Context(), lang)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
})

// Use in templates
{{T "welcome" "name" .Username}}
```

### Performance Tips

1. **Use Hard Caching** for static content that doesn't change frequently:

    ```go
    templatex.WithHardCache(true)
    ```

2. **Precompile Common Layouts** to reduce parsing overhead:

    ```go
    templatex.WithLayouts("common", "base")
    ```

3. **Optimize Template Structure** by minimizing the number of nested templates and layouts

4. **Use Buffer Pooling** (automatically handled by the engine) to reduce memory allocations

5. **Implement Proper Error Handling** to maintain stability under high load:
    ```go
    if err := engine.Render(ctx, w, "template", data); err != nil {
        log.Printf("Template rendering error: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    ```

## Complete Example

```go
package main

import (
    "embed"
    "net/http"

    "github.com/dmitrymomot/templatex"
    "github.com/go-chi/chi/v5"
    "github.com/invopop/ctxi18n"
)

//go:embed translations/*.yml
var translations embed.FS

func main() {
    r := chi.NewRouter()

    // Initialize template engine
    engine, err := templatex.New("templates/",
        templatex.WithLayouts("app_layout", "base_layout"),
    )
    if err != nil {
        panic(err)
    }

    // Load translations
    if err := ctxi18n.LoadWithDefault(translations, "en"); err != nil {
        panic(err)
    }

    r.Use(Localization("en"))

    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        data := struct {
            Title    string
            Username string
        }{
            Title:    "Welcome",
            Username: "John Doe",
        }

        if err := engine.Render(r.Context(), w, "greeter", data,
            "app_layout", "base_layout"); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

    http.ListenAndServe(":8080", r)
}
```

## Performance

The package includes several optimizations:

- Template caching
- Layout chain pre-computation
- Buffer pooling
- Concurrent rendering support

Benchmark results:

```
go test -bench=. -benchmem .

goos: darwin
goarch: arm64
pkg: github.com/dmitrymomot/templatex
cpu: Apple M3 Max
BenchmarkTemplateRenderWithCache/WithoutHardCache-14         	  560354	      2019 ns/op	    2193 B/op	      32 allocs/op
BenchmarkTemplateRenderWithCache/WithHardCache-14            	 2676073	       443.7 ns/op	     904 B/op	       9 allocs/op
BenchmarkTemplateRenderParallelWithCache/WithoutHardCache-14 	  852949	      1224 ns/op	    2197 B/op	      32 allocs/op
BenchmarkTemplateRenderParallelWithCache/WithHardCache-14    	 3049983	       387.9 ns/op	     905 B/op	       9 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithoutHardCache/SingleLayout-14         	  539823	      2033 ns/op	    2041 B/op	      31 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithoutHardCache/TwoLayouts-14           	  554419	      2115 ns/op	    2193 B/op	      32 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithoutHardCache/ThreeLayouts-14         	  545581	      2137 ns/op	    2201 B/op	      32 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithHardCache/SingleLayout-14            	 2895764	       408.6 ns/op	     728 B/op	       8 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithHardCache/TwoLayouts-14              	 2431064	       500.1 ns/op	     904 B/op	       9 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithHardCache/ThreeLayouts-14            	 2386470	       500.9 ns/op	     912 B/op	       9 allocs/op
BenchmarkTemplateRenderString/WithoutHardCache-14                                   	  517159	      2118 ns/op	    2195 B/op	      32 allocs/op
BenchmarkTemplateRenderString/WithHardCache-14                                      	 2327929	       516.5 ns/op	     905 B/op	       9 allocs/op
BenchmarkTemplateRenderHTML/WithoutHardCache-14                                     	  529089	      2126 ns/op	    2195 B/op	      32 allocs/op
BenchmarkTemplateRenderHTML/WithHardCache-14                                        	 2334886	       515.4 ns/op	     905 B/op	       9 allocs/op
PASS
ok  	github.com/dmitrymomot/templatex	20.005s
```

## License

Licensed under the Apache 2.0 License. See [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

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

- Efficient layout system with template caching
- Built-in i18n support via `ctxi18n`
- Thread-safe concurrent rendering
- Rich set of template functions
- Support for multiple output formats (Writer, String, HTML)
- Buffer pooling for optimal performance
- Custom function support
- Multiple template extension support
- Comprehensive error handling

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

### Built-in Template Functions

```go
// String operations
{{upper .Text}}        // Convert to uppercase
{{lower .Text}}        // Convert to lowercase
{{title .Text}}        // Convert to title case
{{trim .Text}}         // Trim whitespace
{{replace .Text "old" "new"}}
{{split .Text ","}}
{{join .Array ","}}

// Conditionals
{{tern .Condition "true" "false"}}
{{isset .Value}}
{{printIf .Condition .Value}}
{{printIfElse .Condition .TrueValue .FalseValue}}

// String checks
{{contains .Text "substring"}}
{{hasPrefix .Text "prefix"}}
{{hasSuffix .Text "suffix"}}

// Other utilities
{{len .Collection}}
{{htmlSafe .HTML}}
{{debug .Data}}        // Pretty print for debugging
{{safeField .Struct "FieldName" "default"}}
```

### Internationalization

```go
// Load translations
//go:embed translations/*.yml
var translations embed.FS

if err := ctxi18n.LoadWithDefault(translations, "en"); err != nil {
    panic(err)
}

// Use in templates
{{T "greeting" "name" .Username}}
```

### Custom Functions

```go
customFuncs := template.FuncMap{
    "customFunc": func(s string) string {
        return strings.ToUpper(s)
    },
}

engine, err := templatex.New("templates/",
    templatex.WithFuncs(customFuncs),
)
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

BenchmarkTemplateRenderWithCache/WithoutHardCache-14         	  577982	      1858 ns/op	    2145 B/op	      29 allocs/op
BenchmarkTemplateRenderWithCache/WithHardCache-14            	 3807852	       314.5 ns/op	     840 B/op	       6 allocs/op
BenchmarkTemplateRenderParallelWithCache/WithoutHardCache-14 	  930882	      1171 ns/op	    2148 B/op	      29 allocs/op
BenchmarkTemplateRenderParallelWithCache/WithHardCache-14    	 3557242	       332.6 ns/op	     841 B/op	       6 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithoutHardCache/SingleLayout-14         	  587005	      1896 ns/op	    1993 B/op	      28 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithoutHardCache/TwoLayouts-14           	  591994	      1954 ns/op	    2145 B/op	      29 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithoutHardCache/ThreeLayouts-14         	  596881	      1975 ns/op	    2153 B/op	      29 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithHardCache/SingleLayout-14            	 4240288	       283.7 ns/op	     680 B/op	       5 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithHardCache/TwoLayouts-14              	 3473145	       349.3 ns/op	     840 B/op	       6 allocs/op
BenchmarkTemplateRenderComplexityWithCache/WithHardCache/ThreeLayouts-14            	 3312826	       362.4 ns/op	     864 B/op	       6 allocs/op
BenchmarkTemplateRenderString/WithoutHardCache-14                                   	  564861	      1982 ns/op	    2147 B/op	      29 allocs/op
BenchmarkTemplateRenderString/WithHardCache-14                                      	 3230229	       371.7 ns/op	     841 B/op	       6 allocs/op
BenchmarkTemplateRenderHTML/WithoutHardCache-14                                     	  578750	      2080 ns/op	    2147 B/op	      29 allocs/op
BenchmarkTemplateRenderHTML/WithHardCache-14                                        	 3294721	       370.6 ns/op	     841 B/op	       6 allocs/op
PASS
ok  	github.com/dmitrymomot/templatex	19.246s
```

## License

Licensed under the Apache 2.0 License. See [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

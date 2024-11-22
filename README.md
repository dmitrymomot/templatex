# templatex

[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/dmitrymomot/templatex)](https://github.com/dmitrymomot/templatex)
[![Go Reference](https://pkg.go.dev/badge/github.com/dmitrymomot/templatex.svg)](https://pkg.go.dev/github.com/dmitrymomot/templatex)
[![License](https://img.shields.io/github/license/dmitrymomot/templatex)](https://github.com/dmitrymomot/templatex/blob/main/LICENSE)

[![Tests](https://github.com/dmitrymomot/templatex/actions/workflows/tests.yml/badge.svg)](https://github.com/dmitrymomot/templatex/actions/workflows/tests.yml)
[![CodeQL Analysis](https://github.com/dmitrymomot/templatex/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/dmitrymomot/templatex/actions/workflows/codeql-analysis.yml)
[![GolangCI Lint](https://github.com/dmitrymomot/templatex/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/dmitrymomot/templatex/actions/workflows/golangci-lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmitrymomot/templatex)](https://goreportcard.com/report/github.com/dmitrymomot/templatex)

A flexible and powerful Go template engine with support for layouts, internationalization (i18n), and useful template functions.

## Features

-   Layout system with nested templates
-   Built-in i18n support using `ctxi18n`
-   Rich set of template functions
-   Thread-safe template rendering
-   Support for HTML, string, and writer output
-   Easy to use API
-   Clean error handling

## Installation

```bash
go get github.com/dmitrymomot/templatex
```

## Usage

### Basic Setup

```go
import "github.com/dmitrymomot/templatex"

// Initialize the template engine
templ, err := templatex.New("templates/", nil)
if err != nil {
    panic(err)
}
```

### Template Structure

The package supports a hierarchical template structure with layouts. Here's an example structure:

```
templates/
├── base_layout.html    # Base layout template
├── app_layout.html     # Application layout
├── footer.html         # Partial template
└── pages/
    └── home.html       # Content template
```

### Layout System

Templates can be nested using the `{{embed}}` function. This allows for flexible layout composition:

```html
<!-- base_layout.html -->
<!doctype html>
<html>
    <head>
        <title>{{.Title}}</title>
    </head>
    <body>
        {{embed}}
        <!-- Content will be injected here -->
        {{template "footer" .}}
    </body>
</html>

<!-- app_layout.html -->
<div class="container">
    {{embed}}
    <!-- Page content will be injected here -->
</div>

<!-- pages/home.html -->
<div class="content">
    <h1>{{.Title}}</h1>
    <p>Welcome, {{.Username}}!</p>
</div>
```

### Rendering Templates

```go
// Render to http.ResponseWriter
err := templ.Render(ctx, w, "pages/home.html", data, "app_layout.html", "base_layout.html")

// Render to string
str, err := templ.RenderString(ctx, "pages/home.html", data, "app_layout.html", "base_layout.html")

// Render to template.HTML
html, err := templ.RenderHTML(ctx, "pages/home.html", data, "app_layout.html", "base_layout.html")
```

### Internationalization (i18n)

The package integrates with `ctxi18n` for internationalization support:

1. Create translation files:

```yaml
# en.yml
en:
    welcome: "Welcome to our app!"
    greeting: "Hello, %{name}!"

# es.yml
es:
    welcome: "¡Bienvenido a nuestra aplicación!"
    greeting: "¡Hola, %{name}!"
```

2. Load translations:

```go
import "github.com/invopop/ctxi18n"

//go:embed *.yml
var translations embed.FS

if err := ctxi18n.LoadWithDefault(translations, "en"); err != nil {
    panic(err)
}
```

3. Use translations in templates:

```html
<h1>{{T "greeting" "name" .Username}}</h1>
<p>{{T "welcome"}}</p>
```

### Built-in Template Functions

The package includes several useful template functions:

-   `upper`: Convert string to uppercase
-   `lower`: Convert string to lowercase
-   `title`: Convert string to title case
-   `tern`: Ternary operator
-   `trim`: Trim whitespace
-   `replace`: Replace string
-   `split`: Split string
-   `join`: Join strings
-   `contains`: Check if string contains substring
-   `hasPrefix`: Check if string starts with prefix
-   `hasSuffix`: Check if string ends with suffix
-   `repeat`: Repeat string
-   `len`: Get length of string/slice/map
-   `htmlSafe`: Mark string as safe HTML
-   `safeField`: Safely access struct fields

### Custom Functions

You can add custom functions when initializing the template engine:

```go
customFuncs := template.FuncMap{
    "myFunc": func(s string) string {
        return strings.ToUpper(s)
    },
}

templ, err := templatex.New("templates/", customFuncs)
```

## Example

Here's a complete example using Chi router:

```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/dmitrymomot/templatex"
)

func main() {
    r := chi.NewRouter()

    // Initialize templates
    templ, _ := templatex.New("templates/", nil)

    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        data := struct {
            Title    string
            Username string
        }{
            Title:    "Home",
            Username: "John Doe",
        }

        err := templ.Render(r.Context(), w, "pages/home.html", data,
            "app_layout.html", "base_layout.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

    http.ListenAndServe(":8080", r)
}
```

## License

Licensed under the Apache 2.0 License. See [LICENSE](LICENSE) for more information.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Benchmarks

```shell
go test -bench=. -benchmem .
goos: darwin
goarch: arm64
pkg: github.com/dmitrymomot/templatex
cpu: Apple M3 Max

BenchmarkTemplateRender-14              	                   36338	     32432 ns/op       35159 B/op	     302 allocs/op
BenchmarkTemplateRenderParallel-14      	                   55305	     22607 ns/op	   35221 B/op	     302 allocs/op
BenchmarkTemplateRenderComplexity/SingleLayout-14         	   45424	     25724 ns/op	   31560 B/op	     268 allocs/op
BenchmarkTemplateRenderComplexity/TwoLayouts-14           	   36391	     32812 ns/op	   35157 B/op	     302 allocs/op
BenchmarkTemplateRenderComplexity/ThreeLayouts-14         	   30818	     39263 ns/op	   36247 B/op	     317 allocs/op
PASS
ok  	github.com/dmitrymomot/templatex	7.757s
```

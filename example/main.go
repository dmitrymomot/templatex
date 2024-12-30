package main

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dmitrymomot/templatex"
	"github.com/go-chi/chi/v5"
	"github.com/invopop/ctxi18n"
)

//go:embed *.yml
var translations embed.FS

func main() {
	r := chi.NewRouter()
	r.Use(Localization("en"))

	templ, _ := templatex.New("templates/", nil)

	// Load translations
	if err := ctxi18n.LoadWithDefault(translations, "en"); err != nil {
		panic(err)
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Title    string
			Username string
			Test     string
		}{
			Title:    "Contacts",
			Username: "John Doe",
		}

		// Execute the template
		if err := templ.Render(r.Context(), w, "greeter", data, "app_layout", "base_layout"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	fmt.Println("Server is running on http://localhost:8080")
	defer fmt.Println("Server stopped")
	if err := http.ListenAndServe(":8080", r); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

// Localization middleware is responsible for handling request localization by
// extracting the preferred language from the Accept-Language header and adding it
// to the context. If no language preference is specified in the header, it falls
// back to the provided default locale. The middleware uses ctxi18n package to
// manage locale-specific functionality.
func Localization(defaultLocale string) func(next http.Handler) http.Handler {
	if defaultLocale == "" {
		defaultLocale = "en"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the current language from the request
			acceptLanguage := r.Header.Get("Accept-Language")
			if acceptLanguage == "" {
				acceptLanguage = defaultLocale
			}

			// Add current language to the context
			ctx, err := ctxi18n.WithLocale(r.Context(), acceptLanguage)
			if err != nil {
				ctx = r.Context()
				slog.ErrorContext(ctx, "Failed to set locale", "error", err)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

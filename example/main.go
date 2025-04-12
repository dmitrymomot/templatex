package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dmitrymomot/templatex"
	"github.com/go-chi/chi/v5"
)

// No translations embed needed

// Use the LocaleContextKey from the main package
var localeKey = templatex.ContextLocaleKey

func main() {
	r := chi.NewRouter()
	r.Use(Localization("en"))

	// No translation loading needed with the simplified approach

	// Initialize template engine with default settings
	templ, _ := templatex.New("templates/", 
		templatex.WithLayouts("app_layout", "base_layout"),
	)

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
// back to the provided default locale.
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

			// Store the language directly in the context using localeKey
			ctx := context.WithValue(r.Context(), localeKey, acceptLanguage)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

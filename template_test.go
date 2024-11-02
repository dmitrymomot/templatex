package templatex_test

import (
	"context"
	"embed"
	"testing"

	"github.com/invopop/ctxi18n"

	"github.com/dmitrymomot/templatex"
)

//go:embed example/*.yml
var translations embed.FS

type pageData struct {
	Title    string
	Username string
	Test     string
}

func BenchmarkTemplateRender(b *testing.B) {
	// Initialize template engine
	templ, err := templatex.New("example/templates/", nil)
	if err != nil {
		b.Fatal(err)
	}

	// Load translations
	if err := ctxi18n.LoadWithDefault(translations, "en"); err != nil {
		b.Fatal(err)
	}

	// Setup test data
	data := pageData{
		Title:    "Contacts",
		Username: "John Doe",
		Test:     "Test message",
	}

	// Create a context with locale
	ctx, err := ctxi18n.WithLocale(context.Background(), "en")
	if err != nil {
		b.Fatal(err)
	}

	// Add a dummy writer that implements io.Writer
	w := &mockWriter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := templ.Render(ctx, w, "greeter.html", data, "app_layout.html", "base_layout.html")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// mockWriter implements io.Writer for testing
type mockWriter struct{}

func (w *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func BenchmarkTemplateRenderParallel(b *testing.B) {
	// Initialize template engine
	templ, err := templatex.New("example/templates/", nil)
	if err != nil {
		b.Fatal(err)
	}

	// Load translations
	if err := ctxi18n.LoadWithDefault(translations, "en"); err != nil {
		b.Fatal(err)
	}

	// Setup test data
	data := pageData{
		Title:    "Contacts",
		Username: "John Doe",
		Test:     "Test message",
	}

	// Create a context with locale
	ctx, err := ctxi18n.WithLocale(context.Background(), "en")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		w := &mockWriter{}
		for pb.Next() {
			err := templ.Render(ctx, w, "greeter.html", data, "app_layout.html", "base_layout.html")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkTemplateRenderComplexity(b *testing.B) {
	cases := []struct {
		name    string
		layouts []string
	}{
		{"SingleLayout", []string{"base_layout.html"}},
		{"TwoLayouts", []string{"app_layout.html", "base_layout.html"}},
		{"ThreeLayouts", []string{"footer.html", "app_layout.html", "base_layout.html"}},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			// Initialize template engine
			templ, err := templatex.New("example/templates/", nil)
			if err != nil {
				b.Fatal(err)
			}

			// Load translations
			if err := ctxi18n.LoadWithDefault(translations, "en"); err != nil {
				b.Fatal(err)
			}

			// Setup test data
			data := pageData{
				Title:    "Contacts",
				Username: "John Doe",
				Test:     "Test message",
			}

			// Create a context with locale
			ctx, err := ctxi18n.WithLocale(context.Background(), "en")
			if err != nil {
				b.Fatal(err)
			}

			w := &mockWriter{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := templ.Render(ctx, w, "greeter.html", data, tc.layouts...)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

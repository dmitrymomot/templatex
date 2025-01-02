package templatex_test

import (
	"context"
	"embed"
	"testing"

	"github.com/dmitrymomot/templatex"
	"github.com/invopop/ctxi18n"
)

//go:embed example/*.yml
var translations embed.FS

type pageData struct {
	Title    string
	Username string
	Test     string
}

// mockWriter implements io.Writer for testing
type mockWriter struct{}

func (w *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func BenchmarkTemplateRenderWithCache(b *testing.B) {
	benchmarks := []struct {
		name      string
		hardCache bool
	}{
		{"WithoutHardCache", false},
		{"WithHardCache", true},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Initialize template engine
			templ, err := templatex.New("example/templates/",
				templatex.WithLayouts("app_layout", "base_layout"),
				templatex.WithHardCache(bm.hardCache),
			)
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
				err := templ.Render(ctx, w, "greeter", data, "app_layout", "base_layout")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkTemplateRenderParallelWithCache(b *testing.B) {
	benchmarks := []struct {
		name      string
		hardCache bool
	}{
		{"WithoutHardCache", false},
		{"WithHardCache", true},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Initialize template engine
			templ, err := templatex.New("example/templates/",
				templatex.WithLayouts("app_layout", "base_layout"),
				templatex.WithHardCache(bm.hardCache),
			)
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
					err := templ.Render(ctx, w, "greeter", data, "app_layout", "base_layout")
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func BenchmarkTemplateRenderComplexityWithCache(b *testing.B) {
	cases := []struct {
		name    string
		layouts []string
	}{
		{"SingleLayout", []string{"base_layout"}},
		{"TwoLayouts", []string{"app_layout", "base_layout"}},
		{"ThreeLayouts", []string{"footer", "app_layout", "base_layout"}},
	}

	cacheSettings := []struct {
		name      string
		hardCache bool
	}{
		{"WithoutHardCache", false},
		{"WithHardCache", true},
	}

	for _, cache := range cacheSettings {
		b.Run(cache.name, func(b *testing.B) {
			for _, tc := range cases {
				b.Run(tc.name, func(b *testing.B) {
					// Initialize template engine
					templ, err := templatex.New("example/templates/",
						templatex.WithLayouts("app_layout", "base_layout"),
						templatex.WithHardCache(cache.hardCache),
					)
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
						err := templ.Render(ctx, w, "greeter", data, tc.layouts...)
						if err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}

func BenchmarkTemplateRenderString(b *testing.B) {
	benchmarks := []struct {
		name      string
		hardCache bool
	}{
		{"WithoutHardCache", false},
		{"WithHardCache", true},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Initialize template engine
			templ, err := templatex.New("example/templates/",
				templatex.WithLayouts("app_layout", "base_layout"),
				templatex.WithHardCache(bm.hardCache),
			)
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
			for i := 0; i < b.N; i++ {
				_, err := templ.RenderString(ctx, "greeter", data, "app_layout", "base_layout")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkTemplateRenderHTML(b *testing.B) {
	benchmarks := []struct {
		name      string
		hardCache bool
	}{
		{"WithoutHardCache", false},
		{"WithHardCache", true},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Initialize template engine
			templ, err := templatex.New("example/templates/",
				templatex.WithLayouts("app_layout", "base_layout"),
				templatex.WithHardCache(bm.hardCache),
			)
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
			for i := 0; i < b.N; i++ {
				_, err := templ.RenderHTML(ctx, "greeter", data, "app_layout", "base_layout")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

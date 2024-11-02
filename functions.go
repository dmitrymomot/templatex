package templatex

import (
	"context"
	"fmt"
	"html/template"
	"reflect"
	"strings"

	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// getTranslator returns a translator function from context or falls back to returning the key
func getTranslator(ctx context.Context) func(string, ...string) string {
	l := ctxi18n.Locale(ctx)
	if l == nil {
		return func(key string, args ...string) string {
			if len(args) == 0 {
				return key
			}
			anyArgs := make([]any, len(args))
			for i, v := range args {
				anyArgs[i] = v
			}
			return fmt.Sprintf(key, anyArgs...)
		}
	}
	return func(s string, a ...string) string {
		argMap := make(i18n.M, len(a)/2)
		for i := 0; i < len(a); i += 2 {
			argMap[a[i]] = a[i+1]
		}
		return l.T(s, argMap)
	}
}

// safeField returns the value of a field from a struct if it exists and is accessible
func safeField(data interface{}, field string) string {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Struct {
		f := v.FieldByName(field)
		if f.IsValid() && f.CanInterface() {
			return f.Interface().(string)
		}
	}
	return "" // Default if field doesn't exist or isn't accessible
}

// defaultFuncs returns a FuncMap with default functions
func defaultFuncs() template.FuncMap {
	return template.FuncMap{
		// Add your custom functions here
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"title": func(s string) string {
			return cases.Title(language.Und).String(s)
		},
		"tern": func(cond bool, t, f interface{}) interface{} {
			if cond {
				return t
			}
			return f
		},
		"trim": func(s string) string {
			return strings.TrimSpace(s)
		},
		"replace": func(s, old, new string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"split": func(s, sep string) []string {
			return strings.Split(s, sep)
		},
		"join": func(elems []string, sep string) string {
			return strings.Join(elems, sep)
		},
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		"hasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
		"hasSuffix": func(s, suffix string) bool {
			return strings.HasSuffix(s, suffix)
		},
		"repeat": func(s string, count int) string {
			return strings.Repeat(s, count)
		},
		"len": func(v interface{}) int {
			switch val := v.(type) {
			case string:
				return len(val)
			case []interface{}:
				return len(val)
			case map[string]interface{}:
				return len(val)
			default:
				return 0
			}
		},
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
		"safeField": safeField,
	}
}

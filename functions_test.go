package templatex_test

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dmitrymomot/templatex"
)

func TestTemplateFunctions_Basic(t *testing.T) {
	engine, err := templatex.New("example/templates/")
	require.NoError(t, err)

	tests := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		// String manipulation tests
		{
			name:     "truncate function",
			template: `{{ "hello world" | truncate "5" }}`,
			expected: "hello...",
		},
		{
			name:     "camelCase function",
			template: `{{ "hello_world" | camelCase }}`,
			expected: "helloWorld",
		},
		{
			name:     "snakeCase function",
			template: `{{ "helloWorld" | snakeCase }}`,
			expected: "hello_world",
		},
		{
			name:     "kebabCase function",
			template: `{{ "helloWorld" | kebabCase }}`,
			expected: "hello-world",
		},
		{
			name:     "slugify function",
			template: `{{ "Hello World!" | slugify }}`,
			expected: "hello-world",
		},
		{
			name:     "matches function",
			template: `{{ matches "^[a-z]+$" "hello" }}`,
			expected: "true",
		},
		{
			name:     "replaceAll function",
			template: `{{ replaceAll "[aeiou]" "*" "hello" }}`,
			expected: "h*ll*",
		},

		// Type conversion tests
		{
			name:     "toString function",
			template: `{{ 123 | toString }}`,
			expected: "123",
		},
		{
			name:     "toInt function / string",
			template: `{{ "123" | toInt }}`,
			expected: "123",
		},
		{
			name:     "toInt function / float",
			template: `{{ 123.45 | toInt }}`,
			expected: "123",
		},
		{
			name:     "toFloat function / string",
			template: `{{ "123.45" | toFloat }}`,
			expected: "123.45",
		},
		{
			name:     "toFloat function / int",
			template: `{{ 123 | toFloat }}`,
			expected: "123",
		},
		{
			name:     "toBool function / string",
			template: `{{ "true" | toBool }}`,
			expected: "true",
		},
		{
			name:     "toBool function / int",
			template: `{{ 1 | toBool }}`,
			expected: "true",
		},
		{
			name:     "toJSON function",
			template: `{{ toJSON . }}`,
			data:     map[string]string{"hello": "world"},
			expected: `{"hello":"world"}`,
		},
		{
			name:     "fromJSON function",
			template: `{{ fromJSON "{\"hello\":\"world\"}" }}`,
			expected: `map[hello:world]`,
		},

		// Math function tests
		{
			name:     "add function",
			template: `{{ add 1 2 }}`,
			expected: "3",
		},
		{
			name:     "sub function",
			template: `{{ sub 5 2 }}`,
			expected: "3",
		},
		{
			name:     "mul function",
			template: `{{ mul 2 3 }}`,
			expected: "6",
		},
		{
			name:     "div function",
			template: `{{ div 6 2 }}`,
			expected: "3",
		},
		{
			name:     "mod function",
			template: `{{ mod 7 3 }}`,
			expected: "1",
		},
		{
			name:     "max function",
			template: `{{ max 1 2 }}`,
			expected: "2",
		},
		{
			name:     "min function",
			template: `{{ min 1 2 }}`,
			expected: "1",
		},
		{
			name:     "abs function",
			template: `{{ abs -5 }}`,
			expected: "5",
		},
		{
			name:     "ceil function",
			template: `{{ ceil 1.1 }}`,
			expected: "2",
		},
		{
			name:     "floor function",
			template: `{{ floor 1.9 }}`,
			expected: "1",
		},
		{
			name:     "round function",
			template: `{{ round 1.5 }}`,
			expected: "2",
		},
		{
			name:     "sum function",
			template: `{{ sum 1 2 3 4 5 }}`,
			expected: "15",
		},
		{
			name:     "avg function",
			template: `{{ avg 1 2 3 4 5 }}`,
			expected: "3",
		},
		{
			name:     "sequence function",
			template: `{{ range sequence 1 3 }}{{ . }}{{ end }}`,
			expected: "123",
		},

		// Date/Time function tests
		{
			name:     "now function type check",
			template: `{{ if now.IsZero }}zero{{ else }}not zero{{ end }}`,
			expected: "not zero",
		},
		{
			name:     "formatTime function",
			template: `{{ $t := now }}{{ formatTime $t "2006" }}`,
			expected: "2025",
		},
		{
			name:     "parseTime function",
			template: `{{ with parseTime "2025-01-29" "2006-01-02" }}{{ formatTime . "2006" }}{{ end }}`,
			expected: "2025",
		},
		{
			name:     "addDate function",
			template: `{{ $t := now }}{{ $t2 := addDate $t 1 0 0 }}{{ if dateAfter $t2 $t }}true{{ else }}false{{ end }}`,
			expected: "true",
		},
		{
			name:     "subDate function",
			template: `{{ $t := now }}{{ $t2 := subDate $t 1 0 0 }}{{ if dateBefore $t2 $t }}true{{ else }}false{{ end }}`,
			expected: "true",
		},
		{
			name:     "dateEqual function",
			template: `{{ $t := now }}{{ dateEqual $t $t }}`,
			expected: "true",
		},
		{
			name:     "dateBefore function",
			template: `{{ $t := now }}{{ $t2 := addDate $t 1 0 0 }}{{ dateBefore $t $t2 }}`,
			expected: "true",
		},
		{
			name:     "dateAfter function",
			template: `{{ $t := now }}{{ $t2 := subDate $t 1 0 0 }}{{ dateAfter $t $t2 }}`,
			expected: "true",
		},
		{
			name:     "dateBetween function",
			template: `{{ $t := now }}{{ $t1 := subDate $t 1 0 0 }}{{ $t2 := addDate $t 1 0 0 }}{{ dateBetween $t $t1 $t2 }}`,
			expected: "true",
		},
		{
			name:     "toUTC function",
			template: `{{ $t := now }}{{ $utc := toUTC $t }}{{ if $utc.IsZero }}zero{{ else }}not zero{{ end }}`,
			expected: "not zero",
		},
		{
			name:     "toLocal function",
			template: `{{ $t := now }}{{ $local := toLocal $t }}{{ if $local.IsZero }}zero{{ else }}not zero{{ end }}`,
			expected: "not zero",
		},
		{
			name:     "unix timestamp function",
			template: `{{ $t := now }}{{ if unix $t }}true{{ else }}false{{ end }}`,
			expected: "true",
		},
		{
			name:     "unix milli timestamp function",
			template: `{{ $t := now }}{{ if unixMilli $t }}true{{ else }}false{{ end }}`,
			expected: "true",
		},
		{
			name:     "duration parse function",
			template: `{{ with durationParse "1h" }}true{{ end }}`,
			expected: "true",
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

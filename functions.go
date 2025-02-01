package templatex

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// defaultFuncs returns a FuncMap with default functions
func defaultFuncs() template.FuncMap {
	return template.FuncMap{
		// String functions
		"upper":      upperString,
		"lower":      lowerString,
		"title":      titleString,
		"trim":       trimString,
		"replace":    replaceString,
		"split":      splitString,
		"join":       join,
		"contains":   containsString,
		"hasPrefix":  hasPrefixString,
		"hasSuffix":  hasSuffixString,
		"repeat":     repeatString,
		"truncate":   truncateString,
		"camelCase":  toCamelCase,
		"snakeCase":  toSnakeCase,
		"kebabCase":  toKebabCase,
		"slugify":    slugify,
		"matches":    regexMatches,
		"replaceAll": regexReplaceAll,

		// Type manipulation
		"len":          getLength,
		"tern":         ternary,
		"isset":        isSet,
		"boolToString": boolToStr,
		"default":      defaultValue,
		"safeField":    safeField,
		"toString":     toString,
		"toInt":        toInt,
		"toFloat":      toFloat,
		"toBool":       toBool,
		"toJSON":       toJSON,
		"fromJSON":     fromJSON,

		// Math functions
		"add":      add,
		"sub":      sub,
		"mul":      mul,
		"div":      div,
		"mod":      mod,
		"max":      max,
		"min":      min,
		"abs":      abs,
		"ceil":     ceil,
		"floor":    floor,
		"round":    round,
		"sum":      sum,
		"avg":      avg,
		"sequence": sequence,

		// Date/Time functions
		"now":           now,
		"formatTime":    formatTime,
		"parseTime":     parseTime,
		"addDate":       addDate,
		"subDate":       subDate,
		"dateEqual":     dateEqual,
		"dateBefore":    dateBefore,
		"dateAfter":     dateAfter,
		"dateBetween":   dateBetween,
		"toUTC":         toUTC,
		"toLocal":       toLocal,
		"unix":          unixTimestamp,
		"unixMilli":     unixMilliTimestamp,
		"durationParse": parseDuration,

		// Debug functions
		"debug":       prettyPrint,
		"printIf":     printIf,
		"printIfElse": printIfElse,

		// HTML functions
		"htmlSafe": toHTML,

		// Placeholders for context-related functions
		"embed":  emptyHTML,
		"T":      translate,
		"ctxVal": contextValue,
	}
}

// String manipulation functions
func upperString(s string) string {
	return strings.ToUpper(s)
}

func lowerString(s string) string {
	return strings.ToLower(s)
}

func titleString(s string) string {
	return cases.Title(language.Und).String(s)
}

func trimString(s string) string {
	return strings.TrimSpace(s)
}

func replaceString(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func splitString(s, sep string) []string {
	return strings.Split(s, sep)
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func hasPrefixString(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func hasSuffixString(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

func repeatString(s string, count int) string {
	return strings.Repeat(s, count)
}

func truncateString(length, str string) string {
	l, err := strconv.Atoi(length)
	if err != nil {
		return str
	}
	if len(str) <= l {
		return str
	}
	return str[:l] + "..."
}

func toCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i := 1; i < len(words); i++ {
		words[i] = strings.Title(words[i])
	}
	return strings.Join(words, "")
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && (unicode.IsUpper(r) || unicode.IsNumber(r)) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func toKebabCase(s string) string {
	return strings.ReplaceAll(toSnakeCase(s), "_", "-")
}

func slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	// Replace special characters with -
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	// Remove leading/trailing -
	s = strings.Trim(s, "-")
	return s
}

func regexMatches(pattern, s string) bool {
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

func regexReplaceAll(pattern, repl, s string) string {
	reg := regexp.MustCompile(pattern)
	return reg.ReplaceAllString(s, repl)
}

// Type manipulation functions
func ternary(cond bool, t, f interface{}) interface{} {
	if cond {
		return t
	}
	return f
}

func getLength(v interface{}) int {
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
}

func isSet(v interface{}) bool {
	return v != nil
}

func boolToStr(b bool) string {
	return fmt.Sprintf("%t", b)
}

func toHTML(html string) template.HTML {
	return template.HTML(html)
}

func emptyHTML() template.HTML {
	return ""
}

func translate(key string, args ...any) string {
	return key
}

func contextValue(key string) string {
	return ""
}

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

// ctxValue returns the value of a key from a context
// It returns an empty string if the key doesn't exist
// It's useful for getting values from a context in a template
func ctxValue(ctx context.Context) func(key string) string {
	return func(key string) string {
		if v := ctx.Value(key); v != nil {
			return fmt.Sprint(v)
		}
		return "" // Default if key doesn't exist
	}
}

// safeField returns the value of a field from a struct if it exists and is accessible
func safeField(data interface{}, field string, fallback ...string) string {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Struct {
		f := v.FieldByName(field)
		if f.IsValid() && f.CanInterface() {
			return f.Interface().(string)
		}
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return "" // Default if field doesn't exist or isn't accessible
}

// defaultValue returns the default value if the value is nil, empty, or zero.
// Usage: {{ .Value | default "default value" }}
func defaultValue(defaultValue, value interface{}) interface{} {
	// Handle nil case first
	if value == nil {
		return defaultValue
	}

	v := reflect.ValueOf(value)

	// Handle special case for pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return defaultValue
		}
		v = v.Elem()
	}

	// Check for zero/empty values based on type
	switch v.Kind() {
	case reflect.String:
		if strings.TrimSpace(v.String()) == "" {
			return defaultValue
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if v.Len() == 0 {
			return defaultValue
		}
	case reflect.Bool:
		if !v.Bool() {
			return defaultValue
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() == 0 {
			return defaultValue
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() == 0 {
			return defaultValue
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() == 0 {
			return defaultValue
		}
	case reflect.Interface:
		if v.IsNil() {
			return defaultValue
		}
	}
	return value
}

// prettyPrint returns a pretty-printed JSON string of the given value.
// If the value cannot be marshaled to JSON, it returns the value as a string.
// This function is useful for debugging purposes.
func prettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}

// printIf returns the data if the condition is true, otherwise it returns an empty string
// Usage: {{ printIf .Condition .Data }}
func printIf(cond bool, data any) string {
	if cond {
		return fmt.Sprintf("%v", data)
	}
	return ""
}

// printIfElse returns the data if the condition is true, otherwise it returns the elseData
// Usage: {{ printIfElse .Condition .Data .ElseData }}
// Example: {{ printIfElse .IsAdmin "Admin" "User" }}
func printIfElse(cond bool, data, elseData any) string {
	if cond {
		return fmt.Sprintf("%v", data)
	}
	return fmt.Sprintf("%v", elseData)
}

// reversed parameters is required to support variadic functions
func join(sep string, v interface{}) string {
	var strs []string
	switch slice := v.(type) {
	case []string:
		strs = slice
	case []interface{}:
		strs = make([]string, len(slice))
		for i, v := range slice {
			strs[i] = fmt.Sprint(v)
		}
	case string:
		strs = []string{slice}
	default:
		if v == nil {
			return ""
		}
		// Handle any other type by converting to string
		strs = []string{fmt.Sprint(v)}
	}
	return strings.Join(strs, sep)
}

// Type conversion functions
func toString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		b, _ := strconv.ParseBool(val)
		return b
	case int:
		return val != 0
	case float64:
		return val != 0
	default:
		return false
	}
}

func toJSON(v interface{}) template.HTML {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return template.HTML(b)
}

func fromJSON(s string) interface{} {
	var v interface{}
	_ = json.Unmarshal([]byte(s), &v)
	return v
}

// Math functions
func add(a, b interface{}) float64 {
	return toFloat(a) + toFloat(b)
}

func sub(a, b interface{}) float64 {
	return toFloat(a) - toFloat(b)
}

func mul(a, b interface{}) float64 {
	return toFloat(a) * toFloat(b)
}

func div(a, b interface{}) float64 {
	return toFloat(a) / toFloat(b)
}

func mod(a, b interface{}) float64 {
	return float64(int(toFloat(a)) % int(toFloat(b)))
}

func max(a, b interface{}) float64 {
	return math.Max(toFloat(a), toFloat(b))
}

func min(a, b interface{}) float64 {
	return math.Min(toFloat(a), toFloat(b))
}

func abs(a interface{}) float64 {
	return math.Abs(toFloat(a))
}

func ceil(a interface{}) float64 {
	return math.Ceil(toFloat(a))
}

func floor(a interface{}) float64 {
	return math.Floor(toFloat(a))
}

func round(a interface{}) float64 {
	return math.Round(toFloat(a))
}

func sum(numbers ...interface{}) float64 {
	var total float64
	for _, n := range numbers {
		total += toFloat(n)
	}
	return total
}

func avg(numbers ...interface{}) float64 {
	if len(numbers) == 0 {
		return 0
	}
	return sum(numbers...) / float64(len(numbers))
}

func sequence(start, end int) []int {
	if start > end {
		return nil
	}
	seq := make([]int, end-start+1)
	for i := range seq {
		seq[i] = start + i
	}
	return seq
}

// Date/Time functions
func now() time.Time {
	return time.Now()
}

func formatTime(t time.Time, layout string) string {
	return t.Format(layout)
}

func parseTime(value, layout string) (time.Time, error) {
	return time.Parse(layout, value)
}

func addDate(t time.Time, years, months, days int) time.Time {
	return t.AddDate(years, months, days)
}

func subDate(t time.Time, years, months, days int) time.Time {
	return t.AddDate(-years, -months, -days)
}

func dateEqual(a, b time.Time) bool {
	return a.Equal(b)
}

func dateBefore(a, b time.Time) bool {
	return a.Before(b)
}

func dateAfter(a, b time.Time) bool {
	return a.After(b)
}

func dateBetween(t, start, end time.Time) bool {
	return (t.After(start) || t.Equal(start)) && (t.Before(end) || t.Equal(end))
}

func toUTC(t time.Time) time.Time {
	return t.UTC()
}

func toLocal(t time.Time) time.Time {
	return t.Local()
}

func unixTimestamp(t time.Time) int64 {
	return t.Unix()
}

func unixMilliTimestamp(t time.Time) int64 {
	return t.UnixMilli()
}

func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

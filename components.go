package templatex

import (
	"bytes"
	"html/template"
	"sync"
)

// ComponentProps represents a map of properties for a component
type ComponentProps map[string]interface{}

var propsPool = sync.Pool{
	New: func() interface{} {
		return make(ComponentProps)
	},
}

// Props creates a new ComponentProps map with the given key-value pairs
func Props(pairs ...interface{}) ComponentProps {
	props := propsPool.Get().(ComponentProps)
	for i := 0; i < len(pairs); i += 2 {
		if i+1 < len(pairs) {
			props[pairs[i].(string)] = pairs[i+1]
		}
	}
	return props
}

// MergeProps merges multiple ComponentProps maps into a single map
func MergeProps(propsList ...ComponentProps) ComponentProps {
	result := Props()
	for _, props := range propsList {
		for k, v := range props {
			result[k] = v
		}
	}
	return result
}

// ReleaseProps returns a ComponentProps map to the pool
func ReleaseProps(props ComponentProps) {
	clear(props)
	propsPool.Put(props)
}

// Component function to be used in templates
func (tm *Engine) componentFunc(name string, props ComponentProps) (template.HTML, error) {
	tmpl := tm.templates.Lookup(name)
	if tmpl == nil {
		return "", ErrTemplateNotFound
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, props); err != nil {
		return "", err
	}

	defer ReleaseProps(props)
	return template.HTML(buf.String()), nil
}

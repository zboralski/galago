// Package trace provides types for trace event collection and analysis.
package trace

import "time"

// Tag represents a trace event category.
// Tags are stored without # prefix; the prefix is added on rendering.
type Tag string

// Standard tags for trace events.
const (
	Setter   Tag = "setter"
	Key      Tag = "key"
	XorNeon  Tag = "xor-neon"
	JniCall  Tag = "jni-call"
	Malloc   Tag = "malloc"
	String   Tag = "string"
	Crypto   Tag = "crypto"
	Network  Tag = "network"
	File     Tag = "file"
	Dynload  Tag = "dynload"
	Lua      Tag = "lua"
	Tolua    Tag = "tolua"
	Fallback Tag = "fallback"
	Libc     Tag = "libc"
	Pthread  Tag = "pthread"
	CxxAbi   Tag = "cxxabi"
	Android  Tag = "android"
	Printf   Tag = "printf"
	Locale   Tag = "locale"
)

// Tags is a collection of tags with helper methods.
type Tags []Tag

// Has returns true if the tag collection contains the given tag.
func (t Tags) Has(tag Tag) bool {
	for _, x := range t {
		if x == tag {
			return true
		}
	}
	return false
}

// Add adds a tag if not already present.
func (t *Tags) Add(tag Tag) {
	if !t.Has(tag) {
		*t = append(*t, tag)
	}
}

// Strings returns tags as strings with # prefix for display.
func (t Tags) Strings() []string {
	out := make([]string, len(t))
	for i, tag := range t {
		out[i] = "#" + string(tag)
	}
	return out
}

// Raw returns tags as strings without # prefix.
func (t Tags) Raw() []string {
	out := make([]string, len(t))
	for i, tag := range t {
		out[i] = string(tag)
	}
	return out
}

// Primary returns the first tag or empty string if none.
func (t Tags) Primary() Tag {
	if len(t) > 0 {
		return t[0]
	}
	return ""
}

// Annotations holds key-value metadata for trace events.
type Annotations map[string]string

// Set adds or updates an annotation.
func (a Annotations) Set(k, v string) {
	a[k] = v
}

// Get retrieves an annotation value.
func (a Annotations) Get(k string) string {
	return a[k]
}

// Has returns true if the annotation exists.
func (a Annotations) Has(k string) bool {
	_, ok := a[k]
	return ok
}

// Event represents a trace event with rich metadata.
type Event struct {
	PC          uint64      // Program counter (return address of stub call)
	Tags        Tags        // Multiple hashtags, first is primary
	Name        string      // Function name (e.g., "malloc", "FindClass")
	Detail      string      // Additional detail (e.g., "size=24", "class=com/foo/Bar")
	Annotations Annotations // Key-value metadata
	Timestamp   time.Time   // When the event occurred
}

// NewEvent creates a new trace event with the given parameters.
func NewEvent(pc uint64, category, name, detail string) *Event {
	return &Event{
		PC:          pc,
		Tags:        Tags{Tag(category)},
		Name:        name,
		Detail:      detail,
		Annotations: make(Annotations),
		Timestamp:   time.Now(),
	}
}

// AddTag adds a tag to the event.
func (e *Event) AddTag(tag Tag) {
	e.Tags.Add(tag)
}

// Annotate sets an annotation on the event.
func (e *Event) Annotate(k, v string) {
	if e.Annotations == nil {
		e.Annotations = make(Annotations)
	}
	e.Annotations.Set(k, v)
}

// PrimaryTag returns the primary (first) tag with # prefix.
func (e *Event) PrimaryTag() string {
	if len(e.Tags) > 0 {
		return "#" + string(e.Tags[0])
	}
	return ""
}

// Enricher enriches trace events based on category and name.
type Enricher func(e *Event)

// DefaultEnricher adds additional tags based on category and name.
func DefaultEnricher(e *Event) {
	if len(e.Tags) == 0 {
		return
	}

	category := string(e.Tags[0])

	switch category {
	case "setter":
		e.AddTag(Key)
		e.Annotate("type", "xxtea")

	case "jni":
		e.AddTag(JniCall)

	case "libc":
		switch e.Name {
		case "malloc", "calloc", "realloc", "free":
			e.AddTag(Malloc)
		case "memcpy", "memmove", "memset":
			e.AddTag(String)
		}

	case "lua":
		e.AddTag(Lua)

	case "tolua":
		e.AddTag(Tolua)

	case "network":
		e.AddTag(Network)

	case "dl":
		e.AddTag(Dynload)
	}
}

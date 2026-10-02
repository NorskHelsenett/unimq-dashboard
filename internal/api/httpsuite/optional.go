package httpsuite

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Optional[T any] struct {
	set   bool
	null  bool
	value T
}

func NewOptional[T any](value T) Optional[T] {
	return Optional[T]{set: true, null: false, value: value}
}

func (o *Optional[T]) UnmarshalJSON(b []byte) error {
	o.set = true

	// explicit null
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		o.null = true
		var zero T
		o.value = zero
		return nil
	}

	o.null = false
	return json.Unmarshal(b, &o.value)
}

func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.set {
		return []byte("null"), nil
	}
	if o.null {
		return []byte("null"), nil
	}
	return json.Marshal(o.value)
}

func (o Optional[T]) IsSet() bool  { return o.set }
func (o Optional[T]) IsNull() bool { return o.set && o.null }
func (o Optional[T]) Value() (T, bool) {
	return o.value, o.set && !o.null
}

// Any returns the value as interface{} to satisfy OptionalAny interface
func (o Optional[T]) Any() (interface{}, bool) {
	if !o.set || o.null {
		return nil, false
	}
	return o.value, true
}

func (o Optional[T]) String() string {
	switch {
	case !o.set:
		return "unset"
	case o.null:
		return "null"
	default:
		return fmt.Sprintf("value(%v)", o.value)
	}
}

func SetUpdate[T any](m map[string]any, column string, o Optional[T]) {
	if !o.IsSet() {
		return
	}
	if o.IsNull() {
		m[column] = nil
		return
	}
	if v, ok := o.Value(); ok {
		m[column] = v
	}
}

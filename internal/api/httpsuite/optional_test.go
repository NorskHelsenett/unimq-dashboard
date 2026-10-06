package httpsuite_test

import (
	"encoding/json"
	"testing"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type optionalFixture struct {
	Name httpsuite.Optional[string] `json:"name"`
	Age  httpsuite.Optional[int]    `json:"age"`
	Flag httpsuite.Optional[bool]   `json:"flag"`
}

func TestOptional_UnmarshalJSON_Unset(t *testing.T) {
	var f optionalFixture
	err := json.Unmarshal([]byte(`{}`), &f)
	require.NoError(t, err)

	assert.False(t, f.Name.IsSet())
	assert.False(t, f.Name.IsNull())
	val, ok := f.Name.Value()
	assert.False(t, ok)
	assert.Equal(t, "", val)
}

func TestOptional_UnmarshalJSON_ExplicitNull(t *testing.T) {
	var f optionalFixture
	err := json.Unmarshal([]byte(`{"name": null}`), &f)
	require.NoError(t, err)

	assert.True(t, f.Name.IsSet())
	assert.True(t, f.Name.IsNull())
	val, ok := f.Name.Value()
	assert.False(t, ok)
	assert.Equal(t, "", val, "value should be the zero value of T when explicitly null")
}

func TestOptional_UnmarshalJSON_ValuePresent(t *testing.T) {
	var f optionalFixture
	err := json.Unmarshal([]byte(`{"name": "High Queue Size", "age": 42, "flag": true}`), &f)
	require.NoError(t, err)

	assert.True(t, f.Name.IsSet())
	assert.False(t, f.Name.IsNull())
	val, ok := f.Name.Value()
	assert.True(t, ok)
	assert.Equal(t, "High Queue Size", val)

	ageVal, ok := f.Age.Value()
	assert.True(t, ok)
	assert.Equal(t, 42, ageVal)

	flagVal, ok := f.Flag.Value()
	assert.True(t, ok)
	assert.True(t, flagVal)
}

func TestOptional_UnmarshalJSON_TypeMismatchErrors(t *testing.T) {
	var f optionalFixture
	// "age" expects an int, giving it a string should produce a decode error.
	err := json.Unmarshal([]byte(`{"age": "not-a-number"}`), &f)
	assert.Error(t, err)
}

func TestOptional_UnmarshalJSON_MalformedJSON(t *testing.T) {
	var f optionalFixture
	err := json.Unmarshal([]byte(`{"name": `), &f)
	assert.Error(t, err)
}

func TestOptional_Any(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		var o httpsuite.Optional[string]
		val, ok := o.Any()
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("null", func(t *testing.T) {
		o := httpsuite.NewOptional("ignored")
		err := o.UnmarshalJSON([]byte("null"))
		require.NoError(t, err)
		val, ok := o.Any()
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("value", func(t *testing.T) {
		o := httpsuite.NewOptional("queue_size")
		val, ok := o.Any()
		assert.True(t, ok)
		assert.Equal(t, "queue_size", val)
	})
}

func TestOptional_MarshalJSON_Unset(t *testing.T) {
	var o httpsuite.Optional[string]
	b, err := o.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, "null", string(b))
}

func TestOptional_MarshalJSON_Null(t *testing.T) {
	var o httpsuite.Optional[string]
	err := o.UnmarshalJSON([]byte("null"))
	require.NoError(t, err)

	b, err := o.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, "null", string(b))
}

func TestOptional_MarshalJSON_Value(t *testing.T) {
	o := httpsuite.NewOptional(1000.0)
	b, err := o.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, "1000", string(b))
}

func TestOptional_MarshalJSON_RoundTrip(t *testing.T) {
	original := optionalFixture{
		Name: httpsuite.NewOptional("High Queue Size"),
		Age:  httpsuite.NewOptional(5),
		Flag: httpsuite.NewOptional(true),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var roundTripped optionalFixture
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	assert.Equal(t, original, roundTripped)
}

func TestOptional_String(t *testing.T) {
	var unset httpsuite.Optional[string]
	assert.Equal(t, "unset", unset.String())

	nullVal := httpsuite.NewOptional("x")
	require.NoError(t, nullVal.UnmarshalJSON([]byte("null")))
	assert.Equal(t, "null", nullVal.String())

	value := httpsuite.NewOptional("queue_size")
	assert.Equal(t, "value(queue_size)", value.String())
}

func TestSetUpdate(t *testing.T) {
	t.Run("unset field is not added to the map", func(t *testing.T) {
		m := map[string]any{}
		var o httpsuite.Optional[string]
		httpsuite.SetUpdate(m, "name", o)
		assert.NotContains(t, m, "name")
	})

	t.Run("null field is set to nil in the map", func(t *testing.T) {
		m := map[string]any{}
		o := httpsuite.NewOptional("ignored")
		require.NoError(t, o.UnmarshalJSON([]byte("null")))
		httpsuite.SetUpdate(m, "name", o)
		require.Contains(t, m, "name")
		assert.Nil(t, m["name"])
	})

	t.Run("value field is set to the value in the map", func(t *testing.T) {
		m := map[string]any{}
		o := httpsuite.NewOptional("High Queue Size")
		httpsuite.SetUpdate(m, "name", o)
		assert.Equal(t, "High Queue Size", m["name"])
	})
}

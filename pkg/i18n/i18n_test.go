package i18n

import (
	"testing"

	"gotest.tools/assert"
)

func TestNewI18N(t *testing.T) {
	locale := Locale{
		Lang: "de",
		Data: map[string]interface{}{
			"key": "value",
			"nested": map[string]interface{}{
				"value": "nested_value",
			},
			"data": map[string]interface{}{
				"singular":   "{var} name",
				"min_length": "{var} min length: {size}",
			},
			"reference": map[string]interface{}{
				"me":    "ref",
				"to_me": "@:reference.me",
			},
		},
	}

	t.Run("toJSON", func(t *testing.T) {
		j, err := locale.JSON()
		assert.NilError(t, err)
		// nolint
		assert.Equal(t, j, `{"data":{"min_length":"{var} min length: {size}","singular":"{var} name"},"key":"value","nested":{"value":"nested_value"},"reference":{"me":"ref","to_me":"@:reference.me"}}`)
	})

	t.Run("dot access", func(t *testing.T) {
		j := locale.Get("nested.value")
		assert.Equal(t, j, "nested_value")
	})

	t.Run("dot access of unknown", func(t *testing.T) {
		j := locale.Get("not.existing.original.returned")
		assert.Equal(t, j, "not.existing.original.returned")
	})

	t.Run("data replacement", func(t *testing.T) {
		j := locale.GetVar("data.singular", map[string]string{"var": "field"})
		assert.Equal(t, j, "field name")
	})

	t.Run("data with variable replacement", func(t *testing.T) {
		j := locale.GetVar("data.min_length", map[string]string{"var": "field", "size": "4"})
		assert.Equal(t, j, "field min length: 4")
	})

	t.Run("reference", func(t *testing.T) {
		j := locale.Get("reference.to_me")
		assert.Equal(t, j, "ref")
	})
}

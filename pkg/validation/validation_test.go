package validation

import (
	"testing"

	"go-webapp-example/pkg/i18n"

	"github.com/stretchr/testify/assert"
)

func TestNewErrorBag(t *testing.T) {
	err := NewErrorBag("subject")
	err.Add("field1", "error1")
	err.Add("field1", "error2")
	err.Add("field2", "error3")

	assert.True(t, err.Failed())

	for field, errs := range err.errors {
		if field == "field1" {
			assert.Lenf(t, errs, 2, "want 2 errors; got %d", len(errs))
			assert.Equalf(t, "validation.error1", errs[0].Message, "want error1; got %s", errs[0].Message)
			assert.Equalf(t, "validation.error2", errs[1].Message, "want error1; got %s", errs[1].Message)
		}
		if field == "field2" {
			assert.Lenf(t, errs, 1, "want 1 error; got %d", len(errs))
			assert.Equalf(t, "validation.error3", errs[0].Message, "want error3; got %s", errs[0].Message)
		}
	}
}

func TestTranslatedErrors(t *testing.T) {
	err := NewErrorBag("subject")
	err.Add("username", "required")
	err.AddData("password", "min_length", map[string]string{"size": "4"})

	locale := &i18n.Locale{
		Lang: "de",
		Data: map[string]interface{}{
			"validation": map[string]interface{}{
				"required":   "{field} wird benötigt",
				"min_length": "{field} {size}",
			},
			"subject": map[string]interface{}{
				"fields": map[string]interface{}{
					"username": "Benutzername",
					"password": "Passwort",
				},
			},
		},
	}

	for key, errs := range err.TranslatedErrors(locale) {
		if key == "username" {
			assert.Equal(t, "Benutzername wird benötigt", errs[0].Translated)
		}
		if key == "password" {
			assert.Equal(t, "Passwort 4", errs[0].Translated)
		}
	}
}

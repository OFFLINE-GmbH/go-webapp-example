package validation

import (
	"fmt"

	"go-webapp-example/pkg/i18n"

	"github.com/pkg/errors"
)

// ErrFailed is returned for failed validation attempts.
var ErrFailed = errors.New("validation failed")

// Errors is a map of field names to an error slice.
type Errors map[string][]Error

// ErrorBag contains error messages for a certain field.
type ErrorBag struct {
	errors  Errors
	subject string
}

// Error represents a single validation error.
type Error struct {
	Key        string            `json:"key"`
	Message    string            `json:"message"`
	Translated string            `json:"translated"`
	Data       map[string]string `json:"data"`
}

// NewErrorBag returns a new empty error bag.
func NewErrorBag(subject string) *ErrorBag {
	return &ErrorBag{
		subject: subject,
		errors:  make(Errors),
	}
}

// ConvertError turns a native error into a validation Errors struct.
func ConvertError(key string, err error) Errors {
	converted := Error{Message: err.Error(), Key: key}
	return Errors{key: []Error{converted}}
}

// NewFromString turns a string into a validation Errors struct.
func NewFromString(key, message string) Errors {
	converted := Error{Message: message, Key: key}
	return Errors{key: []Error{converted}}
}

// Add adds a new error to the ErrorBag.
func (e *ErrorBag) Add(key, message string) {
	e.AddData(key, message, map[string]string{})
}

// AddData adds a new error to the ErrorBag including meta data.
func (e *ErrorBag) AddData(key, message string, data map[string]string) {
	message = "validation." + message
	vErr := Error{
		Key:     key,
		Message: message,
		Data:    data,
	}
	_, ok := e.errors[key]
	if !ok {
		e.errors[key] = []Error{}
	}
	e.errors[key] = append(e.errors[key], vErr)
}

// Get gets all errors for a given key.
func (e *ErrorBag) Get(key string) []Error {
	err, ok := e.errors[key]
	if !ok {
		return []Error{}
	}
	return err
}

// Failed can be used to check if the validation failed (errors are present).
func (e *ErrorBag) Failed() bool {
	return len(e.errors) > 0
}

// Errors returns all error data.
func (e *ErrorBag) Errors() Errors {
	return e.errors
}

// TranslatedErrors returns all error messages in the specified locale.
func (e *ErrorBag) TranslatedErrors(locale *i18n.Locale) Errors {
	allErrs := e.Errors()
	translated := make(Errors, len(allErrs))
	for key, errs := range allErrs {
		// Create an empty Error slice for the current key
		translated[key] = []Error{}
		for _, err := range errs {
			// Translate the field name itself (subject.fields.key)
			translatedField := locale.Get(fmt.Sprintf("%s.fields.%s", e.subject, key))
			// Add the translated field name to the translation context
			data := map[string]string{"field": translatedField}
			// Merge in other error data
			for k, v := range err.Data {
				data[k] = v
			}
			// Translate the error message
			err.Translated = locale.GetVar(err.Message, data)
			// Re-attach the merged error data
			err.Data = data
			// Add the error to the Error slice
			translated[key] = append(translated[key], err)
		}
	}
	return translated
}

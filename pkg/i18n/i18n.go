package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"go-webapp-example/pkg/render"
	"go-webapp-example/pkg/router"

	"github.com/ghodss/yaml"
	"github.com/oleiade/reflections"
)

type ctxKeyType struct{ name string }

var ctxKey = ctxKeyType{"localeCtx"}

// ReferencePrefix used by vue-i18n to reference another locale key.
const ReferencePrefix = "@:"

// Middleware is used to attach the global locale instance to a request context.
func Middleware(locale *Locale) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxKey, locale)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CtxLocale returns the Locale from a context.
func CtxLocale(ctx context.Context) *Locale {
	return ctx.Value(ctxKey).(*Locale)
}

// Locale contains all translation information for a single locale.
type Locale struct {
	Lang string
	Data map[string]interface{}
}

// FromFiles returns a new Locale with data from the filesystem.
func FromFiles(dir, lang string) (*Locale, error) {
	locale := Locale{
		Lang: lang,
		Data: make(map[string]interface{}),
	}
	basePath := fmt.Sprintf("%s/%s/", dir, lang)
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return &locale, err
	}

	for _, file := range files {
		var target interface{}
		contents, err := ioutil.ReadFile(basePath + file.Name())
		if err != nil {
			return &locale, err
		}
		err = yaml.Unmarshal(contents, &target)
		if err != nil {
			return &locale, err
		}
		locale.Data[strings.ReplaceAll(file.Name(), ".yml", "")] = target
	}

	return &locale, nil
}

// Get returns a nested translation string.
func (l Locale) Get(key string) string {
	arr := strings.Split(key, ".")
	var obj interface{} = l.Data
	var err error
	for _, k := range arr {
		obj, err = getProperty(obj, k)
		if err != nil {
			return key
		}
		if obj == nil {
			return key
		}
	}
	value := fmt.Sprintf("%s", obj)
	if strings.HasPrefix(value, ReferencePrefix) {
		return l.Get(strings.Replace(value, ReferencePrefix, "", 1))
	}
	return value
}

// GetVar returns a translated string with variables in it.
func (l Locale) GetVar(key string, vars map[string]string) string {
	out := l.Get(key)
	// Replace named data variables.
	for k, value := range vars {
		out = strings.ReplaceAll(out, fmt.Sprintf("{%s}", k), value)
	}
	return out
}

// JSON returns a json representation of the locale.
func (l Locale) JSON() (string, error) {
	b, err := json.Marshal(l.Data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Handler returns a http handleFunc that returns the
// locale as a json representation.
func HandleFunc(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locale, err := FromFiles(dir, router.Param(r, "locale"))
		if err != nil {
			render.Error(w, err)
			return
		}

		j, err := locale.JSON()
		if err != nil {
			render.Error(w, err)
			return
		}

		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, j)
	}
}

// getProperty loops through a object to get a value via dot notation.
func getProperty(obj interface{}, prop string) (interface{}, error) {
	if reflect.TypeOf(obj).Kind() == reflect.Map {
		val := reflect.ValueOf(obj)
		valueOf := val.MapIndex(reflect.ValueOf(prop))
		if valueOf == reflect.Zero(reflect.ValueOf(prop).Type()) {
			return nil, nil
		}

		idx := val.MapIndex(reflect.ValueOf(prop))
		if !idx.IsValid() {
			return nil, nil
		}
		return idx.Interface(), nil
	}

	prop = strings.Title(prop)

	return reflections.GetField(obj, prop)
}

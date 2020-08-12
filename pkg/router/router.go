package router

import (
	"net/http"

	"github.com/go-chi/chi"
)

type Mux struct {
	mux *chi.Mux
}

func New() *Mux {
	return &Mux{chi.NewRouter()}
}

func NewFromMux(mux chi.Router) *Mux {
	return &Mux{mux: mux.(*chi.Mux)}
}

func (m *Mux) Handle(pattern string, handler http.Handler) {
	m.mux.Handle(pattern, handler)
}

func (m *Mux) Method(method, pattern string, handler http.Handler) {
	m.mux.Method(method, pattern, handler)
}

func (m *Mux) HandleFunc(pattern string, handleFunc http.HandlerFunc) {
	m.mux.HandleFunc(pattern, handleFunc)
}

func (m *Mux) UseMiddleware(middlewares ...func(http.Handler) http.Handler) {
	m.mux.Use(middlewares...)
}

func (m *Mux) Mount(pattern string, handler http.Handler) {
	m.mux.Mount(pattern, handler)
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.mux.ServeHTTP(w, req)
}

func (m *Mux) Group(f func(r *Mux)) {
	m.mux.Group(func(r chi.Router) {
		newMux := NewFromMux(r)
		f(newMux)
	})
}

func Param(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

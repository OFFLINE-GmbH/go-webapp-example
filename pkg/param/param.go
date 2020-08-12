package param

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

// Param returns the url parameter from a http.Request object.
func Param(r *http.Request, key string) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		fmt.Println(rctx.URLParams.Keys)
		return rctx.URLParam(key)
	}
	return ""
}

package render

import (
	"encoding/json"
	"io"
	"net/http"
)

const contentTypeJSON = "application/json; charset=utf-8"

// JSON encodes the given val using the standard json package and writes
// the encoding output to the given writer. If the writer implements the
// http.ResponseWriter interface, then this function will also set the
// proper JSON content-type header with charset as UTF-8. Status will be
// considered only when wr is http.ResponseWriter and in that case, status
// must be a valid status code.
func JSON(w io.Writer, status int, val interface{}) error {
	if hw, ok := w.(http.ResponseWriter); ok {
		hw.Header().Set("Content-type", contentTypeJSON)
		hw.WriteHeader(status)
	}

	return json.NewEncoder(w).Encode(val)
}

// Error returns an error response.
func Error(w http.ResponseWriter, err error) {
	type response struct {
		Success bool
		Error   string
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	_ = json.NewEncoder(w).Encode(response{
		Success: false,
		Error:   err.Error(),
	})
}

package handlers

import (
	"fmt"
	"net/http"
	"os"
)

// send error response to client
func writeErrorResponse(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)

	_, _ = fmt.Fprintf(os.Stderr, "error: %s", body)
	if _, err := w.Write([]byte(body)); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot write to stream: %v\n", err)
		return
	}
}

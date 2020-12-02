package handlers

import "net/http"

// create new configured router
func NewRouter(a Api) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/read_urls", concurrencyLimiter(a.cfg.MaxInRequests, a.ReadUrls))
	return mux
}

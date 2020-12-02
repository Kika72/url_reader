package handlers

import (
	"net/http"
)

func concurrencyLimiter(limit int, next http.HandlerFunc) http.HandlerFunc {
	var limitChan chan struct{}
	if limit > 0 {
		limitChan = make(chan struct{}, limit)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if limit > 0 {
			select {
			case limitChan <- struct{}{}:
				defer func() {
					<-limitChan
				}()
			default:
				writeErrorResponse(w, http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
				return
			}
		}

		next(w, r)
	}
}

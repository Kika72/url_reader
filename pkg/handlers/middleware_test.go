package handlers

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	he "github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"url_reader/pkg/service"
)

func Test_concurrencyLimiter(t *testing.T) {
	mux := http.NewServeMux()

	// prepare test handlers
	{
		mux.HandleFunc("/read_urls_25", concurrencyLimiter(3, func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(25 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		mux.HandleFunc("/read_urls_1000", concurrencyLimiter(3, func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1000 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
	}

	server := httptest.NewServer(mux)
	e := he.New(t, server.URL)

	// check if request counter is released
	for i := 0; i < 4; i++ {
		e.POST("/read_urls_25").
			WithJSON(service.UrlRequests{}).
			Expect().
			Status(http.StatusOK)
	}

	// check limit
	{
		wg := sync.WaitGroup{}
		wg.Add(4)
		counter := int64(4)
		for i := 0; i < 4; i++ {
			go func() {
				defer wg.Done()
				resp := e.POST("/read_urls_1000").
					WithJSON(service.UrlRequests{}).
					Expect().Raw()
				if resp.StatusCode != http.StatusOK {
					atomic.AddInt64(&counter, int64(-1))
				}
			}()
		}
		wg.Wait()
		// one request must fail
		require.Equal(t, int64(3), counter)
	}

}

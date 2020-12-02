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
	"url_reader/config"
	"url_reader/pkg/service"
)

func TestApi_ReadUrls(t *testing.T) {
	cfg := config.Config {
		Port:           0,
		MaxInRequests:  5,
		MaxUrls:        5,
		MaxOutRequests: 2,
		Timeout:        time.Second,
	}
	svc := service.NewUrlReader(cfg)
	api := NewApi(
		WithConfig(cfg),
		WithReader(svc),
	)
	mux := NewRouter(api)
	server := httptest.NewServer(mux)
	e := he.New(t, server.URL)

	e.POST("/read_urls").
		WithJSON(service.UrlRequests{
			Requests: []string{ "http://google.com" },
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Path("$.responses").Array().Length().Equal(1)

	// fail: urls count should be less or equal than Config.MaxUrls
	{
		req := service.UrlRequests{}
		for i := 0; i < 6; i++ {
			req.Requests = append(req.Requests, "http://google.com")
		}

		e.POST("/read_urls").
			WithJSON(req).
			Expect().
			Status(http.StatusBadRequest).
			Body().Equal("too many urls in request")
	}

	//success: urls count more than outgoing requests
	{
		req := service.UrlRequests{}
		for i := 0; i < cfg.MaxUrls; i++ {
			req.Requests = append(req.Requests, "http://google.com")
		}

		e.POST("/read_urls").
			WithJSON(req).
			Expect().
			Status(http.StatusOK).
			JSON().Object().Path("$.responses").Array().Length().Equal(cfg.MaxUrls)
	}

	handler := func (w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * cfg.Timeout)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	testSrv := httptest.NewServer(http.HandlerFunc(handler))
	defer testSrv.Close()
	// fail: time out
	{
		e.POST("/read_urls").
			WithJSON(service.UrlRequests{
				Requests: []string{ testSrv.URL },
			}).
			Expect().
			Status(http.StatusInternalServerError).
			Body().Contains("context deadline exceeded")
	}

	// fail: too many requests
	{
		req := service.UrlRequests{}
		for i := 0; i < cfg.MaxUrls; i++ {
			req.Requests = append(req.Requests, "http://google.com")
		}

		wg := sync.WaitGroup{}
		wg.Add(cfg.MaxInRequests + 1)
		counter := int64(cfg.MaxInRequests + 1)
		for i := 0; i < cfg.MaxInRequests + 1; i++ {
			go func() {
				defer wg.Done()
				resp := e.POST("/read_urls").
					WithJSON(req).
					Expect().Raw()
				if resp.StatusCode != http.StatusOK {
					atomic.AddInt64(&counter, int64(-1))
				}
			}()
		}
		wg.Wait()
		// one request must fail
		require.Equal(t, int64(cfg.MaxInRequests), counter)
	}
}

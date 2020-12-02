package handlers

import (
	"context"
	"time"

	"url_reader/pkg/service"
)

type urlReaderMock struct {
	workTime time.Duration
}

func (r urlReaderMock) Read(ctx context.Context, data service.UrlRequests) (service.UrlResponses, error) {
	time.Sleep(r.workTime)
	return service.UrlResponses{}, nil
}

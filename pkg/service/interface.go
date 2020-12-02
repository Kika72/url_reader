package service

import "context"

type UrlReader interface {
	Read(ctx context.Context, data UrlRequests) (UrlResponses, error)
}

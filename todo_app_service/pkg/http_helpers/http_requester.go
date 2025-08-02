package http_helpers

import (
	"context"
	"io"
	"net/http"
)

type requester struct{}

func NewHttpRequester() *requester {
	return &requester{}
}

func (*requester) NewRequestWithContext(ctx context.Context, httpMethod string, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, httpMethod, url, body)
}

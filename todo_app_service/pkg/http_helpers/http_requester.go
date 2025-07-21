package http_helpers

import (
	"context"
	"io"
	"net/http"
)

type httpRequester struct{}

func NewHttpRequester() *httpRequester {
	return &httpRequester{}
}

func (*httpRequester) NewRequestWithContext(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, url, body)
}

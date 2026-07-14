/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */
package uschess

import (
	"context"
	"net/http"
)

//go:generate go tool oapi-codegen -config oapi-codegen.yaml swagger.json

const (
	defaultAcceptHeader = "application/json"
	defaultAPIServer    = "https://ratings-api.uschess.org"
	defaultUserAgent    = "uschess-go/0.2.0 (+https://github.com/mikeb26/uschess-go)"
)

// NewDefaultClient creates a generated response-aware client for the US Chess
// ratings API. It retries eligible requests and requests JSON responses by
// default. Request editors passed in opts, or per request, can override the
// default Accept and User-Agent headers.
func NewDefaultClient(opts ...ClientOption) (*ClientWithResponses, error) {
	opts = append([]ClientOption{
		WithRequestEditorFn(defaultAcceptHeaderEditor),
		WithUserAgent(defaultUserAgent),
	}, opts...)
	client, err := NewClient(defaultAPIServer, opts...)
	if err != nil {
		return nil, err
	}
	client.Client = newRetryingClientFor(client.Client)
	return &ClientWithResponses{ClientInterface: client}, nil
}

// newRetryingClientFor preserves a caller-provided HTTP client while adding
// retry behavior. ClientOption permits custom HttpRequestDoers as well as
// *http.Client values, so custom doers are adapted to a RoundTripper first.
func newRetryingClientFor(doer HttpRequestDoer) *http.Client {
	if client, ok := doer.(*http.Client); ok {
		return newRetryingClient(client)
	}
	return newRetryingClient(&http.Client{Transport: doerRoundTripper{doer: doer}})
}

type doerRoundTripper struct {
	doer HttpRequestDoer
}

func (t doerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.doer.Do(req)
}

func defaultAcceptHeaderEditor(_ context.Context, req *http.Request) error {
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", defaultAcceptHeader)
	}
	return nil
}

// WithUserAgent sets the User-Agent header on requests made by the client.
// When used with NewDefaultClient, it overrides the default User-Agent.
func WithUserAgent(userAgent string) ClientOption {
	return WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", userAgent)
		return nil
	})
}

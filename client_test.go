/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type testDoer struct {
	req *http.Request
}

type testRoundTripper struct{}

func (testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func (d *testDoer) Do(req *http.Request) (*http.Response, error) {
	d.req = req
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestNewDefaultClientSetsJSONAcceptHeader(t *testing.T) {
	doer := &testDoer{}
	client, err := NewDefaultClient(WithHTTPClient(doer))
	if err != nil {
		t.Fatalf("NewDefaultClient returned an error: %v", err)
	}

	if _, err := client.GetMember(context.Background(), "12641216"); err != nil {
		t.Fatalf("GetMember returned an error: %v", err)
	}
	if got := doer.req.Header.Get("Accept"); got != defaultAcceptHeader {
		t.Errorf("Accept header = %q; want %q", got, defaultAcceptHeader)
	}
}

func TestNewDefaultClientAllowsAcceptHeaderOverride(t *testing.T) {
	doer := &testDoer{}
	client, err := NewDefaultClient(
		WithHTTPClient(doer),
		WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("Accept", "application/problem+json")
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("NewDefaultClient returned an error: %v", err)
	}

	if _, err := client.GetMember(context.Background(), "12641216"); err != nil {
		t.Fatalf("GetMember returned an error: %v", err)
	}
	if got := doer.req.Header.Get("Accept"); got != "application/problem+json" {
		t.Errorf("Accept header = %q; want caller override", got)
	}
}

func TestNewDefaultClientUsesRatingsAPIServerAndRetryingClient(t *testing.T) {
	client, err := NewDefaultClient()
	if err != nil {
		t.Fatalf("NewDefaultClient returned an error: %v", err)
	}

	generatedClient, ok := client.ClientInterface.(*Client)
	if !ok {
		t.Fatalf("ClientInterface type = %T; want *Client", client.ClientInterface)
	}
	if got, want := generatedClient.Server, defaultAPIServer+"/"; got != want {
		t.Errorf("server = %q; want %q", got, want)
	}

	httpClient, ok := generatedClient.Client.(*http.Client)
	if !ok {
		t.Fatalf("HTTP client type = %T; want *http.Client", generatedClient.Client)
	}
	if _, ok := httpClient.Transport.(*retryTransport); !ok {
		t.Errorf("HTTP transport type = %T; want *retryTransport", httpClient.Transport)
	}
}

func TestNewDefaultClientWrapsCallerHTTPClientWithRetries(t *testing.T) {
	callerTransport := &testRoundTripper{}
	callerClient := &http.Client{
		Transport: callerTransport,
		Timeout:   time.Minute,
	}

	client, err := NewDefaultClient(WithHTTPClient(callerClient))
	if err != nil {
		t.Fatalf("NewDefaultClient returned an error: %v", err)
	}

	generatedClient := client.ClientInterface.(*Client)
	httpClient, ok := generatedClient.Client.(*http.Client)
	if !ok {
		t.Fatalf("HTTP client type = %T; want *http.Client", generatedClient.Client)
	}
	if httpClient == callerClient {
		t.Fatal("NewDefaultClient used the caller's HTTP client directly")
	}
	if got, want := httpClient.Timeout, callerClient.Timeout; got != want {
		t.Errorf("timeout = %v; want %v", got, want)
	}
	retry, ok := httpClient.Transport.(*retryTransport)
	if !ok {
		t.Fatalf("HTTP transport type = %T; want *retryTransport", httpClient.Transport)
	}
	if retry.next != callerTransport {
		t.Errorf("retry transport wraps %T; want caller transport", retry.next)
	}
}

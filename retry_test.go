/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type testRetryTransport struct{}

func (*testRetryTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("not called")
}

func TestNewRetryingClientPreservesCallerConfiguration(t *testing.T) {
	transport := &testRetryTransport{}
	jar := &testCookieJar{}
	redirect := func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	base := &http.Client{
		Transport:     transport,
		Timeout:       time.Second,
		Jar:           jar,
		CheckRedirect: redirect,
	}

	client := newRetryingClient(base)

	if client == base {
		t.Fatal("newRetryingClient returned the caller's client directly")
	}
	if client.Timeout != base.Timeout || client.Jar != base.Jar {
		t.Fatal("newRetryingClient did not preserve the caller's client configuration")
	}
	if client.CheckRedirect == nil {
		t.Fatal("newRetryingClient did not preserve CheckRedirect")
	}
	retry, ok := client.Transport.(*retryTransport)
	if !ok {
		t.Fatalf("transport type = %T; want *retryTransport", client.Transport)
	}
	if retry.next != transport {
		t.Errorf("wrapped transport = %T; want caller transport", retry.next)
	}
}

func TestRetryTransportRetriesEligibleResponse(t *testing.T) {
	attempts := 0
	delays := make([]time.Duration, 0, 1)
	transport := &retryTransport{
		next: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts == 1 {
				return response(req, http.StatusServiceUnavailable), nil
			}
			return response(req, http.StatusOK), nil
		}),
		maxAttempts: 2,
		baseDelay:   time.Second,
		maxDelay:    time.Second,
		jitter:      func(delay time.Duration) time.Duration { return delay },
		sleep: func(_ context.Context, delay time.Duration) error {
			delays = append(delays, delay)
			return nil
		},
	}

	resp, err := transport.RoundTrip(request(t, http.MethodGet))
	if err != nil {
		t.Fatalf("RoundTrip returned an error: %v", err)
	}
	defer resp.Body.Close()
	if got, want := attempts, 2; got != want {
		t.Errorf("attempts = %d; want %d", got, want)
	}
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Errorf("status = %d; want %d", got, want)
	}
	if got, want := len(delays), 1; got != want {
		t.Fatalf("sleep calls = %d; want %d", got, want)
	}
	if got, want := delays[0], time.Second; got != want {
		t.Errorf("delay = %v; want %v", got, want)
	}
}

func TestRetryTransportDoesNotRetryUnsafeMethod(t *testing.T) {
	attempts := 0
	transport := &retryTransport{
		next: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			return response(req, http.StatusServiceUnavailable), nil
		}),
		maxAttempts: 4,
		sleep: func(context.Context, time.Duration) error {
			t.Fatal("sleep must not be called")
			return nil
		},
	}

	resp, err := transport.RoundTrip(request(t, http.MethodPost))
	if err != nil {
		t.Fatalf("RoundTrip returned an error: %v", err)
	}
	defer resp.Body.Close()
	if got, want := attempts, 1; got != want {
		t.Errorf("attempts = %d; want %d", got, want)
	}
}

func TestRetryDelayHonorsRetryAfter(t *testing.T) {
	resp := response(request(t, http.MethodGet), http.StatusTooManyRequests)
	resp.Header.Set("Retry-After", "12")

	got := retryDelay(resp, 1, time.Second, 30*time.Second, nil)
	if want := 12 * time.Second; got != want {
		t.Errorf("delay = %v; want %v", got, want)
	}
}

func request(t *testing.T, method string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, "https://example.test/", nil)
	if err != nil {
		t.Fatalf("NewRequest returned an error: %v", err)
	}
	return req
}

func response(req *http.Request, status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("response")),
		Request:    req,
	}
}

type testCookieJar struct{}

func (*testCookieJar) SetCookies(*url.URL, []*http.Cookie) {}

func (*testCookieJar) Cookies(*url.URL) []*http.Cookie { return nil }

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
	"sync"
	"testing"
)

type tournamentTestDoer struct {
	mu       sync.Mutex
	headers  []string
	requests []string
	failPath string
}

func (d *tournamentTestDoer) Do(req *http.Request) (*http.Response, error) {
	d.mu.Lock()
	d.requests = append(d.requests, req.URL.RequestURI())
	d.headers = append(d.headers, req.Header.Get("X-Tournament-Test"))
	d.mu.Unlock()

	status := http.StatusOK
	body := `{"id":"E123","name":"Spring Open","sections":[{"number":2,"name":"Reserve"},{"number":1,"name":"Open"}]}`
	switch req.URL.Path {
	case "/api/v1/rated-events/E123/sections/1/standings":
		if req.URL.Query().Get("Offset") == "100" {
			body = `{"items":[{"firstName":"Carol","ordinal":3}],"offset":100,"pageSize":100,"hasNextPage":false}`
		} else {
			body = `{"items":[{"firstName":"Bob","ordinal":2},{"firstName":"Alice","ordinal":1}],"offset":0,"pageSize":100,"hasNextPage":true}`
		}
	case "/api/v1/rated-events/E123/sections/2/standings":
		body = `{"items":[{"firstName":"Dan","ordinal":1}],"offset":0,"pageSize":100,"hasNextPage":false}`
	}
	if req.URL.Path == d.failPath {
		status = http.StatusInternalServerError
		body = `{"title":"server error"}`
	}

	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func (d *tournamentTestDoer) snapshot() (requests, headers []string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.requests...), append([]string(nil), d.headers...)
}

func TestGetTournament(t *testing.T) {
	doer := &tournamentTestDoer{}
	client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
	if err != nil {
		t.Fatalf("NewClientWithResponses returned an error: %v", err)
	}

	tournament, err := client.GetTournament(context.Background(), "E123", func(_ context.Context, req *http.Request) error {
		req.Header.Set("X-Tournament-Test", "present")
		return nil
	})
	if err != nil {
		t.Fatalf("GetTournament returned an error: %v", err)
	}
	if tournament.Id != "E123" || tournament.Name != "Spring Open" {
		t.Errorf("tournament detail = %+v; want event E123 Spring Open", tournament.RatedEventDetail)
	}
	if got, want := len(tournament.SectionStandings), len(tournament.Sections); got != want {
		t.Fatalf("section standings count = %d; want %d", got, want)
	}
	if got := []string{
		tournament.SectionStandings[0][0].FirstName,
		tournament.SectionStandings[1][0].FirstName,
		tournament.SectionStandings[1][1].FirstName,
		tournament.SectionStandings[1][2].FirstName,
	}; got[0] != "Dan" || got[1] != "Alice" || got[2] != "Bob" || got[3] != "Carol" {
		t.Errorf("section standings = %v; want standings aligned with sections and ordered by ordinal within each section", got)
	}

	requests, headers := doer.snapshot()
	if len(requests) != 4 {
		t.Fatalf("request count = %d; want 4: %v", len(requests), requests)
	}
	seen := make(map[string]bool, len(requests))
	for _, request := range requests {
		seen[request] = true
	}
	for _, want := range []string{
		"/api/v1/rated-events/E123",
		"/api/v1/rated-events/E123/sections/1/standings?Offset=0&Size=100",
		"/api/v1/rated-events/E123/sections/1/standings?Offset=100&Size=100",
		"/api/v1/rated-events/E123/sections/2/standings?Offset=0&Size=100",
	} {
		if !seen[want] {
			t.Errorf("requests = %v; missing %s", requests, want)
		}
	}
	for _, header := range headers {
		if header != "present" {
			t.Errorf("request editor header = %q; want present", header)
		}
	}
}

func TestGetTournamentReturnsStandingsError(t *testing.T) {
	doer := &tournamentTestDoer{failPath: "/api/v1/rated-events/E123/sections/2/standings"}
	client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
	if err != nil {
		t.Fatalf("NewClientWithResponses returned an error: %v", err)
	}

	tournament, err := client.GetTournament(context.Background(), "E123")
	if err == nil {
		t.Fatal("GetTournament returned nil error for a failed standings request")
	}
	if tournament != nil {
		t.Errorf("GetTournament tournament = %+v; want nil on error", tournament)
	}
	if !strings.Contains(err.Error(), "section 2") {
		t.Errorf("GetTournament error = %q; want section context", err)
	}
}

func TestGetTournamentReturnsUnexpectedEventResponse(t *testing.T) {
	doer := &tournamentTestDoer{failPath: "/api/v1/rated-events/E123"}
	client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
	if err != nil {
		t.Fatalf("NewClientWithResponses returned an error: %v", err)
	}

	tournament, err := client.GetTournament(context.Background(), "E123")
	if err == nil {
		t.Fatal("GetTournament returned nil error for an unexpected event response")
	}
	if tournament != nil {
		t.Errorf("GetTournament tournament = %+v; want nil on error", tournament)
	}
	requests, _ := doer.snapshot()
	if len(requests) != 1 {
		t.Errorf("request count = %d; want only event request after its failure", len(requests))
	}
}

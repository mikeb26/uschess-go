/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

type playerTestDoer struct {
	mu       sync.Mutex
	requests []string
	queries  []url.Values
}

func (d *playerTestDoer) Do(req *http.Request) (*http.Response, error) {
	d.mu.Lock()
	d.requests = append(d.requests, req.URL.Path)
	d.queries = append(d.queries, req.URL.Query())
	d.mu.Unlock()

	body := `{"id":"12345678","firstName":"Jane","lastName":"Player"}`
	switch {
	case strings.HasSuffix(req.URL.Path, "/rating-supplements"):
		body = `{"items":[{"ratingSupplementDate":"2026-01-01","ratings":[]}],"offset":0,"pageSize":100,"hasNextPage":false}`
	case strings.HasSuffix(req.URL.Path, "/games"):
		body = `{"items":[{}],"offset":0,"pageSize":100,"hasNextPage":false}`
	case strings.HasSuffix(req.URL.Path, "/sections"):
		body = `{"items":[{}],"offset":0,"pageSize":100,"hasNextPage":false}`
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func (d *playerTestDoer) paths() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.requests...)
}

func (d *playerTestDoer) queryForPath(path string) url.Values {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, requestPath := range d.requests {
		if requestPath == path {
			return d.queries[i]
		}
	}
	return nil
}

func TestGetPlayer(t *testing.T) {
	t.Run("without supplements", func(t *testing.T) {
		doer := &playerTestDoer{}
		client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
		if err != nil {
			t.Fatalf("NewClientWithResponses returned an error: %v", err)
		}

		player, err := client.GetPlayer(context.Background(), "12345678", false, nil, nil)
		if err != nil {
			t.Fatalf("GetPlayer returned an error: %v", err)
		}
		if player.FirstName != "Jane" || player.LastName != "Player" {
			t.Errorf("member detail = %+v; want Jane Player", player.MemberDetail)
		}
		if player.RatingSupplements != nil {
			t.Errorf("RatingSupplements = %v; want nil when not included", player.RatingSupplements)
		}
		if player.MemberRatedGames != nil {
			t.Errorf("MemberRatedGames = %v; want nil when no date is provided", player.MemberRatedGames)
		}
		if player.MemberRatedSections != nil {
			t.Errorf("MemberRatedSections = %v; want nil when no date is provided", player.MemberRatedSections)
		}
		if got := doer.paths(); len(got) != 1 || got[0] != "/api/v1/members/12345678" {
			t.Errorf("request paths = %v; want only member request", got)
		}
	})

	t.Run("with supplements", func(t *testing.T) {
		doer := &playerTestDoer{}
		client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
		if err != nil {
			t.Fatalf("NewClientWithResponses returned an error: %v", err)
		}

		player, err := client.GetPlayer(context.Background(), "12345678", true, nil, nil)
		if err != nil {
			t.Fatalf("GetPlayer returned an error: %v", err)
		}
		if player.FirstName != "Jane" || player.LastName != "Player" {
			t.Errorf("member detail = %+v; want Jane Player", player.MemberDetail)
		}
		if got := len(player.RatingSupplements); got != 1 {
			t.Errorf("RatingSupplements length = %d; want 1", got)
		}

		paths := doer.paths()
		if len(paths) != 2 {
			t.Fatalf("request paths = %v; want two requests", paths)
		}
		seen := make(map[string]bool, len(paths))
		for _, path := range paths {
			seen[path] = true
		}
		if !seen["/api/v1/members/12345678"] || !seen["/api/v1/members/12345678/rating-supplements"] {
			t.Errorf("request paths = %v; want member and supplements requests", paths)
		}
	})

	t.Run("with recent games", func(t *testing.T) {
		doer := &playerTestDoer{}
		client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
		if err != nil {
			t.Fatalf("NewClientWithResponses returned an error: %v", err)
		}

		onOrAfter := time.Date(2026, time.January, 15, 13, 45, 0, 0, time.FixedZone("UTC-5", -5*60*60))
		player, err := client.GetPlayer(context.Background(), "12345678", false, &onOrAfter, nil)
		if err != nil {
			t.Fatalf("GetPlayer returned an error: %v", err)
		}
		if got := len(player.MemberRatedGames); got != 1 {
			t.Errorf("MemberRatedGames length = %d; want 1", got)
		}
		if player.RatingSupplements != nil {
			t.Errorf("RatingSupplements = %v; want nil when not included", player.RatingSupplements)
		}

		paths := doer.paths()
		if len(paths) != 2 {
			t.Fatalf("request paths = %v; want two requests", paths)
		}
		seen := make(map[string]bool, len(paths))
		for _, path := range paths {
			seen[path] = true
		}
		if !seen["/api/v1/members/12345678"] || !seen["/api/v1/members/12345678/games"] {
			t.Errorf("request paths = %v; want member and games requests", paths)
		}
		if got := doer.queryForPath("/api/v1/members/12345678/games").Get("OnOrAfterDate"); got != "2026-01-15" {
			t.Errorf("OnOrAfterDate = %q; want %q", got, "2026-01-15")
		}
	})

	t.Run("with recent sections", func(t *testing.T) {
		doer := &playerTestDoer{}
		client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
		if err != nil {
			t.Fatalf("NewClientWithResponses returned an error: %v", err)
		}

		onOrAfter := time.Date(2026, time.February, 3, 16, 30, 0, 0, time.FixedZone("UTC+8", 8*60*60))
		player, err := client.GetPlayer(context.Background(), "12345678", false, nil, &onOrAfter)
		if err != nil {
			t.Fatalf("GetPlayer returned an error: %v", err)
		}
		if got := len(player.MemberRatedSections); got != 1 {
			t.Errorf("MemberRatedSections length = %d; want 1", got)
		}
		if player.MemberRatedGames != nil {
			t.Errorf("MemberRatedGames = %v; want nil when no date is provided", player.MemberRatedGames)
		}

		paths := doer.paths()
		if len(paths) != 2 {
			t.Fatalf("request paths = %v; want two requests", paths)
		}
		seen := make(map[string]bool, len(paths))
		for _, path := range paths {
			seen[path] = true
		}
		if !seen["/api/v1/members/12345678"] || !seen["/api/v1/members/12345678/sections"] {
			t.Errorf("request paths = %v; want member and sections requests", paths)
		}
		if got := doer.queryForPath("/api/v1/members/12345678/sections").Get("OnOrAfterDate"); got != "2026-02-03" {
			t.Errorf("OnOrAfterDate = %q; want %q", got, "2026-02-03")
		}
	})
}

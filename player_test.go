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

type playerTestDoer struct {
	mu       sync.Mutex
	requests []string
}

func (d *playerTestDoer) Do(req *http.Request) (*http.Response, error) {
	d.mu.Lock()
	d.requests = append(d.requests, req.URL.Path)
	d.mu.Unlock()

	body := `{"id":"12345678","firstName":"Jane","lastName":"Player"}`
	if strings.HasSuffix(req.URL.Path, "/rating-supplements") {
		body = `{"items":[{"ratingSupplementDate":"2026-01-01","ratings":[]}],"offset":0,"pageSize":100,"hasNextPage":false}`
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

func TestGetPlayer(t *testing.T) {
	t.Run("without supplements", func(t *testing.T) {
		doer := &playerTestDoer{}
		client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
		if err != nil {
			t.Fatalf("NewClientWithResponses returned an error: %v", err)
		}

		player, err := client.GetPlayer(context.Background(), "12345678", false)
		if err != nil {
			t.Fatalf("GetPlayer returned an error: %v", err)
		}
		if player.FirstName != "Jane" || player.LastName != "Player" {
			t.Errorf("member detail = %+v; want Jane Player", player.MemberDetail)
		}
		if player.RatingSupplements != nil {
			t.Errorf("RatingSupplements = %v; want nil when not included", player.RatingSupplements)
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

		player, err := client.GetPlayer(context.Background(), "12345678", true)
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
}

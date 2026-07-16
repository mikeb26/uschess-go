/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type playerTestDoer struct {
	mu           sync.Mutex
	requests     []string
	queries      []url.Values
	sectionsBody string
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
	case strings.HasSuffix(req.URL.Path, "/events"):
		body = `{"items":[{"name":"Event","startDate":"2026-02-03"}],"offset":0,"pageSize":100,"hasNextPage":false}`
	case strings.HasSuffix(req.URL.Path, "/games"):
		body = `{"items":[{}],"offset":0,"pageSize":100,"hasNextPage":false}`
	case strings.HasSuffix(req.URL.Path, "/sections"):
		if d.sectionsBody != "" {
			body = d.sectionsBody
		} else {
			body = `{"items":[{"endDate":"2026-02-03"}],"offset":0,"pageSize":100,"hasNextPage":false}`
		}
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
	t.Run("uses default options when nil", func(t *testing.T) {
		doer := &playerTestDoer{}
		client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
		if err != nil {
			t.Fatalf("NewClientWithResponses returned an error: %v", err)
		}

		player, err := client.GetPlayer(context.Background(), "12345678", nil)
		if err != nil {
			t.Fatalf("GetPlayer returned an error: %v", err)
		}
		if player.FirstName != "Jane" || player.LastName != "Player" {
			t.Errorf("member detail = %+v; want Jane Player", player.MemberDetail)
		}
		if len(player.RatingSupplements) != 1 {
			t.Errorf("RatingSupplements = %v; want one supplement", len(player.RatingSupplements))
		}
		if len(player.MemberEvents) != 1 {
			t.Errorf("MemberEvents = %v; want one event", len(player.MemberEvents))
		}
		if len(player.MemberRatedGames) != 1 {
			t.Errorf("MemberRatedGames = %v; want one recent game", len(player.MemberRatedGames))
		}
		if len(player.MemberRatedSections) != 1 {
			t.Errorf("MemberRatedSections = %v; want one recent section", len(player.MemberRatedSections))
		}
		if !player.liveIncluded {
			t.Error("liveIncluded = false; want true from default options")
		}

		paths := doer.paths()
		if len(paths) != 5 {
			t.Fatalf("request paths = %v; want member, supplements, events, games, and sections requests", paths)
		}
		seen := make(map[string]bool, len(paths))
		for _, path := range paths {
			seen[path] = true
		}
		for _, path := range []string{
			"/api/v1/members/12345678",
			"/api/v1/members/12345678/rating-supplements",
			"/api/v1/members/12345678/events",
			"/api/v1/members/12345678/games",
			"/api/v1/members/12345678/sections",
		} {
			if !seen[path] {
				t.Errorf("request paths = %v; want %s", paths, path)
			}
		}
	})

	t.Run("with supplements", func(t *testing.T) {
		doer := &playerTestDoer{}
		client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
		if err != nil {
			t.Fatalf("NewClientWithResponses returned an error: %v", err)
		}

		player, err := client.GetPlayer(context.Background(), "12345678", &GetPlayerOptions{IncludeSupplements: true})
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

	t.Run("with events", func(t *testing.T) {
		doer := &playerTestDoer{}
		client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
		if err != nil {
			t.Fatalf("NewClientWithResponses returned an error: %v", err)
		}

		player, err := client.GetPlayer(context.Background(), "12345678", &GetPlayerOptions{IncludeEvents: true})
		if err != nil {
			t.Fatalf("GetPlayer returned an error: %v", err)
		}
		if got := len(player.MemberEvents); got != 1 {
			t.Errorf("MemberEvents length = %d; want 1", got)
		}

		paths := doer.paths()
		if len(paths) != 2 {
			t.Fatalf("request paths = %v; want member and events requests", paths)
		}
		seen := make(map[string]bool, len(paths))
		for _, path := range paths {
			seen[path] = true
		}
		if !seen["/api/v1/members/12345678"] || !seen["/api/v1/members/12345678/events"] {
			t.Errorf("request paths = %v; want member and events requests", paths)
		}
	})

	t.Run("with recent games", func(t *testing.T) {
		doer := &playerTestDoer{}
		client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
		if err != nil {
			t.Fatalf("NewClientWithResponses returned an error: %v", err)
		}

		onOrAfter := time.Date(2026, time.January, 15, 13, 45, 0, 0, time.FixedZone("UTC-5", -5*60*60))
		player, err := client.GetPlayer(context.Background(), "12345678", &GetPlayerOptions{RecentGamesOnOrAfter: &onOrAfter})
		if err != nil {
			t.Fatalf("GetPlayer returned an error: %v", err)
		}
		if got := len(player.MemberRatedGames); got != 1 {
			t.Errorf("MemberRatedGames length = %d; want 1", got)
		}
		if len(player.RatingSupplements) != 0 {
			t.Errorf("RatingSupplements = %v; want empty when not included", len(player.RatingSupplements))
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
		query := doer.queryForPath("/api/v1/members/12345678/games")
		for _, key := range []string{"OnOrBeforeDate", "RatingSource", "OpponentId", "PreRating", "PostRating"} {
			if _, ok := query[key]; ok {
				t.Errorf("%s = %q; want omitted", key, query.Get(key))
			}
		}
	})

	t.Run("with recent sections", func(t *testing.T) {
		doer := &playerTestDoer{}
		client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
		if err != nil {
			t.Fatalf("NewClientWithResponses returned an error: %v", err)
		}

		onOrAfter := time.Date(2026, time.February, 3, 16, 30, 0, 0, time.FixedZone("UTC+8", 8*60*60))
		player, err := client.GetPlayer(context.Background(), "12345678", &GetPlayerOptions{RecentSectionsOnOrAfter: &onOrAfter})
		if err != nil {
			t.Fatalf("GetPlayer returned an error: %v", err)
		}
		if got := len(player.MemberRatedSections); got != 1 {
			t.Errorf("MemberRatedSections length = %d; want 1", got)
		}
		if len(player.MemberRatedGames) != 0 {
			t.Errorf("MemberRatedGames = %v; want empty when no date is provided", len(player.MemberRatedGames))
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
		query := doer.queryForPath("/api/v1/members/12345678/sections")
		for _, key := range []string{"OnOrBeforeDate", "RatingSource"} {
			if _, ok := query[key]; ok {
				t.Errorf("%s = %q; want omitted", key, query.Get(key))
			}
		}
	})

	t.Run("with live ratings and recent sections uses one sections request", func(t *testing.T) {
		cutoff := mostRecentMonthlyRatingCutoff(time.Now())
		dayAfterCutoff := cutoff.AddDate(0, 0, 1)
		recentSectionsDate := cutoff.AddDate(0, 0, 2)
		doer := &playerTestDoer{sectionsBody: fmt.Sprintf(`{"items":[
			{"endDate":%q,"ratingRecords":[{"postRating":1500}]},
			{"endDate":%q,"ratingRecords":[{"postRating":1600}]},
			{"endDate":%q,"ratingRecords":[{"postRating":1700}]}
		],"offset":0,"pageSize":100,"hasNextPage":false}`,
			cutoff.Format(time.DateOnly), dayAfterCutoff.Format(time.DateOnly), recentSectionsDate.Format(time.DateOnly))}
		client, err := NewClientWithResponses("https://example.test", WithHTTPClient(doer))
		if err != nil {
			t.Fatalf("NewClientWithResponses returned an error: %v", err)
		}

		player, err := client.GetPlayer(context.Background(), "12345678", &GetPlayerOptions{
			IncludeLiveRatings:      true,
			RecentSectionsOnOrAfter: &recentSectionsDate,
		})
		if err != nil {
			t.Fatalf("GetPlayer returned an error: %v", err)
		}
		if got := len(player.MemberRatedSections); got != 1 {
			t.Errorf("MemberRatedSections length = %d; want 1 after trimming to requested date", got)
		}
		if got := len(player.postSupplementRatingRecords); got != 2 {
			t.Errorf("LiveRatings length = %d; want 2 sections ending after cutoff", got)
		}

		const sectionsPath = "/api/v1/members/12345678/sections"
		var sectionRequests int
		for _, path := range doer.paths() {
			if path == sectionsPath {
				sectionRequests++
			}
		}
		if sectionRequests != 1 {
			t.Errorf("sections requests = %d; want 1", sectionRequests)
		}
		if got := doer.queryForPath(sectionsPath).Get("OnOrAfterDate"); got != cutoff.Format(time.DateOnly) {
			t.Errorf("OnOrAfterDate = %q; want earliest date %q", got, cutoff.Format(time.DateOnly))
		}
	})
}

func TestDefaultGetPlayerOptions(t *testing.T) {
	before := time.Now().AddDate(-1, 0, 0)
	first := DefaultGetPlayerOptions()
	second := DefaultGetPlayerOptions()
	after := time.Now().AddDate(-1, 0, 0)

	if !first.IncludeLiveRatings {
		t.Error("DefaultGetPlayerOptions().IncludeLiveRatings = false; want true")
	}
	if !first.IncludeEvents {
		t.Error("DefaultGetPlayerOptions().IncludeEvents = false; want true")
	}
	for name, date := range map[string]*time.Time{
		"RecentGamesOnOrAfter":    first.RecentGamesOnOrAfter,
		"RecentSectionsOnOrAfter": first.RecentSectionsOnOrAfter,
	} {
		if date == nil {
			t.Errorf("DefaultGetPlayerOptions().%s = nil; want one year ago", name)
			continue
		}
		if date.Before(before) || date.After(after) {
			t.Errorf("DefaultGetPlayerOptions().%s = %s; want approximately one year ago", name, *date)
		}
	}
	first.IncludeLiveRatings = false
	if !second.IncludeLiveRatings {
		t.Error("DefaultGetPlayerOptions returned shared options")
	}
	*first.RecentGamesOnOrAfter = time.Time{}
	if second.RecentGamesOnOrAfter.IsZero() {
		t.Error("DefaultGetPlayerOptions returned shared date pointers")
	}
	if first.RecentSectionsOnOrAfter.IsZero() {
		t.Error("DefaultGetPlayerOptions shares its games and sections date pointers")
	}
}

func TestGetLiveRatingRecords(t *testing.T) {
	cutoff := time.Date(2026, time.March, 18, 23, 45, 0, 0, time.FixedZone("CDT", -5*60*60))
	sections := []MemberRatedSection{
		{
			EndDate:       openapi_types.Date{Time: cutoff},
			RatingRecords: []MinimalRatingRecord{{PostRating: 1400}},
		},
		{
			EndDate:       openapi_types.Date{Time: cutoff.AddDate(0, 0, 1)},
			RatingRecords: []MinimalRatingRecord{{PostRating: 1500}, {PostRating: 1600}},
		},
		{
			EndDate:       openapi_types.Date{Time: cutoff.AddDate(0, 0, 2)},
			RatingRecords: []MinimalRatingRecord{{PostRating: 1700}},
		},
	}

	records := getLiveRatingRecords(sections, cutoff)
	if got := len(records); got != 3 {
		t.Fatalf("getLiveRatingRecords returned %d records; want 3", got)
	}
	if records[0].PostRating != 1700 || records[1].PostRating != 1500 || records[2].PostRating != 1600 {
		t.Errorf("getLiveRatingRecords = %+v; want records ordered by section end date descending", records)
	}
}

func TestPlayerLiveRating(t *testing.T) {
	t.Run("returns an error when live ratings were not included", func(t *testing.T) {
		player := &Player{}

		if _, err := player.LiveRatings(); err == nil {
			t.Error("LiveRating returned nil error; want an error")
		}
	})

	t.Run("filters unlisted ratings and uses the most recent record per type", func(t *testing.T) {
		player := &Player{
			liveIncluded: true,
			latestSupplement: RatingSupplement{Ratings: []RatingSupplementSystem{
				{RatingType: RatingTypeR, Rating: 1400, ProvisionalGameCount: 10},
				{RatingType: RatingTypeQ, Rating: 0},
				{RatingType: RatingTypeB, Rating: 1200, ProvisionalGameCount: 5},
			}},
			// Records are ordered from most to least recent.
			postSupplementRatingRecords: []MinimalRatingRecord{
				{RatingType: RatingTypeR, PostRating: 1550, PostProvisionalGameCount: 12},
				{RatingType: RatingTypeB, PostRating: 1300, PostProvisionalGameCount: 6},
				{RatingType: RatingTypeR, PostRating: 1500, PostProvisionalGameCount: 11},
			},
		}

		ratings, err := player.LiveRatings()
		if err != nil {
			t.Fatalf("LiveRating returned an error: %v", err)
		}
		want := []RatingSupplementSystem{
			{RatingType: RatingTypeR, Rating: 1550, ProvisionalGameCount: 12},
			{RatingType: RatingTypeB, Rating: 1300, ProvisionalGameCount: 6},
		}
		if len(ratings) != len(want) {
			t.Fatalf("LiveRating returned %d ratings; want %d: %+v", len(ratings), len(want), ratings)
		}
		for i := range want {
			if ratings[i] != want[i] {
				t.Errorf("LiveRatings()[%d] = %+v; want %+v", i, ratings[i], want[i])
			}
		}
	})
}

func TestMostRecentMonthlyRatingCutoff(t *testing.T) {
	central, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatalf("LoadLocation returned an error: %v", err)
	}
	previousCutoff := time.Date(2025, time.September, 17, 23, 45, 0, 0, central)
	currentCutoff := time.Date(2025, time.October, 15, 23, 45, 0, 0, central)

	for _, tt := range []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{"before cutoff", currentCutoff.Add(-time.Minute), previousCutoff},
		{"at cutoff", currentCutoff, currentCutoff},
		{"after cutoff", currentCutoff.Add(time.Minute), currentCutoff},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := mostRecentMonthlyRatingCutoff(tt.now); !got.Equal(tt.want) {
				t.Errorf("mostRecentMonthlyRatingCutoff(%s) = %s; want %s", tt.now, got, tt.want)
			}
		})
	}
}

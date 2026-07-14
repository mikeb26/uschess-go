/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestCollectPages(t *testing.T) {
	var offsets []int32
	items, err := collectPages(context.Background(), func(_ context.Context, offset int32) (pageResult[int], error) {
		offsets = append(offsets, offset)
		switch offset {
		case 0:
			return pageResult[int]{items: []int{1, 2}, offset: 0, pageSize: 2, hasNextPage: true}, nil
		case 2:
			return pageResult[int]{items: []int{3}, offset: 2, pageSize: 2}, nil
		default:
			t.Fatalf("unexpected offset %d", offset)
			return pageResult[int]{}, nil
		}
	})
	if err != nil {
		t.Fatalf("collectPages returned an error: %v", err)
	}
	if len(items) != 3 || items[0] != 1 || items[1] != 2 || items[2] != 3 {
		t.Fatalf("collectPages returned %v; want [1 2 3]", items)
	}
	if len(offsets) != 2 || offsets[0] != 0 || offsets[1] != 2 {
		t.Fatalf("fetch offsets = %v; want [0 2]", offsets)
	}
}

func TestCollectPagesRejectsNonAdvancingPage(t *testing.T) {
	_, err := collectPages(context.Background(), func(_ context.Context, _ int32) (pageResult[int], error) {
		return pageResult[int]{offset: 0, pageSize: 0, hasNextPage: true}, nil
	})
	if err == nil {
		t.Fatal("collectPages returned nil error for a non-advancing page")
	}
}

func TestGetAllWrappersUseAPISideSorting(t *testing.T) {
	client, err := NewClientWithResponses("https://example.test")
	if err != nil {
		t.Fatalf("NewClientWithResponses returned an error: %v", err)
	}

	stopRequest := errors.New("stop request")
	tests := []struct {
		name   string
		sortBy string
		dir    string
		call   func(RequestEditorFn) error
	}{
		{
			name:   "affiliates",
			sortBy: "Name",
			dir:    "Asc",
			call: func(editor RequestEditorFn) error {
				_, err := client.GetAllAffiliates(context.Background(), editor)
				return err
			},
		},
		{
			name:   "affiliate rated events",
			sortBy: "StartDate",
			dir:    "Desc",
			call: func(editor RequestEditorFn) error {
				_, err := client.GetAllAffiliateRatedEvents(context.Background(), "A123", editor)
				return err
			},
		},
		{
			name:   "Grand Prix standings",
			sortBy: "PointsThisYear",
			dir:    "Desc",
			call: func(editor RequestEditorFn) error {
				_, err := client.GetAllGrandPrixStandings(context.Background(), editor)
				return err
			},
		},
		{
			name:   "Grand Prix sections",
			sortBy: "RatedDate",
			dir:    "Desc",
			call: func(editor RequestEditorFn) error {
				_, err := client.GetAllGrandPrixSections(context.Background(), 2026, editor)
				return err
			},
		},
		{
			name:   "members",
			sortBy: "Name",
			dir:    "Asc",
			call: func(editor RequestEditorFn) error {
				_, err := client.GetAllMembers(context.Background(), editor)
				return err
			},
		},
		{
			name:   "pending events",
			sortBy: "UpdatedOn",
			dir:    "Desc",
			call: func(editor RequestEditorFn) error {
				_, err := client.GetAllPendingEvents(context.Background(), editor)
				return err
			},
		},
		{
			name:   "pending players",
			sortBy: "PairingNumber",
			dir:    "Asc",
			call: func(editor RequestEditorFn) error {
				_, err := client.GetAllPendingPlayers(context.Background(), "E123", "S123", editor)
				return err
			},
		},
		{
			name:   "rated events",
			sortBy: "StartDate",
			dir:    "Desc",
			call: func(editor RequestEditorFn) error {
				_, err := client.GetAllRatedEvents(context.Background(), editor)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var query url.Values
			err := tt.call(func(_ context.Context, req *http.Request) error {
				query = req.URL.Query()
				return stopRequest
			})
			if !errors.Is(err, stopRequest) {
				t.Fatalf("wrapper error = %v; want request editor error", err)
			}
			if got := query.Get("SortBy"); got != tt.sortBy {
				t.Errorf("SortBy = %q; want %q", got, tt.sortBy)
			}
			if got := query.Get("Dir"); got != tt.dir {
				t.Errorf("Dir = %q; want %q", got, tt.dir)
			}
		})
	}
}

func TestLocalSortingContracts(t *testing.T) {
	t.Run("rating supplements", func(t *testing.T) {
		items := []RatingSupplement{
			{RatingSupplementDate: testDate(2024, time.January, 1)},
			{RatingSupplementDate: testDate(2026, time.January, 1)},
			{RatingSupplementDate: testDate(2025, time.January, 1)},
		}
		sortRatingSupplementsByDateDescending(items)
		if !items[0].RatingSupplementDate.After(items[1].RatingSupplementDate.Time) || !items[1].RatingSupplementDate.After(items[2].RatingSupplementDate.Time) {
			t.Errorf("rating supplements are not sorted by date descending: %v", items)
		}
	})

	t.Run("member rated events", func(t *testing.T) {
		items := []RatedEvent{
			{Name: "old", StartDate: testDate(2024, time.January, 1)},
			{Name: "new", StartDate: testDate(2026, time.January, 1)},
			{Name: "middle", StartDate: testDate(2025, time.January, 1)},
		}
		sortRatedEventsByStartDateDescending(items)
		if got := []string{items[0].Name, items[1].Name, items[2].Name}; got[0] != "new" || got[1] != "middle" || got[2] != "old" {
			t.Errorf("rated event order = %v; want [new middle old]", got)
		}
	})

	t.Run("member rated games", func(t *testing.T) {
		items := []MemberRatedGame{
			{Opponent: MemberOpponentGamePlayer{FirstName: "old"}, Event: MinimalRatedEvent{StartDate: testDate(2024, time.January, 1)}},
			{Opponent: MemberOpponentGamePlayer{FirstName: "new"}, Event: MinimalRatedEvent{StartDate: testDate(2026, time.January, 1)}},
			{Opponent: MemberOpponentGamePlayer{FirstName: "middle"}, Event: MinimalRatedEvent{StartDate: testDate(2025, time.January, 1)}},
		}
		sortMemberRatedGamesByEventStartDateDescending(items)
		if got := []string{items[0].Opponent.FirstName, items[1].Opponent.FirstName, items[2].Opponent.FirstName}; got[0] != "new" || got[1] != "middle" || got[2] != "old" {
			t.Errorf("member rated game order = %v; want [new middle old]", got)
		}
	})

	t.Run("member rated sections", func(t *testing.T) {
		items := []MemberRatedSection{
			{SectionName: "old", StartDate: testDate(2024, time.January, 1)},
			{SectionName: "new", StartDate: testDate(2026, time.January, 1)},
			{SectionName: "middle", StartDate: testDate(2025, time.January, 1)},
		}
		sortMemberRatedSectionsByStartDateDescending(items)
		if got := []string{items[0].SectionName, items[1].SectionName, items[2].SectionName}; got[0] != "new" || got[1] != "middle" || got[2] != "old" {
			t.Errorf("member rated section order = %v; want [new middle old]", got)
		}
	})

	t.Run("rated sections", func(t *testing.T) {
		items := []RatedSection{
			{SectionName: "old", RatedDate: testDate(2024, time.January, 1)},
			{SectionName: "new", RatedDate: testDate(2026, time.January, 1)},
			{SectionName: "middle", RatedDate: testDate(2025, time.January, 1)},
		}
		sortRatedSectionsByRatedDateDescending(items)
		if got := []string{items[0].SectionName, items[1].SectionName, items[2].SectionName}; got[0] != "new" || got[1] != "middle" || got[2] != "old" {
			t.Errorf("rated section order = %v; want [new middle old]", got)
		}
	})

	t.Run("rated event standings", func(t *testing.T) {
		items := []Standings{{Ordinal: 3}, {Ordinal: 1}, {Ordinal: 2}}
		sortStandingsByOrdinalAscending(items)
		if got := []int32{items[0].Ordinal, items[1].Ordinal, items[2].Ordinal}; got[0] != 1 || got[1] != 2 || got[2] != 3 {
			t.Errorf("standing order = %v; want [1 2 3]", got)
		}
	})
}

func testDate(year int, month time.Month, day int) openapi_types.Date {
	return openapi_types.Date{Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

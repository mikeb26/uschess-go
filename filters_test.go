/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import (
	"net/http"
	"testing"
)

func TestOptionalFiltersAreOmitted(t *testing.T) {
	tests := []struct {
		name  string
		keys  []string
		build func() (*http.Request, error)
	}{
		{
			name: "affiliates",
			keys: []string{"Fuzzy", "StateCode"},
			build: func() (*http.Request, error) {
				return NewGetAffiliatesPageRequest("https://example.test", &GetAffiliatesPageParams{})
			},
		},
		{
			name: "affiliate rated events",
			keys: []string{"Name", "FromDate", "ToDate", "StateCode", "City", "Affiliate", "ScholasticCode", "Women", "MinSize", "RatingSource", "DomesticStatus"},
			build: func() (*http.Request, error) {
				return NewGetAffiliateRatedEventsRequest("https://example.test", "A123", &GetAffiliateRatedEventsParams{})
			},
		},
		{
			name: "grand prix sections",
			keys: []string{"Search", "StateCode", "IsWomen"},
			build: func() (*http.Request, error) {
				return NewGetGrandPrixSectionsRequest("https://example.test", 2026, &GetGrandPrixSectionsParams{})
			},
		},
		{
			name: "grand prix standings",
			keys: []string{"Search", "StateCode", "IsWomen"},
			build: func() (*http.Request, error) {
				return NewGetGrandPrixStandingsRequest("https://example.test", &GetGrandPrixStandingsParams{})
			},
		},
		{
			name: "members",
			keys: []string{"Fuzzy", "RatingSource", "StateRep", "Jurisdiction", "Gender", "Fide", "Domestic", "MinRating", "MaxRating", "Ranked", "Status", "ExpireStartDate", "ExpireEndDate", "UsePeak", "RatingCutoffFrom", "RatingCutoffTo", "UseUnofficialRatings", "MinAge", "MaxAge", "AgeDate"},
			build: func() (*http.Request, error) {
				return NewGetMembersPageRequest("https://example.test", &GetMembersPageParams{})
			},
		},
		{
			name: "unofficial rank lookup",
			keys: []string{"jurisdiction"},
			build: func() (*http.Request, error) {
				return NewGetUnofficialRankLookupRequest("https://example.test", RatingTypeR, &GetUnofficialRankLookupParams{})
			},
		},
		{
			name: "member rated games",
			keys: []string{"OnOrAfterDate", "OnOrBeforeDate", "RatingSource", "OpponentId", "PreRating", "PostRating"},
			build: func() (*http.Request, error) {
				return NewGetMemberRatedGamesRequest("https://example.test", "12345678", &GetMemberRatedGamesParams{})
			},
		},
		{
			name: "member rated sections",
			keys: []string{"OnOrAfterDate", "OnOrBeforeDate", "RatingSource"},
			build: func() (*http.Request, error) {
				return NewGetMemberRatedSectionsPageRequest("https://example.test", "12345678", &GetMemberRatedSectionsPageParams{})
			},
		},
		{
			name: "pending events",
			keys: []string{"Name", "FromDate", "ToDate", "StateCode", "City", "ScholasticCode", "Women", "GrandPrix", "MinSize", "TimeControl", "RatingSource", "Status", "ReviewStatus", "DomesticStatus", "OwnerId", "AffiliateId"},
			build: func() (*http.Request, error) {
				return NewGetPendingEventsPageRequest("https://example.test", &GetPendingEventsPageParams{})
			},
		},
		{
			name: "pending players",
			keys: []string{"Fuzzy", "State", "WithValidation"},
			build: func() (*http.Request, error) {
				return NewListPendingPlayersRequest("https://example.test", "E123", "S123", &ListPendingPlayersParams{})
			},
		},
		{
			name: "round unpaired players",
			keys: []string{"Fuzzy", "State", "WithValidation"},
			build: func() (*http.Request, error) {
				return NewListRoundUnpairedPlayersRequest("https://example.test", "E123", "S123", 1, &ListRoundUnpairedPlayersParams{})
			},
		},
		{
			name: "unplayed rounds",
			keys: []string{"roundNumber"},
			build: func() (*http.Request, error) {
				return NewListUnplayedRoundsRequest("https://example.test", "E123", "S123", &ListUnplayedRoundsParams{})
			},
		},
		{
			name: "delete unplayed round",
			keys: []string{"playerId"},
			build: func() (*http.Request, error) {
				return NewDeleteUnplayedRoundRequest("https://example.test", "E123", "S123", 1, &DeleteUnplayedRoundParams{})
			},
		},
		{
			name: "promo codes",
			keys: []string{"fuzzy"},
			build: func() (*http.Request, error) {
				return NewGetPromoCodesPageRequest("https://example.test", &GetPromoCodesPageParams{})
			},
		},
		{
			name: "rated events",
			keys: []string{"Name", "FromDate", "ToDate", "StateCode", "City", "ScholasticCode", "Women", "GrandPrix", "MinSize", "RatingSource", "DomesticStatus"},
			build: func() (*http.Request, error) {
				return NewGetRatedEventsPageRequest("https://example.test", &GetRatedEventsPageParams{})
			},
		},
		{
			name: "rated sections",
			keys: []string{"EventName", "PlayerName", "FromDate", "ToDate", "StateCode", "City", "DomesticStatus", "IsWomens", "IsGrandPrix", "IsGrandPrixJr", "IsGrandPrixExtended"},
			build: func() (*http.Request, error) {
				return NewGetRatedSectionsPageRequest("https://example.test", &GetRatedSectionsPageParams{})
			},
		},
		{
			name: "top players report",
			keys: []string{"Date", "Provisional"},
			build: func() (*http.Request, error) {
				return NewGetReportRequest("https://example.test", "definition", &GetReportParams{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.build()
			if err != nil {
				t.Fatalf("request builder returned an error: %v", err)
			}
			query := req.URL.Query()
			for _, key := range tt.keys {
				if _, ok := query[key]; ok {
					t.Errorf("%s = %q; want omitted", key, query.Get(key))
				}
			}
		})
	}
}

func TestOptionalFiltersCanTransmitZeroValues(t *testing.T) {
	falseValue := false
	zero := int32(0)
	empty := ""

	tests := []struct {
		name  string
		build func() (*http.Request, error)
		key   string
		want  string
	}{
		{
			name: "affiliate rated events women",
			build: func() (*http.Request, error) {
				return NewGetAffiliateRatedEventsRequest("https://example.test", "A123", &GetAffiliateRatedEventsParams{Women: &falseValue})
			},
			key:  "Women",
			want: "false",
		},
		{
			name: "members minimum rating",
			build: func() (*http.Request, error) {
				return NewGetMembersPageRequest("https://example.test", &GetMembersPageParams{MinRating: &zero})
			},
			key:  "MinRating",
			want: "0",
		},
		{
			name: "rated sections event name",
			build: func() (*http.Request, error) {
				return NewGetRatedSectionsPageRequest("https://example.test", &GetRatedSectionsPageParams{EventName: &empty})
			},
			key:  "EventName",
			want: "",
		},
		{
			name: "top players provisional",
			build: func() (*http.Request, error) {
				return NewGetReportRequest("https://example.test", "definition", &GetReportParams{Provisional: &falseValue})
			},
			key:  "Provisional",
			want: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.build()
			if err != nil {
				t.Fatalf("request builder returned an error: %v", err)
			}
			query := req.URL.Query()
			if _, ok := query[tt.key]; !ok {
				t.Fatalf("%s was omitted", tt.key)
			}
			if got := query.Get(tt.key); got != tt.want {
				t.Errorf("%s = %q; want %q", tt.key, got, tt.want)
			}
		})
	}
}

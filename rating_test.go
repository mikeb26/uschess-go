/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import "testing"

func TestRatingTypeString(t *testing.T) {
	tests := map[RatingType]string{
		RatingTypeU:     "Unknown",
		RatingTypeR:     "Regular",
		RatingTypeQ:     "Quick",
		RatingTypeB:     "Blitz",
		RatingTypeOR:    "Online Regular",
		RatingTypeOQ:    "Online Quick",
		RatingTypeOB:    "Online Blitz",
		RatingTypeC:     "Correspondence",
		RatingTypeF:     "FIDE",
		RatingTypeCFC:   "Chess Federation of Canada",
		RatingTypeDNR:   "Do Not Rate",
		RatingType("X"): "X",
	}

	for ratingType, want := range tests {
		t.Run(string(ratingType), func(t *testing.T) {
			if got := ratingType.String(); got != want {
				t.Errorf("String() = %q; want %q", got, want)
			}
		})
	}
}

func TestSectionRatingTypeString(t *testing.T) {
	tests := map[SectionRatingType]string{
		SectionRatingTypeU:     "Unknown",
		SectionRatingTypeR:     "Regular",
		SectionRatingTypeQ:     "Quick",
		SectionRatingTypeB:     "Blitz",
		SectionRatingTypeOR:    "Online Regular",
		SectionRatingTypeOQ:    "Online Quick",
		SectionRatingTypeOB:    "Online Blitz",
		SectionRatingTypeD:     "Dual",
		SectionRatingTypeG:     "FIDE G",
		SectionRatingTypeA:     "FIDE A",
		SectionRatingTypeF:     "FIDE F",
		SectionRatingType("X"): "X",
	}

	for ratingType, want := range tests {
		t.Run(string(ratingType), func(t *testing.T) {
			if got := ratingType.String(); got != want {
				t.Errorf("String() = %q; want %q", got, want)
			}
		})
	}
}

func TestRatingSupplementSystemString(t *testing.T) {
	tests := []struct {
		name   string
		rating RatingSupplementSystem
		want   string
	}{
		{
			name: "established rating",
			rating: RatingSupplementSystem{
				RatingType: RatingTypeR,
				Rating:     1800,
			},
			want: "Regular: 1800",
		},
		{
			name: "provisional rating",
			rating: RatingSupplementSystem{
				RatingType:           RatingTypeQ,
				Rating:               1200,
				ProvisionalGameCount: 5,
			},
			want: "Quick: 1200P5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rating.String(); got != tt.want {
				t.Errorf("String() = %q; want %q", got, tt.want)
			}
		})
	}
}

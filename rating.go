/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import "fmt"

// String returns the display name for a rating type.
func (r RatingType) String() string {
	switch r {
	case RatingTypeU:
		return "Unknown"
	case RatingTypeR:
		return "Regular"
	case RatingTypeQ:
		return "Quick"
	case RatingTypeB:
		return "Blitz"
	case RatingTypeOR:
		return "Online Regular"
	case RatingTypeOQ:
		return "Online Quick"
	case RatingTypeOB:
		return "Online Blitz"
	case RatingTypeC:
		return "Correspondence"
	case RatingTypeF:
		return "FIDE"
	case RatingTypeCFC:
		return "Chess Federation of Canada"
	case RatingTypeDNR:
		return "Do Not Rate"
	default:
		return string(r)
	}
}

// String returns the display name for a section rating type.
func (r SectionRatingType) String() string {
	switch r {
	case SectionRatingTypeU:
		return "Unknown"
	case SectionRatingTypeR:
		return "Regular"
	case SectionRatingTypeQ:
		return "Quick"
	case SectionRatingTypeB:
		return "Blitz"
	case SectionRatingTypeOR:
		return "Online Regular"
	case SectionRatingTypeOQ:
		return "Online Quick"
	case SectionRatingTypeOB:
		return "Online Blitz"
	case SectionRatingTypeD:
		return "Dual"
	case SectionRatingTypeG:
		return "FIDE G"
	case SectionRatingTypeA:
		return "FIDE A"
	case SectionRatingTypeF:
		return "FIDE F"
	default:
		return string(r)
	}
}

// String returns the rating system in the format used by US Chess rating supplements.
func (r RatingSupplementSystem) String() string {
	s := fmt.Sprintf("%s: %d", r.RatingType, r.Rating)
	if r.ProvisionalGameCount > 0 {
		s += fmt.Sprintf("P%d", r.ProvisionalGameCount)
	}
	return s
}

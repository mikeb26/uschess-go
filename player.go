/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import (
	"context"
	"fmt"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"golang.org/x/sync/errgroup"
)

// Player is a convenience aggregation of MemberDetail with optional other
// information regarding the member (e.g. rating supplements, recent
// tournaments, etc)
type Player struct {
	MemberDetail
	RatingSupplements   []RatingSupplement
	MemberRatedGames    []MemberRatedGame
	MemberRatedSections []MemberRatedSection
	LiveRatings         []MinimalRatingRecord
}

// GetPlayer retrieves memberID's details and optional aggregate data.
//
// When includeSupplements is true, it retrieves every page of rating
// supplements. When includeLiveRating is true, it includes rating records from
// sections ending after the most recent monthly-rating cutoff. When
// recentGamesOnOrAfterDate is non-nil, it retrieves every page of rated games
// on or after that date. When recentSectionsOnOrAfterDate is non-nil, it
// retrieves every page of rated sections on or after that date. The independent
// requests run concurrently and the first error cancels the remaining work.
func (c *ClientWithResponses) GetPlayer(ctx context.Context, memberID MemberID, includeSupplements, includeLiveRating bool, recentGamesOnOrAfterDate, recentSectionsOnOrAfterDate *time.Time, reqEditors ...RequestEditorFn) (*Player, error) {
	player := &Player{}
	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		response, err := c.GetMemberWithResponse(groupCtx, memberID, reqEditors...)
		if err != nil {
			return err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return fmt.Errorf("GetMember: unexpected response status %d", 0)
			}
			return fmt.Errorf("GetMember: unexpected response status %d: %s", response.StatusCode(), response.Body)
		}
		player.MemberDetail = *response.JSON200
		return nil
	})

	if includeSupplements {
		group.Go(func() error {
			supplements, err := c.GetAllRatingSupplements(groupCtx, memberID, reqEditors...)
			if err != nil {
				return err
			}
			player.RatingSupplements = supplements
			return nil
		})
	}

	if recentGamesOnOrAfterDate != nil {
		group.Go(func() error {
			games, err := c.GetAllMemberRatedGames(groupCtx, memberID, &GetMemberRatedGamesParams{
				OnOrAfterDate: openapi_types.Date{Time: *recentGamesOnOrAfterDate},
			}, reqEditors...)
			if err != nil {
				return err
			}
			player.MemberRatedGames = games
			return nil
		})
	}

	if includeLiveRating || recentSectionsOnOrAfterDate != nil {
		liveRatingCutoff := mostRecentMonthlyRatingCutoff(time.Now())
		sectionsOnOrAfterDate := recentSectionsOnOrAfterDate
		if includeLiveRating && (sectionsOnOrAfterDate == nil || liveRatingCutoff.Before(*sectionsOnOrAfterDate)) {
			sectionsOnOrAfterDate = &liveRatingCutoff
		}

		group.Go(func() error {
			sections, err := c.GetAllMemberRatedSections(groupCtx, memberID, &GetMemberRatedSectionsPageParams{
				OnOrAfterDate: openapi_types.Date{Time: *sectionsOnOrAfterDate},
			}, reqEditors...)
			if err != nil {
				return err
			}
			if recentSectionsOnOrAfterDate != nil {
				player.MemberRatedSections = sectionsOnOrAfter(sections, *recentSectionsOnOrAfterDate)
			}
			if includeLiveRating {
				player.LiveRatings = getLiveRatingRecords(sections, liveRatingCutoff)
			}
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}
	return player, nil
}

// getLiveRatingRecords returns rating records from sections that ended after
// cutoff. Per US Chess's Member Services FAQ
// (https://new.uschess.org/frequently-asked-questions-member-services-area),
// "the cutoff for a monthly ratings list is 11:45 PM Central Time on the 3rd
// Wednesday of the month." EndDate has date-only precision, so a section
// ending on that Wednesday is treated as not live. This also excludes the
// unlikely 11:45 PM--11:59 PM Central edge case on the cutoff date, which
// cannot be distinguished from an earlier end time.
func getLiveRatingRecords(sections []MemberRatedSection, cutoff time.Time) []MinimalRatingRecord {
	var records []MinimalRatingRecord
	for _, section := range sections {
		if !dateAfter(section.EndDate.Time, cutoff) {
			continue
		}
		records = append(records, section.RatingRecords...)
	}
	return records
}

func sectionsOnOrAfter(sections []MemberRatedSection, date time.Time) []MemberRatedSection {
	var filtered []MemberRatedSection
	for _, section := range sections {
		if dateOnOrAfter(section.EndDate.Time, date) {
			filtered = append(filtered, section)
		}
	}
	return filtered
}

func mostRecentMonthlyRatingCutoff(now time.Time) time.Time {
	central, err := time.LoadLocation("America/Chicago")
	if err != nil {
		// Go distributions normally include this IANA time zone. Keep the
		// cutoff calculation usable in an unusually restricted environment.
		central = time.FixedZone("CST", -6*60*60)
	}

	localNow := now.In(central)
	cutoff := thirdWednesdayCutoff(localNow.Year(), localNow.Month(), central)
	if localNow.Before(cutoff) {
		previousMonth := localNow.AddDate(0, -1, 0)
		cutoff = thirdWednesdayCutoff(previousMonth.Year(), previousMonth.Month(), central)
	}
	return cutoff
}

func thirdWednesdayCutoff(year int, month time.Month, location *time.Location) time.Time {
	firstOfMonth := time.Date(year, month, 1, 23, 45, 0, 0, location)
	daysUntilWednesday := (int(time.Wednesday) - int(firstOfMonth.Weekday()) + 7) % 7
	return firstOfMonth.AddDate(0, 0, daysUntilWednesday+14)
}

func dateAfter(date, other time.Time) bool {
	return date.Year() > other.Year() ||
		(date.Year() == other.Year() && (date.Month() > other.Month() ||
			(date.Month() == other.Month() && date.Day() > other.Day())))
}

func dateOnOrAfter(date, other time.Time) bool {
	return !dateBefore(date, other)
}

func dateBefore(date, other time.Time) bool {
	return date.Year() < other.Year() ||
		(date.Year() == other.Year() && (date.Month() < other.Month() ||
			(date.Month() == other.Month() && date.Day() < other.Day())))
}

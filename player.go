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
}

// GetPlayer retrieves memberID's details and optional aggregate data.
//
// When includeSupplements is true, it retrieves every page of rating
// supplements. When recentGamesOnOrAfterDate is non-nil, it retrieves every
// page of rated games on or after that date. When recentSectionsOnOrAfterDate
// is non-nil, it retrieves every page of rated sections on or after that date.
// The independent requests run concurrently and the first error cancels the
// remaining work.
func (c *ClientWithResponses) GetPlayer(ctx context.Context, memberID MemberID, includeSupplements bool, recentGamesOnOrAfterDate, recentSectionsOnOrAfterDate *time.Time, reqEditors ...RequestEditorFn) (*Player, error) {
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

	if recentSectionsOnOrAfterDate != nil {
		group.Go(func() error {
			sections, err := c.GetAllMemberRatedSections(groupCtx, memberID, &GetMemberRatedSectionsPageParams{
				OnOrAfterDate: openapi_types.Date{Time: *recentSectionsOnOrAfterDate},
			}, reqEditors...)
			if err != nil {
				return err
			}
			player.MemberRatedSections = sections
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}
	return player, nil
}

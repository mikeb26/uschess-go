/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// StandingsOneSection contains the standings/crosstable entries for one
// tournament section.
type StandingsOneSection []Standings

// Tournament is a convenience aggregation of RatedEventDetail with
// standings/crosstable information included for each section
type Tournament struct {
	RatedEventDetail
	SectionStandings []StandingsOneSection
}

// GetTournament retrieves eventID's details and standings from each section.
//
// The independent requests run concurrently and the first error cancels the
// remaining work.
func (c *ClientWithResponses) GetTournament(ctx context.Context,
	eventID EventID, reqEditors ...RequestEditorFn) (*Tournament, error) {
	response, err := c.GetRatedEventWithResponse(ctx, eventID, reqEditors...)
	if err != nil {
		return nil, fmt.Errorf("GetRatedEvent: %w", err)
	}
	if response == nil || response.JSON200 == nil {
		if response == nil {
			return nil, unexpectedPaginationResponse("GetRatedEvent", 0, nil)
		}
		return nil, unexpectedPaginationResponse("GetRatedEvent", response.StatusCode(), response.Body)
	}

	tournament := &Tournament{
		RatedEventDetail: *response.JSON200,
		SectionStandings: make([]StandingsOneSection, len(response.JSON200.Sections)),
	}

	group, groupCtx := errgroup.WithContext(ctx)
	for i, section := range tournament.Sections {
		i, section := i, section
		group.Go(func() error {
			standings, err := c.GetAllRatedEventStandings(groupCtx, eventID, section.Number, reqEditors...)
			if err != nil {
				return fmt.Errorf("GetRatedEventStandings for event %s section %d: %w",
					eventID, section.Number, err)
			}
			tournament.SectionStandings[i] = standings
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return tournament, nil
}

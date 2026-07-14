/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import "context"

// GetAllRatedEventStandings retrieves every standings page for eventID and sectionNumber, sorted by ordinal ascending.
func (c *ClientWithResponses) GetAllRatedEventStandings(ctx context.Context, eventID EventID, sectionNumber int32, reqEditors ...RequestEditorFn) ([]Standings, error) {
	pageParams := GetRatedEventStandingsPageParams{Size: defaultPaginationPageSize}
	items, err := collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[Standings], error) {
		pageParams.Offset = offset
		response, err := c.GetRatedEventStandingsPageWithResponse(ctx, eventID, sectionNumber, &pageParams, reqEditors...)
		if err != nil {
			return pageResult[Standings]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[Standings]{}, unexpectedPaginationResponse("GetRatedEventStandingsPage", 0, nil)
			}
			return pageResult[Standings]{}, unexpectedPaginationResponse("GetRatedEventStandingsPage", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[Standings]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
	if err != nil {
		return nil, err
	}
	sortStandingsByOrdinalAscending(items)
	return items, nil
}

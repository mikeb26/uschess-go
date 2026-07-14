/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import "context"

// GetAllPendingPlayers retrieves every pending-player page for pendingEventID and sectionID, sorted by pairing number ascending.
func (c *ClientWithResponses) GetAllPendingPlayers(ctx context.Context, pendingEventID EventID, sectionID SectionID, reqEditors ...RequestEditorFn) ([]PendingPlayer, error) {
	pageParams := ListPendingPlayersParams{
		SortBy: PendingPlayerSortByPairingNumber,
		Dir:    Asc,
	}
	return collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[PendingPlayer], error) {
		pageParams.Offset = offset
		response, err := c.ListPendingPlayersWithResponse(ctx, pendingEventID, sectionID, &pageParams, reqEditors...)
		if err != nil {
			return pageResult[PendingPlayer]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[PendingPlayer]{}, unexpectedPaginationResponse("ListPendingPlayers", 0, nil)
			}
			return pageResult[PendingPlayer]{}, unexpectedPaginationResponse("ListPendingPlayers", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[PendingPlayer]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
}

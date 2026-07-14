/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import "context"

// GetAllMemberAwards retrieves every awards page for memberID.
func (c *ClientWithResponses) GetAllMemberAwards(ctx context.Context, memberID MemberID, reqEditors ...RequestEditorFn) ([]MemberAward, error) {
	pageParams := GetMemberAwardsPageParams{Size: defaultPaginationPageSize}
	return collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[MemberAward], error) {
		pageParams.Offset = offset
		response, err := c.GetMemberAwardsPageWithResponse(ctx, memberID, &pageParams, reqEditors...)
		if err != nil {
			return pageResult[MemberAward]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[MemberAward]{}, unexpectedPaginationResponse("GetMemberAwardsPage", 0, nil)
			}
			return pageResult[MemberAward]{}, unexpectedPaginationResponse("GetMemberAwardsPage", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[MemberAward]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
}

// GetAllMemberDirectorships retrieves every directorships page for memberID.
func (c *ClientWithResponses) GetAllMemberDirectorships(ctx context.Context, memberID MemberID, reqEditors ...RequestEditorFn) ([]MemberDirectorship, error) {
	pageParams := GetMemberDirectorshipsPageParams{Size: defaultPaginationPageSize}
	return collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[MemberDirectorship], error) {
		pageParams.Offset = offset
		response, err := c.GetMemberDirectorshipsPageWithResponse(ctx, memberID, &pageParams, reqEditors...)
		if err != nil {
			return pageResult[MemberDirectorship]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[MemberDirectorship]{}, unexpectedPaginationResponse("GetMemberDirectorshipsPage", 0, nil)
			}
			return pageResult[MemberDirectorship]{}, unexpectedPaginationResponse("GetMemberDirectorshipsPage", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[MemberDirectorship]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
}

// GetAllMemberRatedGames retrieves every rated-games page for memberID matching pageParams, sorted by event start date descending.
//
// When pageParams is nil, it retrieves all rated games. Its Offset is managed while retrieving pages.
func (c *ClientWithResponses) GetAllMemberRatedGames(ctx context.Context, memberID MemberID, pageParams *GetMemberRatedGamesParams, reqEditors ...RequestEditorFn) ([]MemberRatedGame, error) {
	if pageParams == nil {
		pageParams = &GetMemberRatedGamesParams{Size: defaultPaginationPageSize}
	}
	if pageParams.Size <= 0 {
		pageParams.Size = defaultPaginationPageSize
	}

	items, err := collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[MemberRatedGame], error) {
		pageParams.Offset = offset
		response, err := c.GetMemberRatedGamesWithResponse(ctx, memberID, pageParams, reqEditors...)
		if err != nil {
			return pageResult[MemberRatedGame]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[MemberRatedGame]{}, unexpectedPaginationResponse("GetMemberRatedGames", 0, nil)
			}
			return pageResult[MemberRatedGame]{}, unexpectedPaginationResponse("GetMemberRatedGames", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[MemberRatedGame]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
	if err != nil {
		return nil, err
	}
	sortMemberRatedGamesByEventStartDateDescending(items)
	return items, nil
}

// GetAllMemberRatedSections retrieves every rated-sections page for memberID, sorted by start date descending.
func (c *ClientWithResponses) GetAllMemberRatedSections(ctx context.Context, memberID MemberID, reqEditors ...RequestEditorFn) ([]MemberRatedSection, error) {
	pageParams := GetMemberRatedSectionsPageParams{Size: defaultPaginationPageSize}
	items, err := collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[MemberRatedSection], error) {
		pageParams.Offset = offset
		response, err := c.GetMemberRatedSectionsPageWithResponse(ctx, memberID, &pageParams, reqEditors...)
		if err != nil {
			return pageResult[MemberRatedSection]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[MemberRatedSection]{}, unexpectedPaginationResponse("GetMemberRatedSectionsPage", 0, nil)
			}
			return pageResult[MemberRatedSection]{}, unexpectedPaginationResponse("GetMemberRatedSectionsPage", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[MemberRatedSection]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
	if err != nil {
		return nil, err
	}
	sortMemberRatedSectionsByStartDateDescending(items)
	return items, nil
}

// GetAllTopPlayersReportsForMember retrieves every top-player-reports page for memberID.
func (c *ClientWithResponses) GetAllTopPlayersReportsForMember(ctx context.Context, memberID MemberID, reqEditors ...RequestEditorFn) ([]TopPlayerReport, error) {
	pageParams := GetTopPlayersReportForMemberParams{Size: defaultPaginationPageSize}
	return collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[TopPlayerReport], error) {
		pageParams.Offset = offset
		response, err := c.GetTopPlayersReportForMemberWithResponse(ctx, memberID, &pageParams, reqEditors...)
		if err != nil {
			return pageResult[TopPlayerReport]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[TopPlayerReport]{}, unexpectedPaginationResponse("GetTopPlayersReportForMember", 0, nil)
			}
			return pageResult[TopPlayerReport]{}, unexpectedPaginationResponse("GetTopPlayersReportForMember", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[TopPlayerReport]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
}

/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import "context"

// GetAllAffiliateRatedEvents retrieves every rated-events page for affiliateID, sorted by start date descending.
func (c *ClientWithResponses) GetAllAffiliateRatedEvents(ctx context.Context,
	affiliateID AffiliateID, pageParams *GetAffiliateRatedEventsParams,
	reqEditors ...RequestEditorFn) ([]RatedEvent, error) {

	if pageParams == nil {
		pageParams = &GetAffiliateRatedEventsParams{
			SortBy: AffiliateEventSortByStartDate,
			Dir:    Desc,
			Size:   defaultPaginationPageSize,
		}
	}
	if !pageParams.SortBy.Valid() {
		pageParams.SortBy = AffiliateEventSortByStartDate
		pageParams.Dir = Desc
	}
	if pageParams.Size <= 0 {
		pageParams.Size = defaultPaginationPageSize
	}

	return collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[RatedEvent], error) {
		pageParams.Offset = offset
		response, err := c.GetAffiliateRatedEventsWithResponse(ctx, affiliateID, pageParams, reqEditors...)
		if err != nil {
			return pageResult[RatedEvent]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[RatedEvent]{}, unexpectedPaginationResponse("GetAffiliateRatedEvents", 0, nil)
			}
			return pageResult[RatedEvent]{}, unexpectedPaginationResponse("GetAffiliateRatedEvents", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[RatedEvent]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
}

// GetAllMemberEvents retrieves every rated-events page for memberID, sorted
// by start date descending.
func (c *ClientWithResponses) GetAllMemberEvents(ctx context.Context, memberID MemberID, reqEditors ...RequestEditorFn) ([]RatedEvent, error) {
	pageParams := GetMemberRatedEventsPageParams{Size: defaultPaginationPageSize}
	items, err := collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[RatedEvent], error) {
		pageParams.Offset = offset
		response, err := c.GetMemberRatedEventsPageWithResponse(ctx, memberID, &pageParams, reqEditors...)
		if err != nil {
			return pageResult[RatedEvent]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[RatedEvent]{}, unexpectedPaginationResponse("GetMemberRatedEventsPage", 0, nil)
			}
			return pageResult[RatedEvent]{}, unexpectedPaginationResponse("GetMemberRatedEventsPage", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[RatedEvent]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
	if err != nil {
		return nil, err
	}
	sortRatedEventsByStartDateDescending(items)
	return items, nil
}

// GetAllRatedEvents retrieves every rated-events page, sorted by start date descending.
func (c *ClientWithResponses) GetAllRatedEvents(ctx context.Context,
	pageParams *GetRatedEventsPageParams,
	reqEditors ...RequestEditorFn) ([]RatedEvent, error) {

	if pageParams == nil {
		pageParams = &GetRatedEventsPageParams{
			SortBy: RatedEventSortByStartDate,
			Dir:    Desc,
			Size:   defaultPaginationPageSize,
		}
	}
	if !pageParams.SortBy.Valid() {
		pageParams.SortBy = RatedEventSortByStartDate
		pageParams.Dir = Desc
	}
	if pageParams.Size <= 0 {
		pageParams.Size = defaultPaginationPageSize
	}
	return collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[RatedEvent], error) {
		pageParams.Offset = offset
		response, err := c.GetRatedEventsPageWithResponse(ctx, pageParams, reqEditors...)
		if err != nil {
			return pageResult[RatedEvent]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[RatedEvent]{}, unexpectedPaginationResponse("GetRatedEventsPage", 0, nil)
			}
			return pageResult[RatedEvent]{}, unexpectedPaginationResponse("GetRatedEventsPage", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[RatedEvent]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
}

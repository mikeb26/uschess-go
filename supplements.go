/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */
package uschess

import "context"

// GetAllRatingSupplements retrieves every rating-supplements page for memberID, sorted by rating supplement date descending.
func (c *ClientWithResponses) GetAllRatingSupplements(ctx context.Context, memberID MemberID, reqEditors ...RequestEditorFn) ([]RatingSupplement, error) {
	pageParams := GetRatingSupplementsPageParams{}
	items, err := collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[RatingSupplement], error) {
		pageParams.Offset = offset
		response, err := c.GetRatingSupplementsPageWithResponse(ctx, memberID, &pageParams, reqEditors...)
		if err != nil {
			return pageResult[RatingSupplement]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[RatingSupplement]{}, unexpectedPaginationResponse("GetRatingSupplementsPage", 0, nil)
			}
			return pageResult[RatingSupplement]{}, unexpectedPaginationResponse("GetRatingSupplementsPage", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[RatingSupplement]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
	if err != nil {
		return nil, err
	}
	sortRatingSupplementsByDateDescending(items)
	return items, nil
}

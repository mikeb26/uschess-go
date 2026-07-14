/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import "context"

// GetAllAffiliates retrieves every page of affiliates, sorted by name ascending.
func (c *ClientWithResponses) GetAllAffiliates(ctx context.Context,
	pageParams *GetAffiliatesPageParams, reqEditors ...RequestEditorFn) ([]Affiliate, error) {

	if pageParams == nil {
		pageParams = &GetAffiliatesPageParams{
			SortBy: AffiliateSortByName,
			Dir:    Asc,
			Size:   defaultPaginationPageSize,
		}
	}
	if !pageParams.SortBy.Valid() {
		pageParams.SortBy = AffiliateSortByName
		pageParams.Dir = Asc
	}
	if pageParams.Size <= 0 {
		pageParams.Size = defaultPaginationPageSize
	}

	return collectPages(ctx, func(ctx context.Context, offset int32) (pageResult[Affiliate], error) {
		pageParams.Offset = offset
		response, err := c.GetAffiliatesPageWithResponse(ctx, pageParams, reqEditors...)
		if err != nil {
			return pageResult[Affiliate]{}, err
		}
		if response == nil || response.JSON200 == nil {
			if response == nil {
				return pageResult[Affiliate]{}, unexpectedPaginationResponse("GetAffiliatesPage", 0, nil)
			}
			return pageResult[Affiliate]{}, unexpectedPaginationResponse("GetAffiliatesPage", response.StatusCode(), response.Body)
		}
		page := response.JSON200
		return pageResult[Affiliate]{page.Items, page.Offset, page.PageSize, page.HasNextPage}, nil
	})
}

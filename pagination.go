/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

package uschess

import (
	"context"
	"fmt"
	"math"
	"sort"
)

// defaultPaginationPageSize must be positive because the API returns a zero
// PageSize when the Size query parameter is omitted, even if more pages exist.
const defaultPaginationPageSize int32 = 100

type pageResult[T any] struct {
	items       []T
	offset      int32
	pageSize    int32
	hasNextPage bool
}

// collectPages fetches pages from offset zero until the API reports there are no more pages.
func collectPages[T any](ctx context.Context, fetch func(context.Context, int32) (pageResult[T], error)) ([]T, error) {
	var items []T
	var offset int32

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		page, err := fetch(ctx, offset)
		if err != nil {
			return nil, err
		}
		items = append(items, page.items...)
		if !page.hasNextPage {
			return items, nil
		}
		if page.pageSize <= 0 {
			return nil, fmt.Errorf("pagination response at offset %d has hasNextPage=true with a non-positive page size %d", offset, page.pageSize)
		}

		nextOffset := int64(page.offset) + int64(page.pageSize)
		if nextOffset > math.MaxInt32 || nextOffset <= int64(offset) {
			return nil, fmt.Errorf("pagination response cannot advance from offset %d to %d", offset, nextOffset)
		}
		offset = int32(nextOffset)
	}
}

func sortRatingSupplementsByDateDescending(items []RatingSupplement) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].RatingSupplementDate.After(items[j].RatingSupplementDate.Time)
	})
}

func sortRatedEventsByStartDateDescending(items []RatedEvent) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].StartDate.After(items[j].StartDate.Time)
	})
}

func sortMemberRatedGamesByEventStartDateDescending(items []MemberRatedGame) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Event.StartDate.After(items[j].Event.StartDate.Time)
	})
}

func sortMemberRatedSectionsByStartDateDescending(items []MemberRatedSection) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].StartDate.After(items[j].StartDate.Time)
	})
}

func sortRatedSectionsByRatedDateDescending(items []RatedSection) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].RatedDate.After(items[j].RatedDate.Time)
	})
}

func sortStandingsByOrdinalAscending(items []Standings) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Ordinal < items[j].Ordinal
	})
}

func unexpectedPaginationResponse(operation string, statusCode int, body []byte) error {
	return fmt.Errorf("%s: unexpected response status %d: %s", operation, statusCode, body)
}

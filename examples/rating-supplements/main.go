/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

// Command ratingsupplements prints a member's rating supplements.
package main

import (
	"context"
	"fmt"
	"log"

	uschess "github.com/mikeb26/uschess-go"
)

func main() {
	ctx := context.Background()
	memberID := uschess.MemberID("12641216")

	client, err := uschess.NewDefaultClient()
	if err != nil {
		log.Fatal(err)
	}

	supplements, err := client.GetAllRatingSupplements(ctx, memberID)
	if err != nil {
		log.Fatalf("get rating supplements: %v", err)
	}

	for _, supplement := range supplements {
		fmt.Printf("Rating supplement: %s\n", supplement.RatingSupplementDate)
		for _, rating := range supplement.Ratings {
			if rating.Rating == 0 {
				continue
			}

			fmt.Printf("  %s: %d", rating.RatingType, rating.Rating)
			if rating.ProvisionalGameCount > 0 {
				fmt.Printf("P%d", rating.ProvisionalGameCount)
			}
			fmt.Printf("\n")
		}
	}
}

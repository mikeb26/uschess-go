/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

// Command rating-estimate estimates a player's post-event Regular rating.
package main

import (
	"context"
	"fmt"
	"log"

	uschess "github.com/mikeb26/uschess-go"
)

func main() {
	ctx := context.Background()

	client, err := uschess.NewDefaultClient()
	if err != nil {
		log.Fatal(err)
	}

	playerID := uschess.MemberID("12641216") // hikaru nakamura
	opponentIDs := []uschess.MemberID{
		"12743305", // fabiano caruana
		"13145890", // wesley so
		"15218444", // levon aronian
	}
	const score = 2.5
	const ratingType = uschess.RatingTypeR

	estimate, err := client.GetRatingEstimate(
		ctx,
		playerID,
		opponentIDs,
		score,
		ratingType,
	)
	if err != nil {
		log.Fatalf("estimate rating: %v", err)
	}

	fmt.Printf("With a score of %v out of %v in the event:\n", score, len(opponentIDs))
	fmt.Printf("Pre-event %s rating for member %s:            %d\n", ratingType, playerID, estimate.PreRating)
	fmt.Printf("Estimated post-event %s rating for member %s: %d\n", ratingType, playerID, estimate.PostRating)
}

/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

// Command player retrieves a member with every GetPlayer option enabled.
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

	memberID := uschess.MemberID("12641216")
	player, err := client.GetPlayer(ctx, memberID, nil)
	if err != nil {
		log.Fatalf("get player: %v", err)
	}

	fmt.Printf("%s %s (ID: %s)\n", player.FirstName, player.LastName, player.Id)

	if len(player.RatingSupplements) == 0 {
		fmt.Println("\nMost recent rating supplement: none")
	} else {
		supplement := player.RatingSupplements[0]
		fmt.Printf("\nMost recent rating supplement: %s\n", supplement.RatingSupplementDate)
		for _, rating := range supplement.Ratings {
			if rating.Rating != 0 {
				fmt.Printf("  %s\n", rating)
			}
		}
	}

	liveRatings, err := player.LiveRatings()
	if err != nil {
		log.Fatalf("get live ratings: %v", err)
	}
	fmt.Println("\nLive rating:")
	for _, rating := range liveRatings {
		fmt.Printf("  %s\n", rating)
	}

	fmt.Println("\nMost recent 10 games:")
	games := player.MemberRatedGames
	if len(games) > 10 {
		games = games[:10]
	}
	if len(games) == 0 {
		fmt.Println("  none")
	}
	var previousEventID uschess.EventID
	var previousSectionID uschess.SectionID
	for _, game := range games {
		if game.Event.Id != previousEventID {
			fmt.Printf("  Event: %s (%s)\n", game.Event.Name, game.Event.StartDate)
			previousEventID = game.Event.Id
			previousSectionID = ""
		}
		if game.Section.Id != previousSectionID {
			fmt.Printf("    Section: %s\n", game.Section.Name)
			previousSectionID = game.Section.Id
		}
		fmt.Printf(
			"      %s vs. %s %s (%s)\n",
			game.Player.Color,
			game.Opponent.FirstName,
			game.Opponent.LastName,
			game.Player.Outcome,
		)
	}
}

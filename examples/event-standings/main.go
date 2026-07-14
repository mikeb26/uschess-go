/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

// Command eventstandings prints an event's standings, grouped by section.
package main

import (
	"context"
	"fmt"
	"log"

	uschess "github.com/mikeb26/uschess-go"
)

func main() {
	ctx := context.Background()
	eventID := uschess.EventID("202606140693") // Replace with an event ID.

	client, err := uschess.NewDefaultClient()
	if err != nil {
		log.Fatal(err)
	}

	response, err := client.GetRatedEventWithResponse(ctx, eventID)
	if err != nil {
		log.Fatal(err)
	}
	if response.JSON200 == nil {
		log.Fatalf("unexpected event response: HTTP %d: %s", response.StatusCode(), response.Body)
	}

	event := response.JSON200
	fmt.Printf("%s (%s)\n", event.Name, event.Id)
	for _, section := range event.Sections {
		standings, err := client.GetAllRatedEventStandings(ctx, eventID, section.Number)
		if err != nil {
			log.Fatalf("get standings for section %d: %v", section.Number, err)
		}

		fmt.Printf("\nSection %d: %s\n", section.Number, section.Name)
		for _, standing := range standings {
			fmt.Printf("%d. %s %s — %.1f\n", standing.Ordinal, standing.FirstName, standing.LastName, standing.Score)
		}
	}
}

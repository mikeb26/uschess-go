/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

// Command eventdescription prints a summary of a rated event.
package main

import (
	"context"
	"fmt"
	"log"

	uschess "github.com/mikeb26/uschess-go"
)

func main() {
	ctx := context.Background()
	eventID := uschess.EventID("202606140693")

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
	fmt.Printf("Dates: %s through %s\n", event.StartDate, event.EndDate)
	fmt.Printf("Location: %s, %s %s\n", event.City, event.StateCode, event.ZipCode)
	fmt.Printf("Affiliate: %s (%s)\n", event.Affiliate.Name, event.Affiliate.Id)
	fmt.Printf("Status: %s\n", event.Status)
	fmt.Printf("Players: %d; sections: %d; games: %d\n", event.PlayerCount, event.SectionCount, event.GameCount)
}

/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

// Command member fetches a US Chess member profile using the generated OpenAPI client.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mikeb26/uschess-go"
)

func main() {
	ctx := context.Background()

	client, err := uschess.NewDefaultClient()
	if err != nil {
		log.Fatal(err)
	}

	response, err := client.GetMemberWithResponse(ctx, "12641216")
	if err != nil {
		log.Fatal(err)
	}
	if response.JSON200 == nil {
		log.Fatalf("unexpected profile status %d: %s", response.StatusCode(), response.Body)
	}

	member := response.JSON200
	fmt.Printf("id: %s\n", member.Id)
	fmt.Printf("USCFTitle: %s\n", member.UscfTitle)
	fmt.Printf("FirstName: %s\n", member.FirstName)
	fmt.Printf("LastName: %s\n", member.LastName)
	for _, rating := range member.Ratings {
		if rating.RatingType == uschess.RatingTypeR {
			fmt.Printf("regular rating: %d\n", rating.Rating)
			break
		}
	}
}

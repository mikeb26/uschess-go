/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

// Command affiliatedescription prints a summary of a US Chess affiliate.
package main

import (
	"context"
	"fmt"
	"log"

	uschess "github.com/mikeb26/uschess-go"
)

func main() {
	ctx := context.Background()
	affiliateID := uschess.AffiliateID("A5000408") // Replace with an affiliate ID.

	client, err := uschess.NewDefaultClient()
	if err != nil {
		log.Fatal(err)
	}

	response, err := client.GetAffiliateWithResponse(ctx, affiliateID)
	if err != nil {
		log.Fatal(err)
	}
	if response.JSON200 == nil {
		log.Fatalf("unexpected affiliate response: HTTP %d: %s", response.StatusCode(), response.Body)
	}

	affiliate := response.JSON200
	fmt.Printf("%s (%s)\n", affiliate.Name, affiliate.Id)
	fmt.Printf("State: %s\n", affiliate.StateCode)
	fmt.Printf("Status: %s\n", affiliate.Status)
	fmt.Printf("Expiration date: %s\n", affiliate.ExpirationDate)
	fmt.Printf("Last changed: %s\n", affiliate.LastChangedDate)
}

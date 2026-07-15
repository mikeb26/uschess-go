/* Copyright © 2026 Mike Brown. All Rights Reserved.
 *
 * See LICENSE file at the root of this repository for license terms
 */

// Command crosstable writes a CSV crosstable for each section of an event.
package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	uschess "github.com/mikeb26/uschess-go"
)

func main() {
	ctx := context.Background()
	eventID := uschess.EventID("202606140693")

	client, err := uschess.NewDefaultClient()
	if err != nil {
		log.Fatal(err)
	}

	tournament, err := client.GetTournament(ctx, eventID)
	if err != nil {
		log.Fatal(err)
	}

	for i, section := range tournament.Sections {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("Section %v:\n", tournament.Sections[i].Name)
		if err := writeSection(os.Stdout, tournament.SectionStandings[i]); err != nil {
			log.Fatalf("write crosstable for section %d (%s): %v", section.Number, section.Name, err)
		}
	}
}

func writeSection(output io.Writer, standings uschess.StandingsOneSection) error {
	roundCount := maximumRoundCount(standings)
	writer := csv.NewWriter(output)

	header := []string{"#", "Place", "Name", "PreR", "PostR"}
	for round := 1; round <= roundCount; round++ {
		header = append(header, fmt.Sprintf("Rd %d", round))
	}
	header = append(header, "Tot")
	if err := writer.Write(header); err != nil {
		return err
	}

	places := placesByOrdinal(standings)
	for _, standing := range standings {
		record := []string{
			strconv.FormatInt(int64(standing.Ordinal), 10),
			strconv.Itoa(places[standing.Ordinal]),
			standing.FirstName + " " + standing.LastName,
		}
		preRating, postRating := regularRatings(standing.Ratings)
		record = append(record, preRating, postRating)

		outcomes := make(map[int32]uschess.StandingsRound, len(standing.RoundOutcomes))
		for _, outcome := range standing.RoundOutcomes {
			outcomes[outcome.RoundNumber] = outcome
		}
		for round := 1; round <= roundCount; round++ {
			record = append(record, formatRound(outcomes[int32(round)]))
		}
		record = append(record, strconv.FormatFloat(float64(standing.Score), 'f', 1, 32))
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

// placesByOrdinal assigns competition places based on score: equal scores
// share a place, and the next place accounts for all tied players.
func placesByOrdinal(standings uschess.StandingsOneSection) map[int32]int {
	places := make(map[int32]int, len(standings))
	previousScore := float32(-1)
	place := 0
	for index, standing := range standings {
		if index == 0 || standing.Score != previousScore {
			place = index + 1
			previousScore = standing.Score
		}
		places[standing.Ordinal] = place
	}
	return places
}

func maximumRoundCount(standings uschess.StandingsOneSection) int {
	maximum := 0
	for _, standing := range standings {
		for _, outcome := range standing.RoundOutcomes {
			if int(outcome.RoundNumber) > maximum {
				maximum = int(outcome.RoundNumber)
			}
		}
	}
	return maximum
}

func regularRatings(ratings []uschess.RatingRecord) (pre, post string) {
	for _, rating := range ratings {
		if rating.RatingType == uschess.RatingTypeR {
			return formatRating(rating.PreRating), formatRating(rating.PostRating)
		}
	}
	return "", ""
}

func formatRating(rating int32) string {
	if rating == 0 {
		return ""
	}
	return strconv.FormatInt(int64(rating), 10)
}

func formatRound(round uschess.StandingsRound) string {
	if round.Outcome == "" {
		return ""
	}

	result := "?"
	switch round.Outcome {
	case uschess.PlayerOutcomeWin, uschess.PlayerOutcomeWinAsym:
		result = "W"
	case uschess.PlayerOutcomeWinForfeit:
		result = "X"
	case uschess.PlayerOutcomeLoss, uschess.PlayerOutcomeLossAsym:
		result = "L"
	case uschess.PlayerOutcomeForfeit:
		result = "F"
	case uschess.PlayerOutcomeDraw, uschess.PlayerOutcomeDrawAsym, uschess.PlayerOutcomeDrawForfeit:
		result = "D"
	case uschess.PlayerOutcomeByeFull:
		return "B"
	case uschess.PlayerOutcomeByeHalf:
		return "H"
	case uschess.PlayerOutcomeUnpaired:
		return "U"
	case uschess.PlayerOutcomeNotReported:
		return "NR"
	case uschess.PlayerOutcomeReportNoResult:
		return "RNR"
	}
	if round.OpponentOrdinal == 0 {
		return result + "?"
	}
	return result + strconv.FormatInt(int64(round.OpponentOrdinal), 10)
}

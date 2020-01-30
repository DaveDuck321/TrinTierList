package main

import (
	"fmt"
	"math"
)

func updateRankings(elos rankings, winnerID, looserID, categoryID int, weight float64) error {
	category, ok := elos[categoryID]
	if !ok {
		return fmt.Errorf("category not found")
	}
	if _, ok := category[winnerID]; !ok {
		return fmt.Errorf("player ID not found")
	}
	if _, ok := category[looserID]; !ok {
		return fmt.Errorf("player ID not found")
	}

	pWin := 1.0 / (1.0 + math.Pow(10, float64(category[looserID]-category[winnerID])/400.0))
	pLoose := 1.0 / (1.0 + math.Pow(10, float64(category[winnerID]-category[looserID])/400.0))

	category[winnerID] += int(weight * (2 - pWin))
	category[looserID] -= int(weight * (2 - pLoose))

	return nil
}

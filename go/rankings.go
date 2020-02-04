package main

import (
	"fmt"
	"math"
)

func updateRankings(elos rankings, winnerID, looserID, categoryID int, weight float64) (int, int, error) {
	category, ok := elos[categoryID]
	if !ok {
		return 0, 0, fmt.Errorf("category not found")
	}
	if _, ok := category[winnerID]; !ok {
		return 0, 0, fmt.Errorf("player ID not found")
	}
	if _, ok := category[looserID]; !ok {
		return 0, 0, fmt.Errorf("player ID not found")
	}

	pWin := 1.0 / (1.0 + math.Pow(10, float64(category[looserID]-category[winnerID])/400.0))
	pLoose := 1.0 / (1.0 + math.Pow(10, float64(category[winnerID]-category[looserID])/400.0))

	p1Change := int(weight * (2 - pWin))
	p2Change := -int(weight * (2 - pLoose))

	category[winnerID] += p1Change
	category[looserID] += p2Change

	return p1Change, p2Change, nil
}

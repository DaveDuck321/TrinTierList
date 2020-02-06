package main

import (
	"fmt"
	"math"
)

func probabilityOfWin(thisELO, otherELO int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(otherELO-thisELO)/400.0))
}

func updateRankings(elos rankings, winnerID, looserID, categoryID int, weight float64) (int, error) {
	category, ok := elos[categoryID]
	if !ok {
		return 0, fmt.Errorf("category not found")
	}
	if _, ok := category[winnerID]; !ok {
		return 0, fmt.Errorf("player ID not found")
	}
	if _, ok := category[looserID]; !ok {
		return 0, fmt.Errorf("player ID not found")
	}

	pWin := probabilityOfWin(category[winnerID], category[looserID])

	eloChange := int(weight * (1 - pWin))
	category[winnerID] += eloChange
	category[looserID] -= eloChange

	return eloChange, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
)

//BigConst is just a big number
const BigConst = (1 << 8)

type rankings map[int](map[int]int)

func getEngineers() []person {
	var people []person
	data, _ := ioutil.ReadFile("data/people.json")
	json.Unmarshal(data, &people)

	return people
}

func getCategories() []category {
	var categories []category
	data, _ := ioutil.ReadFile("data/categories.json")
	json.Unmarshal(data, &categories)

	return categories
}

func getRankingsResults(people []person, categories []category) rankings {
	rankings := make(rankings)
	data, _ := ioutil.ReadFile("data/matchResult.json")
	json.Unmarshal(data, &rankings)

	//Fills empty result with defaults
	for _, c := range categories {
		if _, ok := rankings[c.ID]; !ok {
			rankings[c.ID] = make(map[int]int)
		}
		for _, p := range people {
			if _, ok := rankings[c.ID][p.ID]; !ok {
				rankings[c.ID][p.ID] = 1000
			}
		}
	}
	return rankings
}

func randomMatch(available []int) (int, int, error) {
	if len(available) == 0 {
		return 0, 0, fmt.Errorf("no available matches")
	}
	matchID := available[rand.Int()%len(available)]

	return decodeMatchID(matchID)
}

func getPeopleFromIDs(peopleMap map[int]person, id1, id2 int) (person, person, error) {
	p1, ok := peopleMap[id1]
	if !ok {
		return person{}, person{}, fmt.Errorf("person not found")
	}
	p2, ok := peopleMap[id2]
	if !ok {
		return person{}, person{}, fmt.Errorf("person not found")
	}
	return p1, p2, nil
}

func decodeMatchID(matchID int) (int, int, error) {
	return matchID % BigConst, matchID / BigConst, nil
}

func getMatchID(p1ID int, p2ID int) int {
	if p1ID > p2ID {
		return p1ID + p2ID*BigConst
	}
	return p1ID*BigConst + p2ID
}

func genPermutations(people []person) []int {
	var permutations []int

	for index, p1 := range people {
		for _, p2 := range people[index+1:] {
			permutations = append(permutations, getMatchID(p1.ID, p2.ID))
		}
	}
	return permutations
}

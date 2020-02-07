package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
)

//BigConst is just a big number
const BigConst = (1 << 8)

type rankings map[int](map[int]int)

func getPeople() []person {
	var people []person
	data, _ := ioutil.ReadFile("data/people.json")
	json.Unmarshal(data, &people)

	return people
}

func makePeopleMap(people []person) map[int]person {
	result := make(map[int]person)
	for _, person := range people {
		result[person.ID] = person
	}
	return result
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

func nextMatch(available []int) (int, int, error) {
	if len(available) == 0 {
		return 0, 0, fmt.Errorf("no available matches")
	}

	id1, id2 := decodeMatchID(available[0])
	if rand.Int()%2 == 0 {
		return id1, id2, nil
	}
	return id2, id1, nil
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

func decodeMatchID(matchID int) (int, int) {
	return matchID % BigConst, matchID / BigConst
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

func splitHostPort(hostname string) (string, string) {
	portIndex := strings.LastIndex(hostname, ":")
	return hostname[:portIndex], hostname[portIndex:]
}

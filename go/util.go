package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

//BigConst is just a big number
const BigConst = (1 << 8)

type rankingsBackup struct {
	Date      time.Time        `json:"time"`
	Rankings  rankings         `json:"rankings"`
	VoteCount map[string](int) `json:"voteCount"`
}

func numberOfPermutations(peopleCount int) int {
	return (peopleCount * (peopleCount - 1)) / 2
}

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

func saveELOs(elos rankings) {
	data, _ := json.Marshal(elos)
	ioutil.WriteFile("data/matchResult.json", data, 0644)
}

func backupMatchResults(votes allAvailableVotes) {
	t := time.Now()
	fileName := fmt.Sprintf("data/backup/%d-%02d-%02d_%02d.json",
		t.Year(), t.Month(), t.Day(), t.Hour())

	savedElosData, _ := ioutil.ReadFile("data/matchResult.json")

	var elos rankings
	json.Unmarshal(savedElosData, &elos)

	numberOfVotes := numberOfPermutations(len(votes))

	playerVoteCount := make(map[string](int))
	for personID, votesByCategory := range votes {
		totalVotes := 0
		for _, votes := range votesByCategory {
			totalVotes += numberOfVotes - len(votes)
		}

		playerVoteCount[personID] = totalVotes
	}
	backupData, _ := json.Marshal(
		rankingsBackup{
			t, elos, playerVoteCount,
		})
	ioutil.WriteFile(fileName, backupData, 0644)
}

func refreshTempDataInterval(votes allAvailableVotes, people []person, categories []category, timePeriod *time.Ticker) {
	for {
		<-timePeriod.C
		backupMatchResults(votes)
		updateAvailableVotes(votes, people, categories)
	}
}

func saveELOsInterval(matchResults rankings, timePeriod *time.Ticker) {
	for {
		<-timePeriod.C
		saveELOs(matchResults)
	}
}

func updateAvailableVotes(votes allAvailableVotes, people []person, categories []category) allAvailableVotes {
	allMatches := genPermutations(people)
	rand.Seed(time.Now().UnixNano())
	for _, person := range people {
		votes[person.CrsID] = make(availableVotes)
		for _, category := range categories {
			pVotes := make([]int, len(allMatches))
			copy(pVotes, allMatches)

			//Display in random order
			rand.Shuffle(len(allMatches), func(i, j int) {
				pVotes[i], pVotes[j] = pVotes[j], pVotes[i]
			})
			votes[person.CrsID][category.ID] = pVotes
		}
	}
	fmt.Println("Votes have been refreshed")
	return votes
}

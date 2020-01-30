package main

import (
	"encoding/json"
	"io/ioutil"
)

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

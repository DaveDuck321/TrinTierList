package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
)

type peopleResponse struct {
	Category category `json:"category"`
	Person1  person   `json:"person1"`
	Person2  person   `json:"person2"`

	Success bool   `json:"success"`
	Error   string `json:"msg"`
}

type person struct {
	ID   int      `json:"id"`
	Name string   `json:"name"`
	Imgs []string `json:"imgs"`
}

type category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type matchResult struct {
	Category int `json:"category"`
	ID1      int `json:"id1"`
	ID2      int `json:"id2"`
	Wins     int `json:"wins"`
}

type voteResult struct {
	WonID      int `json:"won"`
	LostID     int `json:"lost"`
	CategoryID int `json:"category"`
}

func getMatchID(id1, id2 int) string {
	if id1 > id2 {
		return fmt.Sprint(id1, "_", id2)
	}
	return fmt.Sprint(id2, "_", id1)
}

func getMatchResult(won, lost int) int {
	if won > lost {
		return 1
	}
	return -1
}

func recordResult(matchResults map[string](map[string]int), id string, category, result int) error {
	if val, ok := matchResults[strconv.Itoa(category)]; ok {
		val[id] += result
		return nil
	}
	return fmt.Errorf("Unknown category id: %d", category)
}

func mkVote(matchResults map[string](map[string]int)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var voteResults voteResult
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &voteResults)

		matchID := getMatchID(voteResults.WonID, voteResults.LostID)
		matchResult := getMatchResult(voteResults.WonID, voteResults.LostID)

		err := recordResult(matchResults, matchID, voteResults.CategoryID, matchResult)
		fmt.Fprintf(w, `{"success":%s, "msg":"%s"}`, strconv.FormatBool(err == nil), err.Error())
	}
}

func mkEngineers(people []person, categories []category) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cat := rand.Int() % len(categories)
		i1 := rand.Int() % len(people)
		i2 := rand.Int() % len(people)
		for i2 == i1 {
			i2 = rand.Int() % len(people)
		}
		response := peopleResponse{
			categories[cat],
			people[i1],
			people[i2],
			true, "",
		}
		data, _ := json.Marshal(response)
		fmt.Fprintf(w, string(data))
	}
}

func html(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func getEngineers() []person {
	var people []person
	data, _ := ioutil.ReadFile("people.json")
	json.Unmarshal(data, &people)

	return people
}

func getCategories() []category {
	var categories []category
	data, _ := ioutil.ReadFile("categories.json")
	json.Unmarshal(data, &categories)

	return categories
}

func getMatchResults() map[string](map[string]int) {
	matches := make(map[string](map[string]int))
	data, _ := ioutil.ReadFile("matchResult.json")
	json.Unmarshal(data, &matches)

	return matches
}

func main() {
	engineers := getEngineers()
	categories := getCategories()
	matchResults := getMatchResults()

	http.HandleFunc("/people", mkEngineers(engineers, categories))
	http.HandleFunc("/vote", mkVote(matchResults))
	http.HandleFunc("/", html)
	http.ListenAndServe(":8080", nil)
}

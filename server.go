package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	raven "github.com/DaveDuck321/RavenAuthenticationGo"
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
	Name string   `json:"nickname"`
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

func permissionDenied(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(403)
	http.ServeFile(w, r, "errors/forbidden.html")
}

func mkVote(matchResults map[string](map[string]int)) func(raven.RavenIdentity, http.ResponseWriter, *http.Request) {
	return func(identity raven.RavenIdentity, w http.ResponseWriter, r *http.Request) {
		var voteResults voteResult
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &voteResults)
		matchID := getMatchID(voteResults.WonID, voteResults.LostID)
		matchResult := getMatchResult(voteResults.WonID, voteResults.LostID)

		err := recordResult(matchResults, matchID, voteResults.CategoryID, matchResult)
		if err != nil {
			fmt.Fprintf(w, `{"success":false, "msg":"%s"}`, err.Error())
			return
		}
		fmt.Fprintf(w, `{"success":true, "msg":""}`)
	}
}

func mkEngineers(people []person, categories []category) func(raven.RavenIdentity, http.ResponseWriter, *http.Request) {
	return func(identity raven.RavenIdentity, w http.ResponseWriter, r *http.Request) {
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

func mkRedirect(url string) func(identity raven.RavenIdentity, w http.ResponseWriter, r *http.Request) {
	return func(identity raven.RavenIdentity, w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, 300)
	}
}

func html(identity raven.RavenIdentity, w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

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

func getMatchResults() map[string](map[string]int) {
	matches := make(map[string](map[string]int))
	data, _ := ioutil.ReadFile("data/matchResult.json")
	json.Unmarshal(data, &matches)

	return matches
}

func saveMatchResults(saveJson *time.Ticker, matchResults map[string](map[string]int)) {
	for {
		<-saveJson.C
		data, _ := json.Marshal(matchResults)
		ioutil.WriteFile("data/matchResult.json", data, 0644)
	}
}

func main() {
	engineers := getEngineers()
	categories := getCategories()
matchResults:
	saveMatchResultsewAuthenticator()

	saveJson := time.NewTicker(time.Second)
	go saveMatchResults(saveJson, matchResults)

	auth.HandleRavenAuthenticator("/auth/raven", mkRedirect("/"), permissionDenied)
	auth.AuthoriseAndHandle("/people", mkEngineers(engineers, categories), permissionDenied)
	auth.AuthoriseAndHandle("/vote", mkVote(matchResults), permissionDenied)
	auth.AuthoriseAndHandle("/", html, permissionDenied)
	fmt.Println(http.ListenAndServe(":80", nil))
}

/*
https://raven.cam.ac.uk/auth/authenticate.html
?
ver=3&
url=http%3a%2f%2flocalhost%2fauth%2fraven&
date=20200124T165401Z&
iact=yes&
*/

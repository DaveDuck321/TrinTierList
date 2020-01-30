package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
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

func permissionDenied(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(403)
	http.ServeFile(w, r, "errors/forbidden.html")
}

func mkVote(rankings rankings) func(raven.RavenIdentity, http.ResponseWriter, *http.Request) {
	return func(identity raven.RavenIdentity, w http.ResponseWriter, r *http.Request) {
		var voteResults voteResult
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &voteResults)

		err := updateRankings(rankings, voteResults.WonID, voteResults.LostID, voteResults.CategoryID, 60)

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

func saveMatchResults(saveJSON *time.Ticker, matchResults rankings) {
	for {
		<-saveJSON.C
		data, _ := json.Marshal(matchResults)
		ioutil.WriteFile("data/matchResult.json", data, 0644)
	}
}

func main() {
	engineers := getEngineers()
	categories := getCategories()
	matchResults := getRankingsResults(engineers, categories)

	saveJSON := time.NewTicker(time.Second)
	go saveMatchResults(saveJSON, matchResults)

	auth := raven.NewAuthenticator()
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

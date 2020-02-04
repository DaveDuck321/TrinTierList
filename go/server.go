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
	ID    int      `json:"id"`
	CrsID string   `json:"crsID"`
	Name  string   `json:"nickname"`
	Imgs  []string `json:"imgs"`

	ELO int `json:"elo"`
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

type peopleRequest struct {
	CategoryChoice int `json:"category"`
}

type leaderboardResponse struct {
	People     []person   `json:"people"`
	Categories []category `json:"categories"`
	Rankings   rankings   `json:"elos"`
}

//Maps user identity and category to remaining votes
type availableVotes map[int]([]int)
type allAvailableVotes map[string]availableVotes

func permissionDenied(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(403)
	http.ServeFile(w, r, "errors/forbidden.html")
}

// TODO: ELO change
func mkVote(rankings rankings, allVotes allAvailableVotes) func(raven.Identity, http.ResponseWriter, *http.Request) {
	return func(identity raven.Identity, w http.ResponseWriter, r *http.Request) {
		var voteResults voteResult
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &voteResults)

		votes, ok := allVotes[identity.CrsID][voteResults.CategoryID]
		if !ok {
			fmt.Fprintf(w, `{"success":false, "msg":"CRSID or category invaild"}`)
			return
		}

		if votes[0] != getMatchID(voteResults.WonID, voteResults.LostID) {
			fmt.Fprintf(w, `{"success":false, "msg":"Attempted to vote twice"}`)
			return
		}
		allVotes[identity.CrsID][voteResults.CategoryID] = votes[1:]

		err := updateRankings(rankings, voteResults.WonID, voteResults.LostID, voteResults.CategoryID, 60)

		if err != nil {
			fmt.Fprintf(w, `{"success":false, "msg":"%s"}`, err.Error())
			return
		}
		fmt.Fprintf(w, `{"success":true, "msg":""}`)
	}
}

func mkLeaderboard(people []person, categories []category, ranks rankings) func(raven.Identity, http.ResponseWriter, *http.Request) {
	return func(identity raven.Identity, w http.ResponseWriter, r *http.Request) {
		responseObj := leaderboardResponse{
			people, categories, ranks,
		}
		resp, _ := json.Marshal(responseObj)
		fmt.Fprintf(w, string(resp))
	}
}

func mkEngineers(ranks rankings, votes allAvailableVotes, peopleMap map[int]person, categories []category) func(raven.Identity, http.ResponseWriter, *http.Request) {
	categoryMap := make(map[int]category)
	for _, category := range categories {
		categoryMap[category.ID] = category
	}

	return func(identity raven.Identity, w http.ResponseWriter, r *http.Request) {
		var request peopleRequest
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &request)

		cat, ok := categoryMap[request.CategoryChoice]
		if !ok {
			//Choose random category if user's choice is wrong
			cat = categories[rand.Int()%len(categories)]
		}
		votesLeft := votes[identity.CrsID][cat.ID]

		p1ID, p2ID, err := nextMatch(votesLeft)
		if err != nil {
			fmt.Fprintf(w, `{"success":false, "msg":"%s"}`, err.Error())
			return
		}
		p1, p2, err := getPeopleFromIDs(peopleMap, p1ID, p2ID)
		if err != nil {
			fmt.Fprintf(w, `{"success":false, "msg":"%s"}`, err.Error())
			return
		}

		p1.ELO = ranks[cat.ID][p1ID]
		p2.ELO = ranks[cat.ID][p2ID]
		response := peopleResponse{
			cat,
			p1, p2,
			true, "",
		}
		data, _ := json.Marshal(response)
		fmt.Fprintf(w, string(data))
	}
}

func mkRedirect(url string) func(identity raven.Identity, w http.ResponseWriter, r *http.Request) {
	return func(identity raven.Identity, w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, 302)
	}
}

func html(identity raven.Identity, w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func saveMatchResults(saveJSON *time.Ticker, matchResults rankings) {
	for {
		<-saveJSON.C
		data, _ := json.Marshal(matchResults)
		ioutil.WriteFile("data/matchResult.json", data, 0644)
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
	return votes
}

func main() {
	people := getPeople()
	peopleMap := makePeopleMap(people)
	categories := getCategories()
	rankings := getRankingsResults(people, categories)

	matchesRemaining := updateAvailableVotes(make(allAvailableVotes), people, categories)

	saveJSON := time.NewTicker(time.Second)
	go saveMatchResults(saveJSON, rankings)

	auth := raven.NewAuthenticator("http", "localhost", "./keys/pubkey2")
	auth.HandleAuthenticationPath("/auth/raven", mkRedirect("/"), permissionDenied)
	auth.AuthoriseAndHandle("/people", mkEngineers(rankings, matchesRemaining, peopleMap, categories), permissionDenied)
	auth.AuthoriseAndHandle("/leaderboard", mkLeaderboard(people, categories, rankings), permissionDenied)
	auth.AuthoriseAndHandle("/vote", mkVote(rankings, matchesRemaining), permissionDenied)
	auth.AuthoriseAndHandle("/", html, permissionDenied)

	fmt.Println("Listening at port 80...")
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

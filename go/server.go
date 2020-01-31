package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

//Maps user identity and category to remaining votes
type availableVotes map[int]([]int)
type allAvailableVotes map[string]availableVotes

func permissionDenied(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(403)
	http.ServeFile(w, r, "errors/forbidden.html")
}

func mkVote(rankings rankings, allVotes allAvailableVotes) func(raven.RavenIdentity, http.ResponseWriter, *http.Request) {
	return func(identity raven.RavenIdentity, w http.ResponseWriter, r *http.Request) {
		var voteResults voteResult
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &voteResults)

		votes, ok := allVotes[identity.CrsID][voteResults.CategoryID]
		if !ok {
			fmt.Fprintf(w, `{"success":false, "msg":"CRSID or category invaild"}`)
			return
		}
		matchIndex, err := indexOfMatchID(votes, getMatchID(voteResults.WonID, voteResults.LostID))
		if err != nil {
			fmt.Fprintf(w, `{"success":false, "msg":"Attempted to vote twice"}`)
			return
		}
		votes[matchIndex] = votes[0]
		allVotes[identity.CrsID][voteResults.CategoryID] = votes[1:]

		err = updateRankings(rankings, voteResults.WonID, voteResults.LostID, voteResults.CategoryID, 60)

		if err != nil {
			fmt.Fprintf(w, `{"success":false, "msg":"%s"}`, err.Error())
			return
		}
		fmt.Fprintf(w, `{"success":true, "msg":""}`)
	}
}

func mkEngineers(ranks rankings, votes allAvailableVotes, people []person, categories []category) func(raven.RavenIdentity, http.ResponseWriter, *http.Request) {
	peopleMap := make(map[int]person)
	for _, person := range people {
		peopleMap[person.ID] = person
	}

	return func(identity raven.RavenIdentity, w http.ResponseWriter, r *http.Request) {
		//cat := categories[rand.Int()%len(categories)]
		cat := categories[1]
		votesLeft := votes[identity.CrsID][cat.ID]

		p1ID, p2ID, err := randomMatch(votesLeft)
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

func updateAvailableVotes(votes allAvailableVotes, people []person, categories []category) allAvailableVotes {
	allMatches := genPermutations(people)

	for _, person := range people {
		votes[person.CrsID] = make(availableVotes)
		for _, category := range categories {
			votes[person.CrsID][category.ID] = make([]int, len(allMatches))
			copy(votes[person.CrsID][category.ID], allMatches)
		}
	}
	return votes
}

func main() {
	engineers := getEngineers()
	categories := getCategories()
	rankings := getRankingsResults(engineers, categories)

	matchesRemaining := updateAvailableVotes(make(allAvailableVotes), engineers, categories)

	saveJSON := time.NewTicker(time.Second)
	go saveMatchResults(saveJSON, rankings)

	auth := raven.NewAuthenticator()
	auth.HandleRavenAuthenticator("/auth/raven", mkRedirect("/"), permissionDenied)
	auth.AuthoriseAndHandle("/people", mkEngineers(rankings, matchesRemaining, engineers, categories), permissionDenied)
	auth.AuthoriseAndHandle("/vote", mkVote(rankings, matchesRemaining), permissionDenied)
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

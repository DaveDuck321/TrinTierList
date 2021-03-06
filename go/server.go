package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
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

type categoryChoice struct {
	CategoryChoice int `json:"category"`
}

type leaderboardResponse struct {
	Success    bool       `json:"success"`
	People     []person   `json:"people"`
	Categories []category `json:"categories"`
	Rankings   rankings   `json:"elos"`
}

type rankings map[int](map[int]int)

//Maps user identity and category to remaining votes
type availableVotes map[int]([]int)
type allAvailableVotes map[string]availableVotes

func permissionDenied(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Println("User failed to login: ", err)
	http.Redirect(w, r, "/forbidden", 302)
}

func staticServe(identity raven.Identity, w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "www"+r.URL.EscapedPath())
}

func mkServeForbidden(auth raven.Authenticator) func(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("").ParseFiles("www/error/forbidden.html")

	return func(w http.ResponseWriter, r *http.Request) {
		t.ExecuteTemplate(w, "Forbidden", auth.GetRavenLink("/auth/raven"))
	}
}

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

		eloDelta, err := updateRankings(rankings, voteResults.WonID, voteResults.LostID, voteResults.CategoryID, 60)

		if err != nil {
			fmt.Fprintf(w, `{"success":false, "msg":"%s"}`, err.Error())
			return
		}
		fmt.Fprintf(w, `{"success":true, "msg":"", "elo_change":{"winner":%d, "looser":%d}}`, eloDelta, -eloDelta)
	}
}

func mkLeaderboard(people []person, categories []category, ranks rankings) func(raven.Identity, http.ResponseWriter, *http.Request) {
	return func(identity raven.Identity, w http.ResponseWriter, r *http.Request) {
		responseObj := leaderboardResponse{
			true, people, categories, ranks,
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
		var request categoryChoice
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

func main() {
	hostname, port := splitHostPort(os.Args[1])

	people := getPeople()
	peopleMap := makePeopleMap(people)
	categories := getCategories()
	rankings := getRankingsResults(people, categories)

	matchesRemaining := updateAvailableVotes(make(allAvailableVotes), people, categories)

	go saveELOsInterval(rankings, time.NewTicker(12*time.Hour))
	go refreshTempDataInterval(matchesRemaining, people, categories, time.NewTicker(48*time.Hour))

	auth := raven.NewAuthenticator("http", hostname, "./keys/pubkey2")

	auth.HandleAuthenticationPath("/auth/raven", mkRedirect("/"), permissionDenied)
	auth.AuthoriseAndHandle("/api/match", mkEngineers(rankings, matchesRemaining, peopleMap, categories), permissionDenied)
	auth.AuthoriseAndHandle("/api/leaderboard", mkLeaderboard(people, categories, rankings), permissionDenied)
	auth.AuthoriseAndHandle("/api/vote", mkVote(rankings, matchesRemaining), permissionDenied)

	http.HandleFunc("/forbidden", mkServeForbidden(auth))
	auth.AuthoriseAndHandle("/", staticServe, permissionDenied)

	fmt.Println("Listening at port", port)
	fmt.Println(http.ListenAndServe(port, nil))
}

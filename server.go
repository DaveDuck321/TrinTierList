package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
)

type peopleResponse struct {
	Category category `json:"category"`
	Person1  person   `json:"person1"`
	Person2  person   `json:"person2"`
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

func vote(w http.ResponseWriter, r *http.Request) {

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

func main() {
	engineers := getEngineers()
	categories := getCategories()
	fmt.Println(engineers)
	fmt.Println(categories)

	http.HandleFunc("/people", mkEngineers(engineers, categories))
	http.HandleFunc("/vote", vote)
	http.HandleFunc("/", html)
	http.ListenAndServe(":8080", nil)
}

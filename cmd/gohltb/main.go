package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fuzzylimes/gohltb"
)

var gameTitle string
var sortBy string
var details bool
var randomGame bool

func init() {
	flag.StringVar(&gameTitle, "g", "", "Game title to query for")
	flag.StringVar(&sortBy, "s", "name", "How the response should be sorted. Sorts by name by default. Supports: name, main, mainp, comp, averagea, rating, popular, backlog, usersp, playing, speedruns, release")
	flag.BoolVar(&details, "d", false, "Include additional user details about the games.")
	flag.BoolVar(&randomGame, "r", false, "Return a single, random, game based on query.")
	flag.Parse()
}

func main() {
	var d string

	if details {
		d = "user_stats"
	}

	q := &gohltb.HLTBQuery{
		Query:    gameTitle,
		SortBy:   gohltb.SortBy(sortBy),
		Modifier: gohltb.Modifier(d),
		Random:   randomGame,
	}

	client := gohltb.NewDefaultClient()

	games, err := client.SearchGamesByQuery(q)
	if err != nil {
		log.Fatal(err)
	}

	j, err := games.JSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(j)

}

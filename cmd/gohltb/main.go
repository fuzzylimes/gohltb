package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fuzzylimes/gohltb"
)

var query string
var sortBy string
var details bool
var user bool
var randomGame bool

func init() {
	flag.StringVar(&query, "q", "", "Query string. This will be the game title if searching for games, or user name if searching for users.")
	flag.BoolVar(&user, "u", false, "Query users instead of games")
	flag.StringVar(&sortBy, "s", "name", "How the response should be sorted. Sorts by name by default.\nGames support: name, main, mainp, comp, averagea, rating, popular, backlog, usersp, playing, speedruns, release\nUsers support: name, gender, postcount, numcomp, numbacklog")
	flag.BoolVar(&details, "d", false, "Include additional user details when querying games.")
	flag.BoolVar(&randomGame, "r", false, "Return a single, random, game or user.")
	flag.Parse()
}

func main() {
	var d string
	var mode gohltb.QueryType
	if user {
		mode = gohltb.UserQuery
	} else {
		mode = gohltb.GameQuery
	}

	if !isValidSort(sortBy, mode) {
		flag.Usage()
		log.Fatalln("Invalid sortBy parameter")
	}
	if details {
		d = "user_stats"
	}

	q := &gohltb.HLTBQuery{
		Query:     query,
		QueryType: mode,
		SortBy:    gohltb.SortBy(sortBy),
		Modifier:  gohltb.Modifier(d),
		Random:    randomGame,
	}

	client := gohltb.NewDefaultClient()

	if user {
		games, err := client.SearchUsersByQuery(q)
		if err != nil {
			log.Fatal(err)
		}

		j, err := games.JSON()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(j)
	} else {
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

}

func isValidSort(sort string, t gohltb.QueryType) bool {
	games := []string{"name", "main", "mainp", "comp", "averagea", "rating", "popular", "backlog", "usersp", "playing", "speedruns", "release"}
	users := []string{"name", "gender", "postcount", "numcomp", "numbacklog"}
	switch t {
	case gohltb.UserQuery:
		return contains(users, sort)
	case gohltb.GameQuery:
		return contains(games, sort)
	}
	return false
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fuzzylimes/gohltb"
)

var queryVar string

func init() {
	flag.StringVar(&queryVar, "g", "", "game title to query for")
	flag.Parse()
}

func main() {
	games, err := gohltb.SearchGame(queryVar)
	if err != nil {
		log.Fatal(err)
	}
	games, err = games.GetNextPage()
	games, err = games.GetNextPage()
	j, err := games.JSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(j)

}

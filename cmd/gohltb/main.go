package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fuzzylimes/gohltb"
	"github.com/fuzzylimes/gohltb/query"
)

var queryVar string

func init() {
	flag.StringVar(&queryVar, "g", "", "game title to query for")
	flag.Parse()
}

func main() {
	if queryVar == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	games, err := gohltb.SearchGame(queryVar)
	if err != nil {
		log.Fatal(err)
	}
	j, err := games.JSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(j)

}

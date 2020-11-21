package gohltb

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	// QueryURL is the endpoint for queries
	QueryURL string = "https://howlongtobeat.com/search_results?page=1"
)

// GameTimes are the times it takes to complete a game
type GameTimes struct {
	Title         string `json:"title"`
	Main          string `json:"main"`
	MainExtra     string `json:"main-extra"`
	Completionist string `json:"completionist"`
}

// ManyGameTimes is a list of GameTimes
type ManyGameTimes struct {
	Games []*GameTimes `json:"games"`
}

// JSON will convert a ManyGameTImes object into a json stirng
func (m ManyGameTimes) JSON() (string, error) {
	var r string
	s, err := json.MarshalIndent(m.Games, "", "\t")
	if err == nil {
		r = string(s)
	}
	return r, err
}

// SearchGame performs a search for the provided game title
func SearchGame(query string) (*ManyGameTimes, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	form := url.Values{
		"t":           {"games"},
		"queryString": {query},
	}

	req, err := http.NewRequest("POST", QueryURL, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	return parseResponse(resp)
}

// parseResponse converts the http response object into a ManyGameTimes object
func parseResponse(r *http.Response) (*ManyGameTimes, error) {
	games := &ManyGameTimes{}
	var gameslice []*GameTimes
	// Should eventually handle 404 vs 400 vs 500 etc.
	if r.StatusCode != 200 {
		return games, errors.New("Error retrieving data")
	}

	doc, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		return games, err
	}

	if strings.Contains(doc.Nodes[0].Data, "No results") {
		return games, nil
	}

	doc.Find(".search_list_details").Each(func(gameCount int, gameDetails *goquery.Selection) {
		times := gameDetails.Find(".search_list_tidbit.center")
		game := &GameTimes{
			Title:         sanitizeName(gameDetails.Find("h3 > a").Text()),
			Main:          strings.TrimSpace(times.Nodes[0].FirstChild.Data),
			MainExtra:     strings.TrimSpace(times.Nodes[1].FirstChild.Data),
			Completionist: strings.TrimSpace(times.Nodes[2].FirstChild.Data),
		}

		gameslice = append(gameslice, game)
	})

	games.Games = gameslice
	return games, nil

}

// sanitizeName will remove any unexpected characters from a game title
func sanitizeName(s string) string {
	regx := regexp.MustCompile(`[\s]{2,}`)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = regx.ReplaceAllString(s, " ")
	return s
}

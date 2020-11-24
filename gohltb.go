package gohltb

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var client *http.Client

func init() {
	client = &http.Client{
		Timeout: time.Second * 10,
	}
}

const (
	// queryURL is the endpoint for queries
	queryURL string = "https://howlongtobeat.com"
	// urlPrefix is the base URL
	urlPrefix string = "https://howlongtobeat.com/"
)

// GameTimes are the times it takes to complete a game
type GameTimes struct {
	ID            string            `json:"id"`                   // ID in howlongtobeat.com's database
	Title         string            `json:"title"`                // Title of game
	URL           string            `json:"url"`                  // Link to game page
	BoxArtURL     string            `json:"box-art-url"`          // Link to boxart
	Main          string            `json:"main"`                 // Completion time for Main category
	MainExtra     string            `json:"main-extra"`           // Completion time for Main + Extra category
	Completionist string            `json:"completionist"`        // Completion time for Completionist category
	Other         map[string]string `json:"other,omitempty"`      // Times that don't fall under the main 3 time categories
	UserStats     *UserStats        `json:"user-stats,omitempty"` // Optional additional user details (only present when requested)
}

// UserStats are additional stats based on user activity on howlongtobeat.com
// This information is only included when additional user details are requested
type UserStats struct {
	Completed string `json:"completed,omitempty"` // Number of users that have marked the game as Completed
	Rating    string `json:"rating,omitempty"`    // Average User rating
	Backlog   string `json:"backlog,omitempty"`   // Number of users that have placed game in their backlog
	Playing   string `json:"playing,omitempty"`   // Number of users that are currently playing through a game
	Retired   string `json:"retired,omitempty"`   // Number of users that did not finish a game
	SpeedRuns string `json:"speedruns,omitempty"` // Number of users that have submitted speedruns for the game
}

// GameTimesPage is a page of responses
type GameTimesPage struct {
	Games        []*GameTimes `json:"games"`       // Slice of GameTimes
	TotalPages   int          `json:"total-pages"` // Total number of pages
	CurrentPage  int          `json:"curent-page"` // Current page number
	NextPage     int          `json:"next-page"`   // Next page number
	TotalMatches int          `json:"matches"`     // Total query matches
	requestQuery *HLTBQuery   // Query associated with this response
}

// HLTBQuery is the query to be run
type HLTBQuery struct {
	Query         string // String ot query by
	QueryType     QueryType
	SortBy        SortBy        // Specify how data should be sorted
	SortDirection SortDirection // Specify direction data should be sorted
	Platform      Platform
	LengthType    LengthRange
	LengthMin     string
	LengthMax     string
	Modifier      Modifier
	Random        bool
	Page          int
}

// HasNext will check to see if there is a next page that can be retrieved
func (g *GameTimesPage) HasNext() bool {
	return g.NextPage != 0
}

// GetNextPage will return the next page, if present
func (g *GameTimesPage) GetNextPage() (*GameTimesPage, error) {
	if !g.HasNext() {
		return &GameTimesPage{}, errors.New("Page not found")
	}
	query := g.requestQuery
	query.Page = g.NextPage
	return SearchByQuery(query)
}

// JSON will convert a ManyGameTImes object into a json string
func (g *GameTimesPage) JSON() (string, error) {
	var r string
	s, err := json.MarshalIndent(g.Games, "", "  ")
	if err == nil {
		r = string(s)
	}
	return r, err
}

// SearchGame performs a search for the provided game title
func SearchGame(query string) (*GameTimesPage, error) {
	return handleSearch(queryURL, &HLTBQuery{Query: query})
}

// SearchByQuery queries using a set of user defined parameters
func SearchByQuery(q *HLTBQuery) (*GameTimesPage, error) {
	return handleSearch(queryURL, q)
}

func handleSearch(qURL string, q *HLTBQuery) (*GameTimesPage, error) {
	handleDefaults(q)
	form := buildForm(q)

	req, err := http.NewRequest("POST", fmt.Sprintf("%v/search_results?page=%v", qURL, q.Page), strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	pages, err := parseResponse(resp, q)
	if err != nil {
		return pages, err
	}

	return pages, nil
}

func buildForm(q *HLTBQuery) url.Values {
	// Handle the Random parameter string
	var random string
	if q.Random {
		random = "1"
	} else {
		random = "0"
	}

	// build out the query form using expected query values
	return url.Values{
		"queryString": {q.Query},
		"t":           {string(q.QueryType)},
		"sorthead":    {string(q.SortBy)},
		"sortd":       {string(q.SortDirection)},
		"plat":        {string(q.Platform)},
		"length_type": {string(q.LengthType)},
		"length_min":  {q.LengthMin},
		"length_max":  {q.LengthMax},
		"detail":      {string(q.Modifier)},
		"randomize":   {random},
	}
}

func handleDefaults(q *HLTBQuery) {
	if q.QueryType == "" {
		q.QueryType = GameQuery
	}
	if q.SortBy == "" {
		q.SortBy = MostPopular
	}
	if q.SortDirection == "" {
		q.SortDirection = NormalOrder
	}
	if q.LengthType == "" {
		q.LengthType = RangeMainStory
	}
	if q.Page == 0 {
		q.Page = 1
	}
}

// parseResponse converts the http response object into a ManyGameTimes object
func parseResponse(r *http.Response, q *HLTBQuery) (*GameTimesPage, error) {
	games := &GameTimesPage{}
	var gameslice []*GameTimes
	// Should eventually handle 404 vs 400 vs 500 etc.
	if r.StatusCode != 200 {
		return games, errors.New("Error retrieving data")
	}

	doc, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		return games, err
	}

	if strings.Contains(doc.Find("li").Nodes[0].FirstChild.Data, "No results") {
		return games, nil
	}

	// Handle the page numbers
	parsePages(doc, games, q.Page)

	doc.Find("ul > li.back_darkish").Each(func(gameCount int, gameDetails *goquery.Selection) {
		boxArt, _ := gameDetails.Find("div > a > img").Attr("src")
		title := gameDetails.Find("h3 > a")
		url, _ := title.Attr("href")
		id := strings.SplitAfter(url, "game?id=")[1]

		game := &GameTimes{
			ID:        id,
			URL:       urlPrefix + url,
			BoxArtURL: boxArt,
			Title:     sanitizeName(title.Text()),
		}
		userStats := &UserStats{}

		if gameDetails.Find(".search_list_details_block").First().Children().First().Is(".search_list_tidbit_short") {
			otherMap := make(map[string]string)
			gameDetails.Find(".search_list_tidbit_short").Each(func(timeCount int, timeDetail *goquery.Selection) {
				otherMap[timeDetail.Text()] = strings.TrimSpace(timeDetail.Next().Text())
			})
			game.Other = otherMap
		} else {
			gameDetails.Find(".search_list_tidbit.text_white").Each(func(timeCount int, timeDetail *goquery.Selection) {
				nextValue := strings.TrimSpace(timeDetail.Next().Text())
				switch timeType := timeDetail.Text(); {
				case timeType == "Main Story":
					game.Main = nextValue
				case timeType == "Main + Extra":
					game.MainExtra = nextValue
				case timeType == "Completionist":
					game.Completionist = nextValue
				case timeType == "Polled":
					userStats.Completed = nextValue
				case timeType == "Rated":
					userStats.Rating = nextValue
				case timeType == "Backlog":
					userStats.Backlog = nextValue
				case timeType == "Playing":
					userStats.Playing = nextValue
				case timeType == "Speedruns":
					userStats.SpeedRuns = nextValue
				case timeType == "Retired":
					userStats.Retired = nextValue
				}
			})
		}
		if q.Modifier == ShowUserStats {
			game.UserStats = userStats
		}
		gameslice = append(gameslice, game)
	})
	games.Games = gameslice
	games.CurrentPage = q.Page
	games.requestQuery = q
	if games.TotalPages > games.CurrentPage {
		games.NextPage = games.CurrentPage + 1
	}

	return games, nil

}

func parsePages(doc *goquery.Document, games *GameTimesPage, p int) {
	if p == 1 {
		regex := regexp.MustCompile(`Found ([1-9]+) Game`)
		numGames := regex.FindStringSubmatch(doc.Find("h3").Nodes[0].FirstChild.Data)[1]
		numGamesInt, err := strconv.Atoi(numGames)
		if err != nil {
			games.TotalMatches = 0
		} else {
			games.TotalMatches = numGamesInt
		}
	}

	lastPage, err := strconv.Atoi(doc.Find("span.search_list_page").Last().Text())
	if err != nil {
		games.TotalPages = 1
	} else {
		games.TotalPages = lastPage
	}
}

// sanitizeName will remove any unexpected characters from a game title
func sanitizeName(s string) string {
	regx := regexp.MustCompile(`[\s]{2,}`)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = regx.ReplaceAllString(s, " ")
	return s
}

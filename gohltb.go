package gohltb

import (
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

// ResultPage ...
type ResultPage interface {
	HasNext() bool
	GetNextPage() (ResultPage, error)
	JSON() (string, error)
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

func (h *HLTBQuery) gameQuery() {
	h.QueryType = GameQuery
}

func (h *HLTBQuery) userQuery() {
	h.QueryType = UserQuery
}

// SearchGames performs a search for the provided game title
func SearchGames(query string) (*GameResultsPage, error) {
	res, err := handleSearch(queryURL, &HLTBQuery{Query: query})
	return res.(*GameResultsPage), err
}

// SearchGamesByQuery queries using a set of user defined parameters
func SearchGamesByQuery(q *HLTBQuery) (*GameResultsPage, error) {
	q.gameQuery()
	res, err := handleSearch(queryURL, q)
	return res.(*GameResultsPage), err
}

// SearchUsers performs a search for the provided user name
func SearchUsers(query string) (*UserResultsPage, error) {
	res, err := handleSearch(queryURL, &HLTBQuery{Query: query, QueryType: UserQuery})
	return res.(*UserResultsPage), err
}

// SearchUsersByQuery queries using a set of user defined parameters
func SearchUsersByQuery(q *HLTBQuery) (*UserResultsPage, error) {
	q.userQuery()
	res, err := handleSearch(queryURL, q)
	return res.(*UserResultsPage), err
}

func handleSearch(qURL string, q *HLTBQuery) (ResultPage, error) {
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

// buildForm will construct the form payload sent to the server.
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

// handleDefaults will set all of the default form values for the query. It
// handles the cases where user has not provided values where they should be
// included
func handleDefaults(q *HLTBQuery) {
	if q.QueryType == "" {
		q.QueryType = GameQuery
	}
	if q.SortBy == "" {
		q.SortBy = SortByGameName
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
func parseResponse(r *http.Response, q *HLTBQuery) (*GameResultsPage, error) {
	games := &GameResultsPage{}
	var gameslice []*GameResult
	// Should eventually handle 404 vs 400 vs 500 etc.
	if r.StatusCode != 200 {
		return nil, errors.New("Error retrieving data")
	}

	doc, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		return nil, err
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

		game := &GameResult{
			ID:        id,
			URL:       urlPrefix + url,
			BoxArtURL: boxArt,
			Title:     sanitizeTitle(title.Text()),
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

func parsePages(doc *goquery.Document, games *GameResultsPage, p int) {
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

// sanitizeTitle will remove any unexpected characters from a game title
func sanitizeTitle(s string) string {
	regx := regexp.MustCompile(`[\s]{2,}`)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = regx.ReplaceAllString(s, " ")
	return s
}

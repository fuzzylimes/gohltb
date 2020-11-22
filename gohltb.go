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

	"github.com/fuzzylimes/gohltb/query"

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
	queryURL string = "https://howlongtobeat.com/search_results?page="
	// urlPrefix is the base URL
	urlPrefix string = "https://howlongtobeat.com/"
)

// GameTimes are the times it takes to complete a game
type GameTimes struct {
	ID            string    `json:"id"`            // ID in howlongtobeat.com's database
	Title         string    `json:"title"`         // Title of game
	URL           string    `json:"url"`           // Link to game page
	BoxArtURL     string    `json:"box-art-url"`   // Link to boxart
	Main          string    `json:"main"`          // Completion time for Main category
	MainExtra     string    `json:"main-extra"`    // Completion time for Main + Extra category
	Completionist string    `json:"completionist"` // Completion time for Completionist category
	UserStats     UserStats `json:"user-stats"`    // Optional additional user details (only present when requested)
}

// UserStats are additional stats based on user activity on howlongtobeat.com
// This information is only included when additional user details are requested
type UserStats struct {
	Completed string `json:"completed"` // Number of users that have marked the game as Completed
	Rating    string `json:"rating"`    // Average User rating
	Backlog   string `json:"backlog"`   // Number of users that have placed game in their backlog
	Playing   string `json:"playing"`   // Number of users that are currently playing through a game
	Retired   string `json:"retired"`   // Number of users that did not finish a game
}

// GameTimesPage is a page of responses
type GameTimesPage struct {
	Games        []*GameTimes `json:"games"`       // Slice of GameTimes
	TotalPages   int          `json:"total-pages"` // Total number of pages
	CurrentPage  int          `json:"curent-page"` // Current page number
	NextPage     int          `json:"next-page"`   // Next page number
	TotalMatches int          `json:"matches"`     // Total query matches
	RequestQuery *HLTBQuery   // Query associated with this response
}

// HLTBQuery is the query to be run
type HLTBQuery struct {
	Query         string // String ot query by
	QueryType     query.QueryType
	SortBy        query.SortBy        // Specify how data should be sorted
	SortDirection query.SortDirection // Specify direction data should be sorted
	Platform      query.Platform
	LengthType    query.LengthRange
	LengthMin     string
	LengthMax     string
	Modifier      query.Modifier
	Random        bool
	Page          int
}

func (g GameTimesPage) hasNext() bool {
	return g.NextPage != 0
}

// JSON will convert a ManyGameTImes object into a json stirng
func (g GameTimesPage) JSON() (string, error) {
	var r string
	s, err := json.MarshalIndent(g.Games, "", "\t")
	if err == nil {
		r = string(s)
	}
	return r, err
}

// SearchGame performs a search for the provided game title
func SearchGame(query string) (*GameTimesPage, error) {
	return handleSearch(fmt.Sprintf("%v%v", queryURL, "1"), &HLTBQuery{Query: query})
}

// SearchByQuery queries using a set of user defined parameters
func SearchByQuery(q *HLTBQuery) (*GameTimesPage, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	return handleSearch(fmt.Sprintf("%v%v", queryURL, q.Page), q)

}

func handleSearch(qURL string, q *HLTBQuery) (*GameTimesPage, error) {
	err := handleDefaults(q)
	if err != nil {
		return &GameTimesPage{}, err
	}

	form := buildForm(q)

	req, err := http.NewRequest("POST", qURL, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	pages, err := parseResponse(resp)
	if err != nil {
		return pages, err
	}
	pages.CurrentPage = q.Page
	pages.RequestQuery = q
	if pages.TotalPages > pages.CurrentPage {
		pages.NextPage = pages.CurrentPage + 1
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

func handleDefaults(q *HLTBQuery) error {
	if q.QueryType == "" {
		q.QueryType = query.GameQuery
	}
	if q.SortBy == "" {
		q.SortBy = query.MostPopular
	}
	if q.SortDirection == "" {
		q.SortDirection = query.NormalOrder
	}
	if q.LengthType == "" {
		q.LengthType = query.RangeMainStory
	}
	return nil
}

// parseResponse converts the http response object into a ManyGameTimes object
func parseResponse(r *http.Response) (*GameTimesPage, error) {
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
	parsePages(doc, games)

	doc.Find("ul > li.back_darkish").Each(func(gameCount int, gameDetails *goquery.Selection) {
		boxArt, _ := gameDetails.Find("div > a > img").Attr("src")
		title := gameDetails.Find("h3 > a")
		url, _ := title.Attr("href")
		id := strings.SplitAfter(url, "game?id=")[1]
		times := gameDetails.Find(".search_list_tidbit.center")
		game := &GameTimes{
			ID:            id,
			URL:           urlPrefix + url,
			BoxArtURL:     boxArt,
			Title:         sanitizeName(title.Text()),
			Main:          strings.TrimSpace(times.Nodes[0].FirstChild.Data),
			MainExtra:     strings.TrimSpace(times.Nodes[1].FirstChild.Data),
			Completionist: strings.TrimSpace(times.Nodes[2].FirstChild.Data),
		}

		gameslice = append(gameslice, game)
	})

	games.Games = gameslice
	return games, nil

}

func parsePages(doc *goquery.Document, games *GameTimesPage) {
	regex := regexp.MustCompile(`Found ([1-9]+) Game`)
	numGames := regex.FindStringSubmatch(doc.Find("h3").Nodes[0].FirstChild.Data)[1]
	numGamesInt, err := strconv.Atoi(numGames)
	if err != nil {
		games.TotalMatches = 0
	} else {
		games.TotalMatches = numGamesInt
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

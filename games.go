package gohltb

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GameResult are the times it takes to complete a game
type GameResult struct {
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

// GameResultsPage is a page of game responses
type GameResultsPage struct {
	Games        []*GameResult `json:"games"`       // Slice of GameResult
	TotalPages   int           `json:"total-pages"` // Total number of pages
	CurrentPage  int           `json:"curent-page"` // Current page number
	NextPage     int           `json:"next-page"`   // Next page number
	TotalMatches int           `json:"matches"`     // Total query matches
	requestQuery *HLTBQuery    // Query associated with this response
	hltbClient   *HLTBClient   // Client for further requests
}

// HasNext will check to see if there is a next page that can be retrieved
func (g *GameResultsPage) HasNext() bool {
	return g.NextPage != 0
}

// GetNextPage will return the next page, if present
func (g *GameResultsPage) GetNextPage() (*GameResultsPage, error) {
	if !g.HasNext() {
		return &GameResultsPage{}, errors.New("Page not found")
	}
	query := g.requestQuery
	query.Page = g.NextPage
	return g.hltbClient.SearchGamesByQuery(query)
}

// JSON will convert a game object into a json string
func (g *GameResultsPage) JSON() (string, error) {
	var r string
	s, err := json.MarshalIndent(g.Games, "", "  ")
	if err == nil {
		r = string(s)
	}
	return r, err
}

func (g *GameResultsPage) setTotalPages(p int) {
	g.TotalPages = p
}

func (g *GameResultsPage) setTotalMatches(m int) {
	g.TotalMatches = m
}

// SearchGames performs a search for the provided game title
func (h *HLTBClient) SearchGames(query string) (*GameResultsPage, error) {
	return gameSearch(h, &HLTBQuery{Query: query})
}

// SearchGamesByQuery queries using a set of user defined parameters
func (h *HLTBClient) SearchGamesByQuery(q *HLTBQuery) (*GameResultsPage, error) {
	return gameSearch(h, q)
}

func gameSearch(h *HLTBClient, q *HLTBQuery) (*GameResultsPage, error) {
	handleGameDefaults(q)
	doc, err := searchQuery(h, q)
	if err != nil {
		return nil, err
	}
	if !hasResults(doc) {
		return &GameResultsPage{}, nil
	}
	res, err := parseGameResponse(doc, q)
	if err != nil {
		return nil, err
	}
	res.hltbClient = h
	return res, nil
}

// handleGameDefaults will set all of the default form values for the query.
func handleGameDefaults(q *HLTBQuery) {
	q.QueryType = GameQuery
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

// parseGameResponse converts the http response object into a ManyGameTimes object
func parseGameResponse(doc *goquery.Document, q *HLTBQuery) (*GameResultsPage, error) {
	games := &GameResultsPage{}
	var gameslice []*GameResult

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

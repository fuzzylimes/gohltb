package gohltb

import (
	"encoding/json"
	"errors"
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
}

// HasNext will check to see if there is a next page that can be retrieved
func (g *GameResultsPage) HasNext() bool {
	return g.NextPage != 0
}

// GetNextPage will return the next page, if present
func (g *GameResultsPage) GetNextPage() (ResultPage, error) {
	if !g.HasNext() {
		return &GameResultsPage{}, errors.New("Page not found")
	}
	query := g.requestQuery
	query.Page = g.NextPage
	return SearchGamesByQuery(query)
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

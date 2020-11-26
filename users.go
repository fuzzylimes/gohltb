package gohltb

import (
	"encoding/json"
	"errors"
)

// UserResult is a user that uses howlongtobeat.com
type UserResult struct {
	ID        string            `json:"id"`                  // ID in howlongtobeat.com's database
	Name      string            `json:"name"`                // User's display name
	URL       string            `json:"url"`                 // Link to user's page
	AvatarURL string            `json:"avatar-url"`          // Link to avatar
	Location  string            `json:"location,omitempty"`  // User's location
	Backlog   string            `json:"backlog,omitempty"`   // Number of games in user's backlog
	Complete  string            `json:"complete,omitempty"`  // Number of games the user's completed
	Gender    string            `json:"gender,omitempty"`    // User's gender
	Posts     string            `json:"posts,omitempty"`     // Number of user's forum posts
	Age       int               `json:"age,omitempty"`       // User's age
	Accolades map[string]string `json:"accolades,omitempty"` // User accolades for activity on the site. This is things like years of service.
}

// UserResultsPage is a page of user responses
type UserResultsPage struct {
	Users        []*UserResult `json:"users"`       // Slice of UserResult
	TotalPages   int           `json:"total-pages"` // Total number of pages
	CurrentPage  int           `json:"curent-page"` // Current page number
	NextPage     int           `json:"next-page"`   // Next page number
	TotalMatches int           `json:"matches"`     // Total query matches
	requestQuery *HLTBQuery    // Query associated with this response
}

// HasNext will check to see if there is a next page that can be retrieved
func (u *UserResultsPage) HasNext() bool {
	return u.NextPage != 0
}

// GetNextPage will return the next page, if present
func (u *UserResultsPage) GetNextPage() (ResultPage, error) {
	if !u.HasNext() {
		return &UserResultsPage{}, errors.New("Page not found")
	}
	query := u.requestQuery
	query.Page = u.NextPage
	return SearchUsersByQuery(query)
}

// JSON will convert user object into a json string
func (u *UserResultsPage) JSON() (string, error) {
	var r string
	s, err := json.MarshalIndent(u.Users, "", "  ")
	if err == nil {
		r = string(s)
	}
	return r, err
}

package gohltb

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// UserResult is a user that uses howlongtobeat.com
type UserResult struct {
	ID        string   `json:"id"`                  // ID in howlongtobeat.com's database
	Name      string   `json:"name"`                // User's display name
	URL       string   `json:"url"`                 // Link to user's page
	AvatarURL string   `json:"avatar-url"`          // Link to avatar
	Location  string   `json:"location,omitempty"`  // User's location
	Backlog   string   `json:"backlog,omitempty"`   // Number of games in user's backlog
	Complete  string   `json:"complete,omitempty"`  // Number of games the user's completed
	Gender    string   `json:"gender,omitempty"`    // User's gender
	Posts     string   `json:"posts,omitempty"`     // Number of user's forum posts
	Age       int      `json:"age,omitempty"`       // User's age
	Accolades []string `json:"accolades,omitempty"` // User accolades for activity on the site. This is things like years of service.
}

// UserResultsPage is a page of user responses
type UserResultsPage struct {
	Users        []*UserResult `json:"users"`       // Slice of UserResult
	TotalPages   int           `json:"total-pages"` // Total number of pages
	CurrentPage  int           `json:"curent-page"` // Current page number
	NextPage     int           `json:"next-page"`   // Next page number
	TotalMatches int           `json:"matches"`     // Total query matches
	requestQuery *HLTBQuery    // Query associated with this response
	hltbClient   *HLTBClient   // Client for further requests
}

// HasNext will check to see if there is a next page that can be retrieved
func (u *UserResultsPage) HasNext() bool {
	return u.NextPage != 0
}

// GetNextPage will return the next page, if present
func (u *UserResultsPage) GetNextPage() (*UserResultsPage, error) {
	if !u.HasNext() {
		return &UserResultsPage{}, errors.New("Page not found")
	}
	query := u.requestQuery
	query.Page = u.NextPage
	return u.hltbClient.SearchUsersByQuery(query)
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

func (u *UserResultsPage) setTotalPages(p int) {
	u.TotalPages = p
}

func (u *UserResultsPage) setTotalMatches(m int) {
	u.TotalMatches = m
}

// SearchUsers performs a search for the provided user name
func (h *HLTBClient) SearchUsers(query string) (*UserResultsPage, error) {
	return userSearch(h, &HLTBQuery{Query: query})
}

// SearchUsersByQuery queries using a set of user defined parameters
func (h *HLTBClient) SearchUsersByQuery(q *HLTBQuery) (*UserResultsPage, error) {
	return userSearch(h, q)
}

func userSearch(h *HLTBClient, q *HLTBQuery) (*UserResultsPage, error) {
	handleUserDefaults(q)
	doc, err := searchQuery(h, q)
	if err != nil {
		return nil, err
	}
	if !hasResults(doc) {
		return &UserResultsPage{}, nil
	}
	res, err := parseUserResponse(doc, q)
	if err != nil {
		return nil, err
	}
	res.hltbClient = h
	return res, nil
}

// handleUserDefaults will set all of the default form values for the query.
func handleUserDefaults(q *HLTBQuery) {
	q.QueryType = UserQuery
	if q.SortBy == "" {
		q.SortBy = SortByUserTopPosters
	}
	if q.SortDirection == "" {
		q.SortDirection = NormalOrder
	}
	if q.Page == 0 {
		q.Page = 1
	}
}

// parseUserResponse converts the http response object into a ManyGameTimes object
func parseUserResponse(doc *goquery.Document, q *HLTBQuery) (*UserResultsPage, error) {
	users := &UserResultsPage{}
	var userslice []*UserResult

	// Handle the page numbers
	parsePages(doc, users, q.Page)

	doc.Find("ul > li.back_darkish").Each(func(userCount int, userDetails *goquery.Selection) {
		var location string
		var accolades []string

		avatar, _ := userDetails.Find("div > a > img").Attr("src")
		name := userDetails.Find("h3 > a")
		url, _ := name.Attr("href")
		id := strings.SplitAfter(url, "user?n=")[1]
		if ok := userDetails.Find("h4").Length(); ok > 0 {
			location = userDetails.Find("h4").First().Text()
		}
		if ok := userDetails.Find(".search_list_details > h3 > span").Length(); ok > 0 {
			accoladeSpan := userDetails.Find(".search_list_details > h3 > span").First()
			accoladeSpan.Find("span").Each(func(aCount int, accolade *goquery.Selection) {
				if v, exists := accolade.Attr("title"); exists {
					accolades = append(accolades, v)
				}
			})
		}

		user := &UserResult{
			ID:        id,
			URL:       urlPrefix + url,
			AvatarURL: urlPrefix + avatar,
			Name:      name.Text(),
			Location:  location,
			Accolades: accolades,
		}

		userDetails.Find(".search_list_tidbit.text_white").Each(func(userDetailCount int, userDetail *goquery.Selection) {
			nextValue := strings.TrimSpace(userDetail.Next().Text())
			switch category := userDetail.Text(); {
			case category == "Backlog":
				user.Backlog = nextValue
			case category == "Complete":
				user.Complete = nextValue
			case category == "Gender":
				user.Gender = nextValue
			case category == "Age":
				if i, err := strconv.Atoi(nextValue); err == nil {
					user.Age = i
				}
			case category == "Posts":
				user.Posts = nextValue
			}
		})

		userslice = append(userslice, user)
	})
	users.Users = userslice
	users.CurrentPage = q.Page
	users.requestQuery = q
	if users.TotalPages > users.CurrentPage {
		users.NextPage = users.CurrentPage + 1
	}

	return users, nil

}

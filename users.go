package gohltb

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// UserResult are the data object for Users. They contain information about a user
// that can be collected from the user query response. Users are the people on
// howlongtobeat.com that contribute completion times to the site or track their game
// collections.
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

// UserResultsPage is a page of user responses. It is the data model that will be
// returned for any User query. Provides ability to move between pages of User
// results.
type UserResultsPage struct {
	Users        []*UserResult `json:"users"`       // Slice of UserResult
	TotalPages   int           `json:"total-pages"` // Total number of pages
	CurrentPage  int           `json:"curent-page"` // Current page number
	NextPage     int           `json:"next-page"`   // Next page number
	TotalMatches int           `json:"matches"`     // Total query matches
	requestQuery *HLTBQuery    // Query associated with this response
	hltbClient   *HLTBClient   // Client that was used for the initial request, needed for retrieving later pages
}

// HasNext will check to see if there is a next page that can be retrieved
func (u *UserResultsPage) HasNext() bool {
	return u.NextPage != 0
}

// GetNextPage will return the next page, if it exists. Uses the client
// from the initial request to make additional queries.
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

// setTotalPages for Pages interface
func (u *UserResultsPage) setTotalPages(p int) {
	u.TotalPages = p
}

// setTotalMatches for Pages interface
func (u *UserResultsPage) setTotalMatches(m int) {
	u.TotalMatches = m
}

// SearchUsers performs a search for the provided user name. Will populate
// default values for the remaining query parameters aside from the provided
// user name.
//
// Note: A query with an empty string will query ALL users.
func (h *HLTBClient) SearchUsers(query string) (*UserResultsPage, error) {
	return userSearch(h, &HLTBQuery{Query: query})
}

// SearchUsersByQuery queries using a set of user defined parameters. Used
// for running more complex queries. Takes in a HLTBQuery object, with
// parameters tailored to a user query. You may include any of the following:
//     * Query - user name
//     * QueryType - UserQuery
//     * SortBy - one of the various "SortByUser" options
//     * SortDirection - sort direcion, based on SortBy
//     * Random - whether or not to retrieve random value
//     * Page - page of responses to return
//
// Note: A query with an empty "Query" string will query ALL users.
func (h *HLTBClient) SearchUsersByQuery(q *HLTBQuery) (*UserResultsPage, error) {
	return userSearch(h, q)
}

// userSearch is the central method for running user queries
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

// parseUserResponse parses the http response object into a UserResultsPage object
func parseUserResponse(doc *goquery.Document, q *HLTBQuery) (*UserResultsPage, error) {
	users := &UserResultsPage{}
	var userslice []*UserResult

	// Handle the page numbers
	parsePages(doc, users, q.Page)

	// Handle each user
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
		// Handle any user accolades - these are awards earned by users
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

		// There are set of different options that a user can choose to enter, none of
		// which are mandatory. This will handle each of them, if they're present.
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

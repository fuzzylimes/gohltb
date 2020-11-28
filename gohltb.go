package gohltb

import (
	"errors"
	"fmt"
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
	urlPrefix string = queryURL + "/"
)

// HLTBClient is the main client used to interact with the APIs. You should
// be creating a client via one of the two "New" methods, either NewDefaultClient
// or NewCustomClient. Either will be needed to perform game or user queries.
type HLTBClient struct {
	Client *HTTPClient
}

// HTTPClient handles the connectivity details for the client. This is handled
// automatically when using the NewDefaultClient constructor, but can optionally
// be supplied when creating a NewCustomClient. This allows the user to create
// their own http.Client and pass it in for use. This can be useful if you want
// to use something other than the default configuration.
type HTTPClient struct {
	Client  *http.Client
	baseURL string
}

// Pages are a type of data structure for paged responses.
type Pages interface {
	setTotalPages(int)
	setTotalMatches(int)
}

// HLTBQuery is a query object used to run user defined queries. Both game
// and user queries both use the HLTBQuery as their means of query. The library
// handles any mandatory parameters for you, making all parameters optional when
// constructing a query.
//
// example: client.SearchGamesByQuery(&HLTBQuery{query: "Mario", random: true})
type HLTBQuery struct {
	Query         string 		// String to query by
	QueryType     QueryType		// Type of query to perform - games or users
	SortBy        SortBy        // Specify how data should be sorted
	SortDirection SortDirection // Specify direction data should be sorted
	Platform      Platform		// Platform to query against (only used with game queries)
	LengthType    LengthRange	// Optional filter based on completion times (games only)
	LengthMin     string		// Optional min length for LengthType (games only)
	LengthMax     string		// Optional max length for LengthType (games only)
	Modifier      Modifier		// Toggle additional filter methods (games only)
	Random        bool			// Return a single, random, entry based on parameters
	Page          int			// Page number to return
}

// NewDefaultClient will create a new HLTBClient with default parameters. This
// is what you should use to create you client if you don't need to use a specific
// http.Client/settings
func NewDefaultClient() *HLTBClient {
	return &HLTBClient{
		Client: &HTTPClient{
			Client:  client,
			baseURL: queryURL,
		},
	}
}

// NewCustomClient will create a new HLTBClient with provided HTTPClient. Users
// are able to pass in an HTTPClient with their desired http.Client.
func NewCustomClient(c *HTTPClient) *HLTBClient {
	if c.baseURL == "" {
		c.baseURL = queryURL
	}
	if c.Client == nil {
		c.Client = client
	}
	return &HLTBClient{
		Client: c,
	}
}

// searchQuery is a general helper method used by both Game and User queries. It
// handles the common activies shared between both query types. This is where
// data is scraped from howlongtobeat.com
func searchQuery(c *HLTBClient, q *HLTBQuery) (*goquery.Document, error) {
	form := buildForm(q)

	req, err := http.NewRequest("POST", fmt.Sprintf("%v/search_results?page=%v", c.Client.baseURL, q.Page), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.Client.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("Error retrieving data")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// buildForm will construct the form payload sent to the server. The query
// to the server expects a very specific form; this method handles the
// construction of that form based on query values.
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

// hasResults checks to see if the query has any matches. If no matches found,
// returns false
func hasResults(doc *goquery.Document) bool {
	return !strings.Contains(doc.Find("li").Nodes[0].FirstChild.Data, "No results")
}

// parsePages is a general helper function that handles collecting:
//    * the total number of matches (only available when querying the first page)
//    * the total number of pages
// Utilized for both game and user queries
func parsePages(doc *goquery.Document, page Pages, p int) {
	if p == 1 {
		// regex to extract number of games/users
		regex := regexp.MustCompile(`Found ([1-9]+)`)
		numMatches := regex.FindStringSubmatch(doc.Find("h3").Nodes[0].FirstChild.Data)[1]
		numMatchesInt, err := strconv.Atoi(numMatches)
		if err != nil {
			// All pages after 1 will be set to 0, this is a TODO
			page.setTotalMatches(0)
		} else {
			page.setTotalMatches(numMatchesInt)
		}
	}

	// scan page for the last page element and convert from string to int
	lastPage, err := strconv.Atoi(doc.Find("span.search_list_page").Last().Text())
	if err != nil {
		// if we can't find it, set it to 1
		page.setTotalPages(1)
	} else {
		page.setTotalPages(lastPage)
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

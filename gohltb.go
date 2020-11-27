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
	urlPrefix string = "https://howlongtobeat.com/"
)

// HLTBClient is the main client used to interact with the APIs. You should
// be creating a client via one of the two "New" methods
type HLTBClient struct {
	Client *HTTPClient
}

// HTTPClient specifies client specific parameters. We automatically populate
// with some default client parameters, but this allows the user to overwrite
// with whatever values work better for them.
type HTTPClient struct {
	Client  *http.Client
	baseURL string
}

// Pages are a type of data structure for paged responses
type Pages interface {
	setTotalPages(int)
	setTotalMatches(int)
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

// NewDefaultClient will create a new HLTBClient with default parameters
func NewDefaultClient() *HLTBClient {
	return &HLTBClient{
		Client: &HTTPClient{
			Client:  client,
			baseURL: queryURL,
		},
	}
}

// NewCustomClient will create a new HLTBClient with custom parameters
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

func hasResults(doc *goquery.Document) bool {
	return !strings.Contains(doc.Find("li").Nodes[0].FirstChild.Data, "No results")
}

func parsePages(doc *goquery.Document, page Pages, p int) {
	if p == 1 {
		regex := regexp.MustCompile(`Found ([1-9]+)`)
		numMatches := regex.FindStringSubmatch(doc.Find("h3").Nodes[0].FirstChild.Data)[1]
		numMatchesInt, err := strconv.Atoi(numMatches)
		if err != nil {
			page.setTotalMatches(0)
		} else {
			page.setTotalMatches(numMatchesInt)
		}
	}

	lastPage, err := strconv.Atoi(doc.Find("span.search_list_page").Last().Text())
	if err != nil {
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

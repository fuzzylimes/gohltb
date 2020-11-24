package gohltb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeCall(file string, q *HLTBQuery) (*GameTimesPage, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return &GameTimesPage{}, err
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(data))
	}))
	defer ts.Close()

	mockURL := ts.URL

	resp, err := handleSearch(mockURL, q)
	// j, _ := resp.JSON()
	// fmt.Println(j)
	return resp, nil
}

func TestValidHtmlParse(t *testing.T) {
	res, err := makeCall("testdata/basic_response.html", &HLTBQuery{Query: "pokemon red"})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	if len(res.Games) != 2 {
		fmt.Printf("Got %v, expected 2", len(res.Games))
		t.Fail()
	}
}

func TestNonStandardParse(t *testing.T) {
	res, err := makeCall("testdata/non_standard.html", &HLTBQuery{Query: "pokemon red"})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	if res.Games[0].Other == nil {
		t.Fail()
	}
	if res.Games[0].Other["Vs."] != "2½ Hours" {
		fmt.Printf("Got %v, expected 2½ Hours", res.Games[0].Other["Vs."])
		t.Fail()
	}
}

func TestRandomParse(t *testing.T) {
	res, err := makeCall("testdata/basic_response.html", &HLTBQuery{Query: "pokemon red", Random: true})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	if len(res.Games) != 2 {
		fmt.Printf("Got %v, expected 2", len(res.Games))
		t.Fail()
	}
}

func TestUserStatsParse(t *testing.T) {
	res, err := makeCall("testdata/userstats.html", &HLTBQuery{Page: 2, Modifier: "user_stats"})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	if len(res.Games) != 2 {
		fmt.Printf("Got %v, expected 2", len(res.Games))
		t.Fail()
	}
	if res.Games[1].UserStats.Completed != "5.3K" {
		fmt.Printf("Got %v, expected 5.3K", res.Games[1].UserStats.Completed)
		t.Fail()
	}
}

func TestMultiplePages(t *testing.T) {
	res, err := makeCall("testdata/multipage.html", &HLTBQuery{})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	if res.Games[0].Title != "!4RC4N01D!" {
		fmt.Printf("Got %v, expected !4RC4N01D!", res.Games[0].Title)
		t.Fail()
	}
	if res.Games[0].ID != "53256" {
		fmt.Printf("Got %v, expected 53256", res.Games[0].ID)
		t.Fail()
	}
	if res.CurrentPage != 1 {
		fmt.Printf("Got %v, expected 1", res.CurrentPage)
		t.Fail()
	}
	if res.NextPage != 2 {
		fmt.Printf("Got %v, expected 2", res.NextPage)
		t.Fail()
	}
	if res.TotalMatches != 42848 {
		fmt.Printf("Got %v, expected 42848", res.TotalMatches)
		t.Fail()
	}
	if res.TotalPages != 2143 {
		fmt.Printf("Got %v, expected 2143", res.TotalPages)
		t.Fail()
	}
	if !res.HasNext() {
		fmt.Println("There should be a next page")
		t.Fail()
	}
}

func TestNoMatchParse(t *testing.T) {
	res, err := makeCall("testdata/notfound.html", &HLTBQuery{Query: "bugsnaxasdf"})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	if len(res.Games) != 0 {
		fmt.Printf("Got %v, expected 0", len(res.Games))
		t.Fail()
	}
}

// Negative tests
func TestNon200Response(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/bugsnax.html")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		fmt.Fprintln(w, string(data))
	}))
	defer ts.Close()

	mockURL := ts.URL

	resp, err := handleSearch(mockURL, &HLTBQuery{Query: "bugsnax"})
	if err.Error() != "Error retrieving data" {
		fmt.Printf("Got %v, expected Error retrieving data", err.Error())
		t.Fail()
	}
	if resp.Games != nil {
		fmt.Println("Expected nil object")
		t.Fail()
	}
}

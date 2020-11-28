package gohltb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeGameCall(file string, q *HLTBQuery) (*GameResultsPage, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return &GameResultsPage{}, err
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(data))
	}))
	defer ts.Close()

	mockURL := ts.URL
	client := NewCustomClient(&HTTPClient{baseURL: mockURL})

	resp, err := client.SearchGamesByQuery(q)
	return resp, err
}

func TestValidGameParse(t *testing.T) {
	res, err := makeGameCall("testdata/games/basic_response.html", &HLTBQuery{Query: "pokemon red"})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	if len(res.Games) != 2 {
		fmt.Printf("Got %v, expected 2", len(res.Games))
		t.Fail()
	}
	res.JSON()
}

func TestNonStandardGameParse(t *testing.T) {
	res, err := makeGameCall("testdata/games/non_standard.html", &HLTBQuery{Query: "pokemon red"})
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

func TestRandomGameParse(t *testing.T) {
	res, err := makeGameCall("testdata/games/basic_response.html", &HLTBQuery{Query: "pokemon red", Random: true})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	if len(res.Games) != 2 {
		fmt.Printf("Got %v, expected 2", len(res.Games))
		t.Fail()
	}
}

func TestUserStatsParse(t *testing.T) {
	res, err := makeGameCall("testdata/games/userstats.html", &HLTBQuery{Page: 2, Modifier: "user_stats"})
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

func TestMultipleGamePages(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/games/multipage.html")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	data2, err := ioutil.ReadFile("testdata/games/multipage2.html")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := r.URL.Query()["page"]
		if page[0] == "1" {
			fmt.Fprintln(w, string(data))
		} else {
			fmt.Fprintln(w, string(data2))
		}
	}))
	defer ts.Close()

	mockURL := ts.URL
	client := NewCustomClient(&HTTPClient{baseURL: mockURL})

	res, err := client.SearchGamesByQuery(&HLTBQuery{})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
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

	res, err = res.GetNextPage()
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	if res.CurrentPage != 2 {
		fmt.Printf("Got %v, expected 2", res.CurrentPage)
		t.Fail()
	}
	if res.NextPage != 3 {
		fmt.Printf("Got %v, expected 3", res.NextPage)
		t.Fail()
	}
	if res.TotalMatches != 0 {
		fmt.Printf("Got %v, expected 0", res.TotalMatches)
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

// Negative tests
func TestNoMatchGameParse(t *testing.T) {
	res, err := makeGameCall("testdata/games/notfound.html", &HLTBQuery{Query: "bugsnaxasdf"})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	if len(res.Games) != 0 {
		fmt.Printf("Got %v, expected 0", len(res.Games))
		t.Fail()
	}
}

func TestNon200GameResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	}))
	defer ts.Close()

	mockURL := ts.URL
	client := NewCustomClient(&HTTPClient{baseURL: mockURL})

	resp, err := client.SearchGamesByQuery(&HLTBQuery{Query: "bugsnax"})
	if err.Error() != "Error retrieving data" {
		fmt.Printf("Got %v, expected Error retrieving data", err.Error())
		t.Fail()
	}
	if resp != nil {
		fmt.Println("Expected nil object")
		t.Fail()
	}
}

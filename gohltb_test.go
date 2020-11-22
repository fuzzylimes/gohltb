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
	j, _ := resp.JSON()
	fmt.Println(j)
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

func TestSingleGameParse(t *testing.T) {
	res, err := makeCall("testdata/bugsnax.html", &HLTBQuery{Query: "bugsnax"})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	if len(res.Games) != 1 {
		fmt.Printf("Got %v, expected 1", len(res.Games))
		t.Fail()
	}
	if res.Games[0].Title != "Bugsnax" {
		fmt.Printf("Got %v, expected Bugsnax", res.Games[0].Title)
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

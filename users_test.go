package gohltb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeUserCall(file string, q *HLTBQuery) (*UserResultsPage, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return &UserResultsPage{}, err
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(data))
	}))
	defer ts.Close()

	mockURL := ts.URL
	client := NewCustomClient(&HTTPClient{baseURL: mockURL})

	resp, err := client.SearchUsersByQuery(q)
	return resp, err
}

func TestValidHtmlParse(t *testing.T) {
	res, err := makeUserCall("testdata/users/mixed_response.html", &HLTBQuery{Query: "pokemon red"})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	if len(res.Users) != 2 {
		fmt.Printf("Got %v, expected 2", len(res.Users))
		t.Fail()
	}
	user1 := res.Users[0]
	user2 := res.Users[1]
	if len(user1.Accolades) != 5 {
		fmt.Printf("Got %v, expected 5", len(user1.Accolades))
		t.Fail()
	}
	if user2.Accolades != nil {
		fmt.Printf("Got %v, expected nil", user2.Accolades)
		t.Fail()
	}
	if user1.Location == "" {
		fmt.Printf("Got %v, expected non-empty", user1.Location)
		t.Fail()
	}
	if user2.Location != "" {
		fmt.Printf("Got %v, expected empty", user2.Location)
		t.Fail()
	}
	if user1.Age != 39 {
		fmt.Printf("Got %v, expected 39", user1.Age)
		t.Fail()
	}
	if user2.Age != 0 {
		fmt.Printf("Got %v, expected 0", user2.Age)
		t.Fail()
	}
	res.JSON()
}

func TestMultipleUserPages(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/users/multipage.html")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	data2, err := ioutil.ReadFile("testdata/users/multipage2.html")
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

	res, err := client.SearchUsersByQuery(&HLTBQuery{})
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
	if res.TotalMatches != 255629 {
		fmt.Printf("Got %v, expected 255629", res.TotalMatches)
		t.Fail()
	}
	if res.TotalPages != 12782 {
		fmt.Printf("Got %v, expected 12782", res.TotalPages)
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
	if res.TotalPages != 12782 {
		fmt.Printf("Got %v, expected 12782", res.TotalPages)
		t.Fail()
	}
	if !res.HasNext() {
		fmt.Println("There should be a next page")
		t.Fail()
	}
}

func TestRandomUserParse(t *testing.T) {
	res, err := makeUserCall("testdata/users/mixed_response.html", &HLTBQuery{Query: "asdf", Random: true})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	if len(res.Users) != 2 {
		fmt.Printf("Got %v, expected 2", len(res.Users))
		t.Fail()
	}
}

// Negative tests
func TestNoMatchUserParse(t *testing.T) {
	res, err := makeUserCall("testdata/users/notfound.html", &HLTBQuery{Query: "asdfasdfasdfedawe"})
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	if len(res.Users) != 0 {
		fmt.Printf("Got %v, expected 0", len(res.Users))
		t.Fail()
	}
}

func TestNon200UserResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	}))
	defer ts.Close()

	mockURL := ts.URL
	client := NewCustomClient(&HTTPClient{baseURL: mockURL})

	resp, err := client.SearchUsersByQuery(&HLTBQuery{Query: "bugsnax"})
	if err.Error() != "Error retrieving data" {
		fmt.Printf("Got %v, expected Error retrieving data", err.Error())
		t.Fail()
	}
	if resp != nil {
		fmt.Println("Expected nil object")
		t.Fail()
	}
}

package gohltb

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func makeCall(file string) (*ManyGameTimes, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return &ManyGameTimes{}, err
	}
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(data)),
	}

	res, err := parseResponse(mockResponse)
	if err != nil {
		return &ManyGameTimes{}, err
	}
	j, _ := res.JSON()
	fmt.Println(j)
	return res, nil
}

func TestValidHtmlParse(t *testing.T) {
	res, err := makeCall("testdata/basic_response.html")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	if len(res.Games) != 2 {
		fmt.Printf("Got %v, expected 2", len(res.Games))
		t.Fail()
	}
}

func TestSingleGameParse(t *testing.T) {
	res, err := makeCall("testdata/bugsnax.html")
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


func TestNoMatch(t *testing.T) {
	res, err := makeCall("testdata/notfound.html")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	if len(res.Games) != 0 {
		fmt.Printf("Got %v, expected 0", len(res.Games))
		t.Fail()
	}
}
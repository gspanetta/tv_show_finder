package main

import (
	"testing"
	"bytes"
	"strings"
)

func TestUserInputString(t *testing.T) {
	var context Context
	var stdin bytes.Buffer
	var userInput string
	context.input = &stdin

	stdin.Write([]byte("1234"))
	userInput = getUserInputStr(&context, "test")

	if strings.Compare(userInput, "1234") != 0 {
		t.Errorf("Expected: 1234; got: %v", userInput)
	}

	stdin.Write([]byte("hello world\n"))
	userInput = getUserInputStr(&context, "test")

	if strings.Compare(userInput, "hello world") != 0 {
		t.Errorf("Expected: hello world; got: %v", userInput)
	}

	stdin.Write([]byte("hello world\r\n"))
	userInput = getUserInputStr(&context, "test")

	if strings.Compare(userInput, "hello world") != 0 {
		t.Errorf("Expected: hello world; got: %v", userInput)
	}

}

func TestUserInputNumber(t *testing.T) {
	var context Context
	var stdin bytes.Buffer
	context.input = &stdin

	stdin.Write([]byte("12"))
	num, err := getUserInputNum(&context, "test")

	if err != nil {
		t.Errorf("getUserInputNum threw an error: %v on input '12'", err)
	}

	if num != 12 {
		t.Errorf("Expected: 12; got: %v", num)
	}

	stdin.Write([]byte("hello world\n"))
	num, err = getUserInputNum(&context, "test")

	if err == nil {
		t.Errorf("getUserInputNum didn't throw an error on string input")
	}

	if num != 0 {
		t.Errorf("Expected: 0; got: %v", num)
	}
}

type TmdbTestResponse struct {
	Status_code		int		`json:"status_code"`
	Status_message  string	`json:"status_message"`
	Success			bool	`json:"success"`
}

type TmdbTestWrongResponse struct {
	Status_code		int		`json:"status_code"`
	Status_message  string	`json:"status_message"`
	Success			int		`json:"success"`
}

func TestHttpRequest(t *testing.T) {
	var testResponse TmdbTestResponse
	var testWrongResponse TmdbTestWrongResponse
	var url string = "https://api.themoviedb.org/3/search/"
	var wrongUrl string = "https://api.themoviedbasdf.org/3/search/"

	err := httpRequest(url, &testResponse)
	if err != nil {
		t.Errorf("httpRequest threw an error: %v", err)
	}

	err = httpRequest(wrongUrl, &testResponse)
	if err == nil {
		t.Errorf("httpRequest didn't threw an error when a non-existing url was used")
	}

	err = httpRequest(url, &testWrongResponse)
	if err == nil {
		t.Errorf("httpRequest didn't threw an error when a wrong struct was used")
	}
}
package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"strconv"
	"testing"
)

// get path to files
func init() {
	_, filename, _, _ := runtime.Caller(0)
	// The ".." may change depending on you folder structure
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestRenewHistoryGet(t *testing.T) {

	msg := make(chan string)

	// Initialize handler instance
	handler := RenewHistoryHandler(msg)

	// Set up infrastructure to be used for invocation - important: wrap handler function in http.HandlerFunc()
	server := httptest.NewServer(http.HandlerFunc(handler))
	// Ensure it is torn down properly at the end
	defer server.Close()

	// Create client instance
	client := http.Client{}

	// Retrieve content from server
	res, err := client.Get(server.URL + RENEW_HISTORY_ENDPOINT + "NOR/" + "?begin=1965&end=1966")
	if err != nil {
		t.Fatal("Get request to URL failed:", err.Error())
	}
	// Decode array
	testStruct := []history{}
	err2 := json.NewDecoder(res.Body).Decode(&testStruct)

	if err2 != nil {
		t.Fatal("Error during decoding:", err2.Error())
	}

	if len(testStruct) != 2 {
		t.Fatal("Number of returned countries are wrong : " + strconv.Itoa(len(testStruct)))
	}
	//Expected output
	expected := []history{
		{Entity: "Norway", Code: "NOR", Year: 1965, Percentage: 67.87996},
		{Entity: "Norway", Code: "NOR", Year: 1966, Percentage: 65.3991},
	}

	assert.EqualValues(t, expected, testStruct, "The expected and actual output should be the same")

	// Perform content check

}

func TestMean(t *testing.T) {

	msg := make(chan string)

	// Initialize handler instance
	handler := RenewHistoryHandler(msg)

	// Set up infrastructure to be used for invocation - important: wrap handler function in http.HandlerFunc()
	server := httptest.NewServer(http.HandlerFunc(handler))
	// Ensure it is torn down properly at the end
	defer server.Close()

	// Create client instance
	client := http.Client{}

	// URL under which server is instantiated
	fmt.Println("URL: ", server.URL)

	// Retrieve content from server

	res, err := client.Get(server.URL + RENEW_HISTORY_ENDPOINT)
	if err != nil {
		t.Fatal("Get request to URL failed:", err.Error())
	}

	// Decode array
	testStruct := []history{}
	err2 := json.NewDecoder(res.Body).Decode(&testStruct)
	if err2 != nil {
		t.Fatal("Error during decoding:", err2.Error())
	}

	if len(testStruct) != 105 {
		t.Fatal("Number of returned countries are wrong wrong: " + strconv.Itoa(len(testStruct)))
	}

	// Perform content check
}

func TestSortByvalue(t *testing.T) {

	msg := make(chan string)

	// Initialize handler instance
	handler := RenewHistoryHandler(msg)
	// do something with the reque

	// Set up infrastructure to be used for invocation - important: wrap handler function in http.HandlerFunc()
	server := httptest.NewServer(http.HandlerFunc(handler))
	// Ensure it is torn down properly at the end
	defer server.Close()

	// Create client instance
	client := http.Client{}

	// URL under which server is instantiated
	fmt.Println("URL: ", server.URL)

	// Retrieve content from server

	res, err := client.Get(server.URL + RENEW_HISTORY_ENDPOINT + "NOR/" + "?sortByValue=true")
	if err != nil {
		t.Fatal("Get request to URL failed:", err.Error())
	}

	// Decode array
	testStruct := []history{}
	err2 := json.NewDecoder(res.Body).Decode(&testStruct)
	if err2 != nil {
		t.Fatal("Error during decoding:", err2.Error())
	}

	if len(testStruct) != 57 {
		t.Fatal("Number of returned countries are wrong : " + strconv.Itoa(len(testStruct)))
	}

}

func TestInvalidRequest(t *testing.T) {
	msg := make(chan string)

	// Initialize handler instance
	handler := RenewHistoryHandler(msg)

	// Set up infrastructure to be used for invocation - important: wrap handler function in http.HandlerFunc()
	server := httptest.NewServer(http.HandlerFunc(handler))
	// Ensure it is torn down properly at the end
	defer server.Close()
	testStruct := []struct {
		description string
		url         string
		method      string
		StatusCode  int
		error       string
	}{
		{
			description: "Rest method supported method",
			url:         server.URL + RENEW_HISTORY_ENDPOINT,
			method:      http.MethodPost,
			StatusCode:  http.StatusMethodNotAllowed,
			error: "REST Method '" + http.MethodPost + "' not supported. Currently only '" + http.MethodGet +
				" is supported.",
		},
		{
			description: "Bad request",
			url:         server.URL + RENEW_HISTORY_ENDPOINT + "hjk/jkg",
			method:      http.MethodGet,
			StatusCode:  http.StatusBadRequest,
			error:       "unexpected format",
		},
		{
			description: "Bad begin query",
			url:         server.URL + RENEW_HISTORY_ENDPOINT + "?begin=dsas",
			method:      http.MethodGet,
			StatusCode:  http.StatusBadRequest,
			error:       "Invalid parameter must be an integer",
		},
		{
			description: "Bad end query",
			url:         server.URL + RENEW_HISTORY_ENDPOINT + "?end=dsalk",
			method:      http.MethodGet,
			StatusCode:  http.StatusBadRequest,
			error:       "Invalid parameter must be an integer",
		},

		{
			description: "Begin year more than end year",
			url:         server.URL + RENEW_HISTORY_ENDPOINT + "?begin=2000&end=1980",
			method:      http.MethodGet,
			StatusCode:  http.StatusBadRequest,
			error:       "unexpected format, make sure year inputed is valid and begin is < than end",
		},

		{
			description: "invalid sortbyvalue query",
			url:         server.URL + RENEW_HISTORY_ENDPOINT + "?sortByValue=sdad",
			method:      http.MethodGet,
			StatusCode:  http.StatusBadRequest,
			error:       "Invalid parameter must be a bool value",
		},
		{
			description: "invalid isocode",
			url:         server.URL + RENEW_HISTORY_ENDPOINT + "Norr",
			method:      http.MethodGet,
			StatusCode:  http.StatusBadRequest,
			error:       "iso code not found",
		},
	}
	for _, test := range testStruct {
		t.Run(test.description, func(t *testing.T) {
			client := &http.Client{}
			req, err := http.NewRequest(test.method, test.url, nil)
			if err != nil {
				t.Errorf("Test: %s. Error with request %s and url %s: %s", test.description, test.method, test.url, err.Error())
			}

			res, err := client.Do(req)
			if err != nil {
				t.Errorf("Test: %s. Error running %s request: %s", test.description, test.method, err.Error())
			}

			assert.Equal(t, test.StatusCode, res.StatusCode)
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Test: %s. Error reading response : %s", test.description, err.Error())
			}
			assert.Equal(t, test.error+"\n", string(body))
		})
	}
}

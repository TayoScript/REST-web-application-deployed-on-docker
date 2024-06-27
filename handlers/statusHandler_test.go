package handlers

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	// The ".." may change depending on you folder structure
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestStatusHandler(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(StatusHandler))
	defer server.Close()

	client := http.Client{}

	// Retrieve content from server
	res, err := client.Get(server.URL + STATUS_ENPOINT)
	if err != nil {
		t.Fatal("Get request to URL failed:", err.Error())
	}
	// Decode array
	testStruct := []Diagnostics{}
	err2 := json.NewDecoder(res.Body).Decode(&testStruct)

	if err2 != nil {
		t.Fatal("Error during decoding:", err2.Error())
	}
	expected := []Diagnostics{
		{CountriesApi: 200, NotificationDb: 200, Webhooks: webhooksNotifcation(), Version: "v1", Uptime: uptime()},
	}

	assert.EqualValues(t, expected, testStruct, "The expected and actual output should be the same")
}

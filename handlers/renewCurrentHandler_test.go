package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestReadCSV(t *testing.T) {

	csvData, err := ReadCSV(RENEWABLE_DATA_CSV)
	if err != nil {
		t.Error(err)
	}

	// Check if the CSV has the correct headers, these are hardcoded values that are expected to always stay the same
	if csvData[0][CSV_COL_ENTITY] != "Entity" {
		t.Error("Invalid header, column A! Expected: Entity")
	}

	if csvData[0][CSV_COL_CODE] != "Code" {
		t.Error("Invalid header, column B! Expected: Code")
	}

	if csvData[0][CSV_COL_YEAR] != "Year" {
		t.Error("Invalid header, column C! Expected: Year")
	}

	if csvData[0][CSV_COL_RENEWABLES] != "Renewables (% equivalent primary energy)" {
		t.Error("Invalid header, column D! Expected: Renewables (% equivalent primary energy)")
	}

	// Check for some entries
	found := []RenewableDataEntry{}
	for _, entry := range csvData {
		if entry[CSV_COL_ENTITY] == "Norway" && entry[CSV_COL_YEAR] == "2021" {
			percentage, err := strconv.ParseFloat(entry[CSV_COL_RENEWABLES], 64)
			if err != nil {
				t.Error("Failed to parse float for data entry")
			}
			found = append(found, RenewableDataEntry{
				Name:       entry[CSV_COL_ENTITY],
				ISOCode:    entry[CSV_COL_CODE],
				Year:       entry[CSV_COL_YEAR],
				Percentage: percentage,
			})
		}
		if entry[CSV_COL_ENTITY] == "Sweden" && entry[CSV_COL_YEAR] == "2021" {
			percentage, err := strconv.ParseFloat(entry[CSV_COL_RENEWABLES], 64)
			if err != nil {
				t.Error("Failed to parse float for data entry")
			}
			found = append(found, RenewableDataEntry{
				Name:       entry[CSV_COL_ENTITY],
				ISOCode:    entry[CSV_COL_CODE],
				Year:       entry[CSV_COL_YEAR],
				Percentage: percentage,
			})
		}
		if entry[CSV_COL_ENTITY] == "France" && entry[CSV_COL_YEAR] == "1990" {
			percentage, err := strconv.ParseFloat(entry[CSV_COL_RENEWABLES], 64)
			if err != nil {
				t.Error("Failed to parse float for data entry")
			}
			found = append(found, RenewableDataEntry{
				Name:       entry[CSV_COL_ENTITY],
				ISOCode:    entry[CSV_COL_CODE],
				Year:       entry[CSV_COL_YEAR],
				Percentage: percentage,
			})
		}
		if entry[CSV_COL_CODE] == "USA" && entry[CSV_COL_YEAR] == "1980" {
			percentage, err := strconv.ParseFloat(entry[CSV_COL_RENEWABLES], 64)
			if err != nil {
				t.Error("Failed to parse float for data entry")
			}
			found = append(found, RenewableDataEntry{
				Name:       entry[CSV_COL_ENTITY],
				ISOCode:    entry[CSV_COL_CODE],
				Year:       entry[CSV_COL_YEAR],
				Percentage: percentage,
			})
		}
	}
	if len(found) != 4 {
		t.Error(fmt.Sprintf("Unable to find the 4 test entries in the data, found %d entries", len(found)))
		t.Error(found)
	}
}

func TestGetCountryCodeMapping(t *testing.T) {
	// Get CSV data
	csvData, err := ReadCSV(RENEWABLE_DATA_CSV)
	if err != nil {
		t.Fatal(err)
	}

	// Get mapping
	mapping := GetCountryCodeMapping(csvData)

	// Test it
	code := mapping["norway"]

	if code != "nor" {
		t.Error("Expected: nor, Got: " + code)
	}

	code = mapping["sweden"]

	if code != "swe" {
		t.Error("Expected: swe, Got: " + code)
	}

	code = mapping["germany"]

	if code != "deu" {
		t.Error("Expected: deu, Got: " + code)
	}

	code = mapping["ukraine"]

	if code != "ukr" {
		t.Error("Expected: ukr, Got: " + code)
	}
}

func TestGetLatestYears(t *testing.T) {
	// Get CSV data
	csvData, err := ReadCSV(RENEWABLE_DATA_CSV)
	if err != nil {
		t.Fatal(err)
	}

	// Get latest years
	years, err := GetLatestYears(csvData)
	if err != nil {
		t.Fatal(err)
	}

	// Check if years are correct (all should have 2021 unless datafile has been changed)
	y := years["nor"]

	if y != "2021" {
		t.Error("Expected: 2021, Got: " + y)
	}

	y = years["swe"]

	if y != "2021" {
		t.Error("Expected: 2021, Got: " + y)
	}

	y = years["deu"]

	if y != "2021" {
		t.Error("Expected: 2021, Got: " + y)
	}

	y = years["ukr"]

	if y != "2021" {
		t.Error("Expected: 2021, Got: " + y)
	}
}

func TestRenewCurrentHandler(t *testing.T) {
	// Initialize data & handler
	csvData, err := ReadCSV(RENEWABLE_DATA_CSV)
	if err != nil {
		t.Fatal(err)
	}

	years, err := GetLatestYears(csvData)
	if err != nil {
		t.Fatal(err)
	}

	msg := make(chan string)

	handler := RenewCurrentHandler(csvData, years, msg)

	// Setup server
	server := httptest.NewServer(http.HandlerFunc(handler))
	// Close it at the end
	defer server.Close()

	// Setup client
	client := http.Client{}

	// Request A: Norway using code 'nor'

	// Make a request to the server
	req, err := http.NewRequest(http.MethodGet, server.URL+RENEW_CURRENT_ENDPOINT+"nor", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// Decode the response data
	resdata := []RenewableDataEntry{}

	json.NewDecoder(res.Body).Decode(&resdata)
	res.Body.Close()

	// Find the data that should be there
	if len(resdata) != 1 {
		t.Error(fmt.Sprintf("Request A: Length, Expected: 1, Got: %d", len(resdata)))
	}

	for _, entry := range resdata {
		if entry.Name != "Norway" && entry.Year != "2021" {
			t.Error(fmt.Sprintf("Request A: Data, Expected: Norway, 2021, Got: %s, %s", entry.Name, entry.Year))
		}
	}

	// Request B: Norway using name 'norway' with neighbours

	// Make a request to the server
	req, err = http.NewRequest(http.MethodGet, server.URL+RENEW_CURRENT_ENDPOINT+"norway?neighbours=true", nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// Decode the response data
	resdata = []RenewableDataEntry{}

	json.NewDecoder(res.Body).Decode(&resdata)
	res.Body.Close()

	// Find the data that should be there
	if len(resdata) != 4 {
		t.Error(fmt.Sprintf("Request A: Length, Expected: 4, Got: %d", len(resdata)))
	}

	found := false

	// Check for some the data present
	for _, entry := range resdata {
		if entry.Name == "Finland" && entry.Year == "2021" {
			found = true
		}
	}
	if !found {
		t.Error("Unable to find expected: Finland, 2021 in returned data!")
	}

	// See if the correct message was sent
	// TODO: Try to get this working
	/*
		m := <-msg
		if m != "nor" {
			t.Error(fmt.Sprintf("Request A: Channel, Expected: nor, Got: %s", m))
		}
	*/
}

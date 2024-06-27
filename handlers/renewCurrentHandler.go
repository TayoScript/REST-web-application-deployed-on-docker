package handlers

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

func RenewCurrentHandler(csvData [][]string, years map[string]string, msg chan string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Log requests
		log.Println("Started", r.Method, "on", r.URL)
		defer log.Println("Finished", r.Method, "on", r.URL)

		// Check for HTTP method (only GET is allowed)
		if r.Method != http.MethodGet {
			log.Println("E: Invalid method for this endpoint")
			http.Error(w, "Method not supported for this endpoint", http.StatusMethodNotAllowed)
		}

		// Parameters
		var country string = ""
		var neighbours bool = false

		// Get parameters

		// Country
		urlParts := strings.Split(r.URL.String(), "/")
		if len(urlParts) >= 6 {
			if len(urlParts[5]) != 0 {
				argumentParts := strings.Split(urlParts[5], "?")
				country = argumentParts[0]
				fmt.Printf("DEBUG: Country parameter set to '%s'\n", argumentParts[0])
			}
		}

		// Neighbours flag
		neighboursStr := r.URL.Query().Get("neighbours")
		neighbours, err := strconv.ParseBool(neighboursStr)
		if err == nil {
			fmt.Printf("DEBUG: Neighbours parameter set to '%s'\n", neighboursStr)
		} else {
			fmt.Printf("DEBUG: Neighbours parameter not set\n")
		}

		fmt.Println("Building response data...")
		// Build the response data
		res := []RenewableDataEntry{}
		// Check for parameters
		if country != "" {
			res = BuildResponse(csvData, years, country, neighbours)
		} else {
			res = BuildResponseAll(csvData, years)
		}

		// Check if no data was found for the request
		if country != "" && len(res) < 1 {
			msg := fmt.Sprintf("Your request returned no data, this may be because your supplied country code is invalid or there are no data available for your country code!")
			msg += fmt.Sprintf("\nYour query:\nCountry code = %s, Neighbour flag = %s", country, strconv.FormatBool(neighbours))
			http.Error(w, msg, http.StatusNotFound)
			return
		}

		fmt.Printf("Response built, contains %d entries\n", len(res))

		fmt.Println("Sending JSON...")
		// Send the json
		jsonResponse := json.NewEncoder(w)

		w.Header().Set("Content-Type", "application/json")
		err = jsonResponse.Encode(res)

		if err != nil {
			log.Println("E: There was an error generating the JSON data")
			http.Error(w, "Internal Server Error: There was an error generating the JSON data", http.StatusInternalServerError)
			return
		}

		// Send messages only if a country was specified and not running tests
		if country != "" && flag.Lookup("test.v") == nil {
			// Send a message to the channel if a country was requested
			msg <- strings.ToLower(country)
		}
	}

}

// Determine latest year for each country
func GetLatestYears(data [][]string) (map[string]string, error) {

	latestYear := make(map[string]string)

	for idx, entry := range data {
		// Skip title row
		if idx == 0 {
			continue
		}

		// Skip entities that doesn't have a code (non-countries)
		if len(entry[CSV_COL_CODE]) != 3 {
			continue
		}

		// Get latest year for entry
		year, ok := latestYear[entry[CSV_COL_CODE]]
		if ok {
			// Get year in map and from CSV entry
			year_int, err := strconv.Atoi(year)
			if err != nil {
				log.Println("E: Unable to convert year to integer")
				return nil, errors.New("Unable to convert year to intege")
			}
			new_year, err := strconv.Atoi(entry[CSV_COL_YEAR])
			if err != nil {
				log.Println("E: Unable to convert year to integer")
				return nil, errors.New("Unable to convert year to integer")
			}
			// Check if the new year found is larger
			if new_year > year_int {
				latestYear[strings.ToLower(entry[CSV_COL_CODE])] = entry[CSV_COL_YEAR]
				latestYear[strings.ToLower(entry[CSV_COL_ENTITY])] = entry[CSV_COL_YEAR]
			}
		} else {
			// Save a possible year
			latestYear[strings.ToLower(entry[CSV_COL_CODE])] = entry[CSV_COL_YEAR]
			latestYear[strings.ToLower(entry[CSV_COL_ENTITY])] = entry[CSV_COL_YEAR]
		}
	}

	return latestYear, nil
}

// Determine country/code mapping
func GetCountryCodeMapping(data [][]string) map[string]string {
	mapping := make(map[string]string)
	for _, entry := range data {
		_, ok := mapping[strings.ToLower(entry[CSV_COL_ENTITY])]
		if !ok {
			mapping[strings.ToLower(entry[CSV_COL_ENTITY])] = strings.ToLower(entry[CSV_COL_CODE])
		}
	}
	return mapping
}

// Function that gets the neighbour countries for a given country
func GetNeighbours(country string, code bool) ([]string, error) {
	var res *http.Response
	var err error

	if code {
		// Get data from Countries API
		if flag.Lookup("test.v") == nil {
			res, err = http.Get(COUNTRY_API_BASE_ENDPOINT + "alpha/" + country)
		} else {
			req := httptest.NewRequest(http.MethodGet, "/"+country, nil)
			w := httptest.NewRecorder()
			CountriesStubHandler(w, req)
			res = w.Result()
		}
	} else {
		// Get data from Countries API
		if flag.Lookup("test.v") == nil {
			res, err = http.Get(COUNTRY_API_BASE_ENDPOINT + "name/" + country)
		} else {
			req := httptest.NewRequest(http.MethodGet, "/"+country, nil)
			w := httptest.NewRecorder()
			CountriesStubHandler(w, req)
			res = w.Result()
		}
	}
	if err != nil {
		fmt.Println("E: There was an error contacting the Countries API")
		return nil, err
	}

	// Check for status code
	if res.StatusCode != http.StatusOK {
		fmt.Println("E: Countries API returned " + strconv.Itoa(res.StatusCode))
		return nil, errors.New("Countries API returned" + strconv.Itoa(res.StatusCode))
	}

	// Decode JSON to get neighbours
	jsonData := json.NewDecoder(res.Body)

	data := []CountriesAPICountry{}

	err = jsonData.Decode(&data)
	if err != nil {
		fmt.Println("E: There was an error decoding the data from the Countries API")
		return nil, err
	}

	if len(data) != 1 {
		fmt.Printf("There should only be one country returned, %d countries was returned!", len(data))
		return nil, errors.New("There should only be one country returned, " + strconv.Itoa(len(data)) + " countries was returned!")
	}
	return data[0].Borders, nil
}

// Build response data (for all countries)
func BuildResponseAll(csvData [][]string, years map[string]string) []RenewableDataEntry {
	// Generate the response data
	data := []RenewableDataEntry{}

	for idx, entry := range csvData {
		// Skip title row
		if idx == 0 {
			continue
		}

		// Skip entities that doesn't have a code (non-countries)
		if len(entry[CSV_COL_CODE]) != 3 {
			continue
		}

		// Check if entry has data for latest year
		if entry[CSV_COL_YEAR] == years[strings.ToLower(entry[CSV_COL_CODE])] {
			fmt.Printf("Adding %s, %s...\n", entry[CSV_COL_ENTITY], entry[CSV_COL_YEAR])

			// Parse float
			renewablePercentage, err := strconv.ParseFloat(entry[CSV_COL_RENEWABLES], 64)
			if err != nil {
				log.Println("E: There was an error parsing renewable percentage, probably an error with the dataset")
				// Try next entry
				continue
			}

			// Create the response entry
			newRenewableDataEntry := RenewableDataEntry{
				Name:       entry[CSV_COL_ENTITY],
				ISOCode:    entry[CSV_COL_CODE],
				Year:       entry[CSV_COL_YEAR],
				Percentage: renewablePercentage,
			}

			// Add it
			data = append(data, newRenewableDataEntry)
		}
	}

	return data
}

// Build response data for a single country (and possibly its neighbours)
func BuildResponse(csvData [][]string, years map[string]string, country string, neighbours bool) []RenewableDataEntry {
	// Generate the response data
	data := []RenewableDataEntry{}
	// Allowed countries
	countries := []string{country}

	if neighbours {
		// Get the countries (main + neighbours)
		var neighbourCountries []string
		neighbourCountries, err := GetNeighbours(country, true)
		// Try again, if retrieving using code was unsuccessful
		if err != nil {
			println("E: Failed to retrieve neighbours for country with code")
			neighbourCountries, err = GetNeighbours(country, false)
			if err != nil {
				println("E: Failed to retrieve neighbours for country with name")
				// TODO: Implement proper error return
				return nil
			}
		}

		for _, country := range neighbourCountries {
			countries = append(countries, country)
		}
	}

	// Start looking through the CSV data
	for idx, entry := range csvData {
		// Skip title row
		if idx == 0 {
			continue
		}

		// Skip entities that doesn't have a code (non-countries)
		if len(entry[CSV_COL_CODE]) != 3 {
			continue
		}

		// Skip entity that doesn't match any of the countries
		included := false
		for _, country := range countries {
			if strings.ToLower(entry[CSV_COL_CODE]) == strings.ToLower(country) {
				included = true
			} else if strings.ToLower(entry[CSV_COL_ENTITY]) == strings.ToLower(country) {
				included = true
			}
		}
		if !included {
			continue
		}

		// Check if entry has data for latest year
		if entry[CSV_COL_YEAR] == years[strings.ToLower(entry[CSV_COL_CODE])] ||
			entry[CSV_COL_YEAR] == years[strings.ToLower(entry[CSV_COL_ENTITY])] {

			fmt.Printf("Adding %s, %s...\n", entry[CSV_COL_ENTITY], entry[CSV_COL_YEAR])

			// Parse float
			renewablePercentage, err := strconv.ParseFloat(entry[CSV_COL_RENEWABLES], 64)
			if err != nil {
				log.Println("E: There was an error parsing renewable percentage, probably an error with the dataset")
				// Try next entry
				continue
			}

			// Create the response entry
			newRenewableDataEntry := RenewableDataEntry{
				Name:       entry[CSV_COL_ENTITY],
				ISOCode:    entry[CSV_COL_CODE],
				Year:       entry[CSV_COL_YEAR],
				Percentage: renewablePercentage,
			}

			// Add it
			data = append(data, newRenewableDataEntry)

			// If there is enough entries, do an early return
			if len(data) >= len(countries) {
				return data
			}
		}
	}
	return data
}

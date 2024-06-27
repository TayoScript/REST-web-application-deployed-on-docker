package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func RenewHistoryHandler(msg chan string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			RenewHistoryGet(w, r, msg)
		default:
			http.Error(w, "REST Method '"+r.Method+"' not supported. Currently only '"+http.MethodGet+
				" is supported.", http.StatusNotImplemented)

			return
		}
	}
}

/*
Empty handler as default handler
*/
func RenewHistoryGet(w http.ResponseWriter, r *http.Request, msg chan string) {
	parts := strings.Split(r.URL.Path, "/")
	//if the length of the split is 6. we add an empty string to ensure the bad request does not go out of bounds
	if len(parts) == 6 {
		parts = append(parts, "")
	}

	//Check for bad requests. Bad request means incorrect path format
	if !(len(parts) > 5 && (len(parts) <= 7 && parts[6] == "")) {
		http.Error(w, "unexpected format", http.StatusBadRequest)
		return

	}

	var isoCode string
	// making userInput big leters to compare to csv file
	isoCode = strings.ToUpper(parts[5])

	var begin int
	var end int

	if beginQuery := r.URL.Query().Get("begin"); beginQuery != "" {
		var err error
		begin, err = strconv.Atoi(beginQuery)
		if err != nil {
			http.Error(w, "Invalid parameter must be an integer", http.StatusBadRequest)
			return
		}
	}

	if endQuery := r.URL.Query().Get("end"); endQuery != "" {
		var err error
		end, err = strconv.Atoi(endQuery)
		if err != nil {
			http.Error(w, "Invalid parameter must be an integer", http.StatusBadRequest)
			return
		}
	}

	//Check for bad requests. Bad request means incorrect path format
	if begin == 0 && end == 0 {
		begin = beginYear
		end = endYear
	} else if begin == 0 {
		begin = beginYear
	} else if end == 0 {
		end = endYear
	} else if begin > end {
		http.Error(w, "unexpected format, year inputed is wrong", http.StatusBadRequest)
		return
	}

	// Get sortByValue for url
	var err error
	sortByValueQuery := r.URL.Query().Get("sortByValue")
	sortByValue := false
	if sortByValueQuery != "" {
		sortByValue, err = strconv.ParseBool(sortByValueQuery)
		if err != nil {
			http.Error(w, "Invalid parameter must be a bool value", http.StatusBadRequest)
			return
		}
	}
	// Open the CSV file
	csv, err := ReadCSV(RENEWABLE_DATA_CSV)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// renew history struct
	var rHistory []history
	//map for storing the sum of renewables for each entity also store IsoCode.
	EntitySums := make(map[string]map[string]float64)
	//map for storing the count of columns for each entity also store IsoCode.
	EntityCounts := make(map[string]map[string]int)

	for _, record := range csv {
		Entity := record[0]                                // entity column
		code := record[1]                                  // Code column
		renewables, _ := strconv.ParseFloat(record[3], 64) // renewables column

		// making sure no countries are repeated
		if _, ok := EntitySums[Entity]; !ok {
			EntitySums[Entity] = make(map[string]float64)
			EntityCounts[Entity] = make(map[string]int)
		}

		EntitySums[Entity][code] += renewables // sum of renewables for the entity plus storing IsoCode.
		EntityCounts[Entity][code]++           // incrementing the number of columns for the entity until the next different one

	}

	if isoCode != "" {
		// going thorugh the csv file
		for i := range csv {
			// if the isocode in the csv file is the same as the isocode from the url
			if csv[i][1] == isoCode {
				year, _ := strconv.Atoi(csv[i][2])                 // year column
				renewables, _ := strconv.ParseFloat(csv[i][3], 64) // renewables column
				// get the years
				if year <= end && year >= begin {
					rHistory = append(rHistory, history{
						Entity:     csv[i][0],
						Code:       isoCode,
						Year:       year,
						Percentage: renewables,
					})
				}
			}
		}
	} else {
		for name, sums := range EntitySums { // going through  Entity and sums of renewables
			for code, sum := range sums { // going through IsoCode and sums of renewables
				count := EntityCounts[name][code] // get the number of columns for the Entity and store isoCode
				meanRenewables := (sum) / float64(count)
				rHistory = append(rHistory, history{
					Entity:     name,
					Code:       code,
					Percentage: meanRenewables,
				})
			}
		}
	}
	// sorting by percentage from lowest to highest if " true " is inputted
	if sortByValue == true {
		sort.Slice(rHistory, func(i, j int) bool {
			return rHistory[i].Percentage < rHistory[j].Percentage
		})
	}
	if len(rHistory) == 0 {
		http.Error(w, "iso code not found", http.StatusBadRequest)
		return
	}

	if len(rHistory) != 0 && isoCode != "" {
		msg <- strings.ToLower(isoCode)
	}

	// Set the API response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode rHistory
	err = json.NewEncoder(w).Encode(rHistory)
	if err != nil {
		http.Error(w, "Error during encoding"+err.Error(), http.StatusInternalServerError)
		return
	} else {
		http.Error(w, "", http.StatusNoContent)
	}

}

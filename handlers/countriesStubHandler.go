package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func CountriesStubHandler(w http.ResponseWriter, r *http.Request) {
	// Log requests
	log.Println("Received", r.Method, "request on", r.URL)
	defer log.Println("Finished", r.Method, "request on", r.URL)

	// Check if incoming request method is not get
	if r.Method != http.MethodGet {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	// Get URL part (can support multiple countries)
	urlParts := strings.Split(r.URL.Path, "/")
	country := urlParts[1]

	// Return the appropriate file
	switch country {
	case "nor":
		file, err := GetFile("./res/countries_norway.json")
		if err != nil {
			http.Error(w, "Could not load file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(file))
		break
	case "norway":
		file, err := GetFile("./res/countries_norway.json")
		if err != nil {
			http.Error(w, "Could not load file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(file))
		break
	}

}

func GetFile(filename string) ([]byte, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}

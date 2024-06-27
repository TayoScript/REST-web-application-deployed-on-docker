package main

import (
	"assignment-2/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	// Closing the firestore client
	defer handlers.CloseClient()

	// Find port
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("Port has not been set. Assigning default port:", handlers.DEFAULT_PORT)
		port = handlers.DEFAULT_PORT
	}

	// Setup messaging channel (to have renewable handlers notify the invocation process in the main function)
	msg := make(chan string)

	http.HandleFunc("/", handlers.DefaultHandler)
	http.HandleFunc(handlers.RENEW_HISTORY_ENDPOINT, handlers.RenewHistoryHandler(msg))
	http.HandleFunc(handlers.NOTIFICATION_ENDPOINT, handlers.NotificationHandler)
	http.HandleFunc(handlers.STATUS_ENPOINT, handlers.StatusHandler)

	// Load and generate needed data
	// CSV reading

	data, err := handlers.ReadCSV(handlers.RENEWABLE_DATA_CSV)
	if err != nil {
		log.Println("There was an error reading the file")
		return
	}

	// Get the latest years for each country in the dataset
	years, err := handlers.GetLatestYears(data)

	if err != nil {
		log.Println("There was an error with the data")
		return
	}

	http.HandleFunc(handlers.RENEW_CURRENT_ENDPOINT, handlers.RenewCurrentHandler(data, years, msg))

	// Get code ---> country name mapping for listener
	mapping := handlers.GetCountryCodeMapping(data)

	// Start a listener for messages from handler
	go listener(msg, mapping)

	log.Println("Running on port:", port)

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// Listener for incoming messages from handlers
func listener(msg chan string, mapping map[string]string) {
	// Keeps track of the number of invocations since server start
	invocations := make(map[string]int64)

	for m := range msg {
		country := m
		// Try to see if there is a mapping to a code (name input)
		name, ok := mapping[country]
		if ok {
			country = name
		}

		_, ok = invocations[country]
		if ok {
			invocations[country] += 1
		} else {
			invocations[country] = 1
		}

		handlers.WebhookInvocation(country, int(invocations[country]))
	}
}

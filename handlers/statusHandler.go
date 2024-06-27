package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var startTime time.Time

func uptime() float64 {
	return time.Since(startTime).Seconds()
}

func init() {
	startTime = time.Now()
}

func statusCodeCountry() int {
	r, err := http.Get(COUNTRY_API_ALL_ENDPOINT)
	if err != nil {

	}

	return r.StatusCode
}

func statusCodeNotificationDB() int {
	cols := client.Collections(ctx)
	_, err := cols.Next()
	if err != nil {
		log.Println("Connection to the database is gone")
		return http.StatusInternalServerError
	} else {
		return http.StatusOK
	}
}

func webhooksNotifcation() int {

	alldocs, _ := client.Collection(WEBHOOKS_COLLECTION).Documents(ctx).GetAll()
	return len(alldocs)

}

func StatusHandler(w http.ResponseWriter, r *http.Request) {

	Diagnostics := Diagnostics{
		CountriesApi:   statusCodeCountry(),
		NotificationDb: statusCodeNotificationDB(),
		Webhooks:       webhooksNotifcation(),
		Version:        AppVersion,
		Uptime:         uptime()}
	w.Header().Add("content-type", "application/json")
	encoder := json.NewEncoder(w)

	err := encoder.Encode(Diagnostics)
	if err != nil {
		http.Error(w, "Error during encoding: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

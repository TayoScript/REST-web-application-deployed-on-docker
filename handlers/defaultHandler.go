package handlers

import "net/http"

// The default handler for the webservice.
func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}

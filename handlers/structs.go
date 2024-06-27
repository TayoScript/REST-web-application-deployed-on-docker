package handlers

// Used to hold each entry in output of renewable current endpoint
type RenewableDataEntry struct {
	Name       string  `json:"name"`
	ISOCode    string  `json:"isoCode"`
	Year       string  `json:"year"`
	Percentage float64 `json:"percentage"`
}

// Holds the relevant Countries API data

type CountriesAPICountry struct {
	Borders []string `json:"borders"`
}

type history struct {
	Entity     string  `json:"entity"`
	Code       string  `json:"iso_code,omitempty"`
	Year       int     `json:"year,omitempty"`
	Percentage float64 `json:"percentage"`
}

type Webhook struct {
	URL     string `json:"url"`
	Country string `json:"country"`
	Calls   int    `json:"calls"`
}

type WebhookRegistered struct {
	Webhook_id interface{} `json:"webhook_id"`
	Url        interface{} `json:"url"`
	Country    interface{} `json:"country"`
	Calls      interface{} `json:"calls"`
}

type Notification struct {
	WebhookID string `json:"webhook_id"`
	Country   string `json:"country"`
	Calls     int    `json:"calls"`
}

type Diagnostics struct {
	CountriesApi   int     `json:"countriesapi"`
	NotificationDb int     `json:"notification_db"`
	Webhooks       int     `json:"webhooks"`
	Version        string  `json:"version"`
	Uptime         float64 `json:"uptime"`
}

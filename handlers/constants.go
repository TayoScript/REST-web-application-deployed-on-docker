package handlers

// SETTINGS

// CSV FILE SETTINGS
const RENEWABLE_DATA_CSV = "../renewable-share-energy.csv"

// Firestore credentials
const FIRESTORE_ACCOUNT_KEY = "/credentials/accountkey.json"
const FIRESTORE_ACCOUNT_KEY_LOCAL = "./.credentials/accountkey.json"

// COLUMNS
const CSV_COL_ENTITY = 0
const CSV_COL_CODE = 1
const CSV_COL_YEAR = 2
const CSV_COL_RENEWABLES = 3

const beginYear int = 1965
const endYear int = 2021

const AppVersion = "v1"

// WEBSERVICE ENDPOINTS
const RENEW_CURRENT_ENDPOINT = "/energy/v1/renewables/current/"
const RENEW_HISTORY_ENDPOINT = "/energy/v1/renewables/history/"

// NOTIFICATION_ENDPOINT The endpoint to register a webhook for notifications on countries
const NOTIFICATION_ENDPOINT = "/energy/v1/notifications/"
const STATUS_ENPOINT = "/energy/v1/status/"

// EXTERNAL REST API ENDPOINTS

// COUNTRY_API_ENDPOINT the URL to the country REST API
const COUNTRY_API_BASE_ENDPOINT = "http://129.241.150.113:8080/v3.1/"
const COUNTRY_API_ALL_ENDPOINT = "http://129.241.150.113:8080/v3.1/all"
const COUNTRY_API_ALPHA_ENDPOINT = "http://129.241.150.113:8080/v3.1/alpha/"

// PORTS

// DEFAULT_PORT  The default port given to the web service
const DEFAULT_PORT = "8080"

// COLLECTIONS

// WEBHOOKS_COLLECTION The collection that stores all registered webhooks
const WEBHOOKS_COLLECTION = "webhooks"

// ALL_COUNTRIES_COLLECTION The collection that stores the webhooks not registered to any country
const ALL_COUNTRIES_COLLECTION = "all-countries"

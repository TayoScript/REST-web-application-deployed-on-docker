package handlers

import (
	"bytes"
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
	"strings"
)

var ctx context.Context
var client *firestore.Client
var app *firebase.App

func init() {
	log.Println("Initializing Firebase")
	ctx = context.Background()

	sa := option.WithCredentialsFile(FIRESTORE_ACCOUNT_KEY)
	var err error
	app, err = firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}

func CloseClient() {
	log.Println("Closing firestore client")
	err := client.Close()
	if err != nil {
		log.Fatal("Closing the firebase client failed. Error:", err)
	}
}

func NotificationHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		log.Println("POST method used with notification endpoint")
		notificationPost(w, r)
	case http.MethodGet:
		log.Println("GET method used with notification endpoint")
		notificationGet(w, r)
	case http.MethodDelete:
		log.Println("DELETE method used with notification endpoint")
		notificationDelete(w, r)
	default:
		http.Error(w, "Method "+r.Method+" not supported.", http.StatusMethodNotAllowed)
		return
	}
}

func notificationPost(w http.ResponseWriter, r *http.Request) {
	empty := Webhook{}
	webhook := decodeBody(w, r)
	if webhook == empty {
		return
	}
	if !validateWebhook(w, webhook) {
		return
	}
	// Change the country code to uppercase
	webhook.Country = strings.ToUpper(webhook.Country)
	// Adding the webhook to the 'webhooks' collection which has all registered webhooks.
	id, _, err := client.Collection(WEBHOOKS_COLLECTION).Add(ctx, webhook)
	if err != nil {
		log.Println("Error when adding webhook to the webhook collection. Error: " + err.Error())
		http.Error(w, "Error when adding webhook to the webhook collection. Error "+err.Error(), http.StatusBadRequest)
		return
	}
	// If the webhook is not registering to any specific country then store it in the 'all' collection
	// else store it in the collection belonging to the country it is registering to.
	if webhook.Country == "" {
		_, err := client.Collection(ALL_COUNTRIES_COLLECTION).Doc(id.ID).Create(ctx, webhook)
		if err != nil {
			log.Println("Error when adding webhook to the all collection. Error: " + err.Error())
			http.Error(w, "Error when adding webhook to the all collection.  Error: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		_, err := client.Collection(webhook.Country).Doc(id.ID).Create(ctx, webhook)
		if err != nil {
			log.Println("Error when adding webhook to " + webhook.Country + " collection. Error: " + err.Error())
			http.Error(w, "Error when adding webhook to "+webhook.Country+" collection.  Error: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// The ID
	log.Println("New webhook added to firestore. ID returned: " + id.ID)

	webhookMarshall, err := json.MarshalIndent(webhook, "", " ")
	webhookID, err := json.MarshalIndent(map[string]string{"webhook_id": id.ID}, "", " ")
	log.Println("Webhook:\n" + string(webhookMarshall) + "\nHas been registered.")
	http.Error(w, string(webhookID), http.StatusCreated)
}

func notificationGet(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	ID := parts[4]
	// if no id is given then retrieve all the webhooks
	if ID == "" {
		log.Println("Get all Webhooks")
		iter := client.Collection(WEBHOOKS_COLLECTION).Documents(ctx)
		var webhooks []WebhookRegistered
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatalf("Failed to iterate: %v", err)
			}

			m := doc.Data()
			webh := WebhookRegistered{
				Webhook_id: doc.Ref.ID,
				Url:        m["URL"],
				Country:    m["Country"],
				Calls:      m["Calls"],
			}
			webhooks = append(webhooks, webh)
		}

		w.Header().Add("content-type", "application/json")
		err := json.NewEncoder(w).Encode(webhooks)
		if err != nil {
			log.Println("Error encoding the array of webhooks. Error: ", err.Error())
			http.Error(w, "Error encoding the array of webhooks. Error"+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		log.Println("Get webhook with id:", parts[4])

		res := client.Collection(WEBHOOKS_COLLECTION).Doc(ID)

		doc, err2 := res.Get(ctx)
		if err2 != nil {
			log.Println("Error extracting body of returned document of message " + ID)
			http.Error(w, "Error extracting body of returned document of message "+ID, http.StatusInternalServerError)
			return
		}

		m := doc.Data()
		webhook := WebhookRegistered{
			Webhook_id: doc.Ref.ID,
			Url:        m["URL"],
			Country:    m["Country"],
			Calls:      m["Calls"],
		}

		w.Header().Add("content-type", "application/json")
		err := json.NewEncoder(w).Encode(webhook)
		if err != nil {
			log.Println("Error encoding the array of webhooks. Error: ", err.Error())
			http.Error(w, "Error encoding the array of webhooks. Error"+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func notificationDelete(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id := parts[4]
	if id != "" {
		log.Println("Attempting to delete webhook with ID:", id)

		// Get the webhook document from the 'webhooks' collection
		doc := client.Collection(WEBHOOKS_COLLECTION).Doc(id)
		// Get the snapshot of the document
		docSnap, err := doc.Get(ctx)
		if err != nil {
			errMsg := fmt.Sprintln("Error retrieving webhook with ID:", id)
			if status.Code(err) == codes.NotFound {
				errMsg = fmt.Sprintln(errMsg, "ERROR: There is no webhook with ID:", id)
			} else {
				errMsg = fmt.Sprintln(errMsg, "ERROR: ", err.Error())
			}
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}

		log.Println("Found webhook with ID:", doc.ID)

		country := fmt.Sprint(docSnap.Data()["Country"])
		// Check to see if the webhook was registered to a country
		if country == "" {
			// Delete the webhook from the no-country collection
			doc2 := client.Collection(ALL_COUNTRIES_COLLECTION).Doc(id)
			_, err2 := doc2.Delete(ctx)
			if err2 != nil {
				log.Println("There was an error deleting the webhook from the "+ALL_COUNTRIES_COLLECTION+" collection. ERROR:", err.Error())
				http.Error(w, "There was an error deleting the webhook from the "+ALL_COUNTRIES_COLLECTION+" collection.. ERROR: "+err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			// Delete the webhook from the collection belonging to the country it was registered to
			doc2 := client.Collection(country).Doc(id)
			_, err2 := doc2.Delete(ctx)
			if err2 != nil {
				log.Println("There was an error deleting the webhook from the "+country+" collection. ERROR:", err.Error())
				http.Error(w, "There was an error deleting the webhook from the "+country+" collection.. ERROR: "+err.Error(), http.StatusBadRequest)
				return
			}
		}
		// Delete the webhook from the 'webhooks' collection
		_, err = doc.Delete(ctx)
		if err != nil {
			log.Println("There was an error deleting the webhook. ERROR:", err.Error())
			http.Error(w, "There was an error deleting the webhook. ERROR: "+err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("Successfully deleted webhook")
		http.Error(w, "Successfully deleted webhook", http.StatusOK)
	} else {
		log.Println("An ID to a webhook has to be given")
		http.Error(w, "An ID to a webhook as to be given.", http.StatusBadRequest)
	}
}

func WebhookInvocation(country string, calls int) {
	log.Println("Sending notifications on country:", country)

	// Turn the country code to Uppercase
	country = strings.ToUpper(country)
	// Get the collection of webhook documents belonging to given country
	countryDocs := client.Collection(country).Documents(ctx)
	// Get the collections of webhook documents registered to all countries
	allCountriesDocs := client.Collection(ALL_COUNTRIES_COLLECTION).Documents(ctx)

	// See if any webhook registered to given country should get notified based on its call frequency
	for {
		doc, err := countryDocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate country docs. Error %v", err)
		}

		data := doc.Data()
		notification := Notification{
			WebhookID: doc.Ref.ID,
			Country:   country,
			Calls:     calls,
		}
		callFrequency := int(data["Calls"].(int64))
		url := fmt.Sprint(data["URL"])
		if calls%callFrequency == 0 {
			content, _ := json.MarshalIndent(notification, " ", "")
			go func() {
				_, err := http.Post(url, "application/json", bytes.NewBuffer(content))
				if err != nil {
					log.Println("There was an error sending a POST call to the URL of webhook. ERROR: ", err.Error())
					return
				}
			}()

			//log.Println("Webhook invoked:", notification.WebhookID)
			//fmt.Println("Status message: " + res.Status)
		}
	}
	// See if any webhook registered to all countries should get notified based on its call frequency
	for {
		doc, err := allCountriesDocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate all countries docs. Error %v", err)
		}

		data := doc.Data()
		notification := Notification{
			WebhookID: doc.Ref.ID,
			Country:   country,
			Calls:     calls,
		}
		callFrequency := int(data["Calls"].(int64))
		url := fmt.Sprint(data["URL"])
		if calls%callFrequency == 0 {
			content, _ := json.MarshalIndent(notification, " ", "")
			go func() {
				_, err := http.Post(url, "application/json", bytes.NewBuffer(content))
				if err != nil {
					log.Println("There was an error sending a POST call to the URL of webhook. ERROR: ", err.Error())
					return
				}
			}()
			//log.Println("Webhook invoked:", notification.WebhookID)
			//fmt.Println("Status message: " + res.Status)
		}
	}
}

func decodeBody(w http.ResponseWriter, r *http.Request) Webhook {
	webhook := Webhook{}
	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		demoWebhook := map[string]interface{}{
			"url":     "(string)The URL to be triggered upon an invoked event",
			"country": "(string)The country for which the triggered event applies to (if empty, i.e. \"\", then it applies to any invocation)",
			"calls":   "(int)The number of invocations after which a notification is triggered (has to be 1 or higher)",
		}
		demoMarshall, err2 := json.MarshalIndent(demoWebhook, "", " ")
		if err2 != nil {
			http.Error(w, "Error marshalling the demo webhook", http.StatusInternalServerError)
			return Webhook{}
		}
		http.Error(w, "There was an error decoding the request body: \n\t"+err.Error()+
			"\n\nThe required structure of the webhook:\n"+
			string(demoMarshall)+
			"\nCheck that the structure is correct and resubmit.", http.StatusBadRequest)
		return Webhook{}
	}
	return webhook
}

func validateWebhook(w http.ResponseWriter, webhook Webhook) bool {
	if webhook.URL == "" || webhook.Calls < 1 {
		demoWebhook := map[string]interface{}{
			"url":     "(string)The URL to be triggered upon an invoked event",
			"country": "(string)The country for which the triggered event applies to (if empty, i.e. \"\", then it applies to any invocation)",
			"calls":   "(int)The number of invocations after which a notification is triggered (has to be 1 or higher)",
		}

		demoMarshall, err := json.MarshalIndent(demoWebhook, "", " ")
		if err != nil {
			http.Error(w, "Error marshalling the demo webhook", http.StatusInternalServerError)
			return false
		}

		http.Error(w, "The input body does not fulfill the webhook specification\n"+
			"The required structure of the webhook:\n"+
			string(demoMarshall)+
			"\nCheck that the structure is correct and resubmit.", http.StatusBadRequest)
		return false
	}
	// Validate the country code of the country a webhook is registering to.
	if webhook.Country != "" {
		res, err := http.Get(COUNTRY_API_ALPHA_ENDPOINT + webhook.Country)
		if err != nil {
			log.Println("Error validating the country code.\n\tERROR:", err.Error())
			http.Error(w, "Error validating the country code.\n\tERROR: "+err.Error(), http.StatusBadRequest)
			return false
		}
		if res.StatusCode != http.StatusOK {
			log.Println("ERROR. The country code of the webhook is not a valid country code.")
			http.Error(w, "ERROR. The country code of the webhook is not a valid country code.", http.StatusBadRequest)
			return false
		}
	}

	return true
}

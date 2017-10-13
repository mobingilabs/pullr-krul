package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Event struct {
	CreatedAt      time.Time           `json:"createdAt"`
	Source         string              `json:"source"`
	Payload        string              `json:"payload"`
	PayloadHeaders map[string][]string `json:"payloadHeaders"` // TODO: make it map[string]string, concatenate header values (`;`)
}

const EVENT_SOURCE_GITHUB = "github"
const EVENT_SOURCE_DOCKER_REGISTRY = "docker-registry"

// Dummy events storage for testing, may end up having concurrency issues
var Events = []Event{}

func dockerRegistryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("New registry event arrived...")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// TODO: Better error handling
		log.Fatal(err)
	}

	var notification RegistryNotification

	// TODO: Maybe just pass the payload as it is if the parsing fails
	err = json.Unmarshal(body, &notification)
	if err != nil || len(notification.Events) == 0 {
		log.Println("Event payload is not right...")
		w.WriteHeader(400)
		return
	}

	for _, event := range notification.Events {
		eventJson, err := json.Marshal(event)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
			w.WriteHeader(500)
			// TODO: Better error handling
			return
		}

		// TODO: Maybe not to send all headers?
		Events = append(Events, Event{CreatedAt: time.Now(), Payload: string(eventJson), PayloadHeaders: r.Header, Source: EVENT_SOURCE_DOCKER_REGISTRY})
	}

	w.WriteHeader(200)
}

func githubHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("New github event arrived...")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		w.WriteHeader(500)
		// TODO: Better error handling
		return
	}

	Events = append(Events, Event{CreatedAt: time.Now(), Payload: string(body), PayloadHeaders: r.Header, Source: EVENT_SOURCE_GITHUB})

	w.WriteHeader(200)
}

func index(w http.ResponseWriter, r *http.Request) {
	eventsJson, err := json.Marshal(Events)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, eventsJson)
}

func main() {
	http.HandleFunc("/registry", dockerRegistryHandler)
	http.HandleFunc("/github", githubHandler)
	http.HandleFunc("/", index)
	http.ListenAndServe(":3000", nil)
}

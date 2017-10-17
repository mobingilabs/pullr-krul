package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Event is the common event wrapper structure for the events coming from sources
// like github, docker registry and others
type Event struct {
	CreatedAt      time.Time           `json:"createdAt"`
	Source         string              `json:"source"`
	Payload        string              `json:"payload"`
	PayloadHeaders map[string][]string `json:"payloadHeaders"` // TODO: make it map[string]string, concatenate header values (`;`)
}

const eventSourceGithub = "github"
const eventSourceDockerRegistry = "docker-registry"

// Dummy events storage for testing, may end up having concurrency issues
var events = []Event{}

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
		log.Println("event payload is not right...")
		w.WriteHeader(400)
		return
	}

	for _, event := range notification.Events {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
			w.WriteHeader(500)
			// TODO: Better error handling
			return
		}

		// TODO: Maybe not to send all headers?
		events = append(events, Event{CreatedAt: time.Now(), Payload: string(eventJSON), PayloadHeaders: r.Header, Source: eventSourceDockerRegistry})
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

	events = append(events, Event{CreatedAt: time.Now(), Payload: string(body), PayloadHeaders: r.Header, Source: eventSourceGithub})

	w.WriteHeader(200)
}

func index(w http.ResponseWriter, r *http.Request) {
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(eventsJSON)
}

func main() {
	http.HandleFunc("/registry", LogRequest("dockerRegistryHandler", dockerRegistryHandler))
	http.HandleFunc("/github", LogRequest("githubHandler", githubHandler))
	http.HandleFunc("/", LogRequest("index", index))

	hostport := "0.0.0.0:80"
	log.Printf("Krul start listening at %v...", hostport)
	log.Fatal(http.ListenAndServe(hostport, nil))
}

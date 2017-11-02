package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
)

//go:generate go run gen_version.go

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

func main() {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/registry", LogRequest("dockerRegistryHandler", dockerRegistryHandler)).Methods("POST")
	api.HandleFunc("/github", LogRequest("githubHandler", githubHandler)).Methods("POST")
	api.HandleFunc("/version", LogRequest("version", version))
	api.HandleFunc("/", LogRequest("index", index))
	r.PathPrefix("/").HandlerFunc(LogRequest("404 Not Found", notFound))

	http.Handle("/", r)

	hostport := "0.0.0.0:80"
	log.Printf("Krul start listening at %v...", hostport)
	log.Fatal(http.ListenAndServe(hostport, nil))
}

func ok(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\": 200}")
}

func dockerRegistryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("New registry event arrived...")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		// TODO: Better error handling
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var notification RegistryNotification

	// TODO: Maybe just pass the payload as it is if the parsing fails
	err = json.Unmarshal(body, &notification)
	if err != nil || len(notification.Events) == 0 {
		log.Println("event payload is not right...")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	for _, event := range notification.Events {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
			// TODO: Better error handling
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		// TODO: Maybe not to send all headers?
		events = append(events, Event{CreatedAt: time.Now(), Payload: string(eventJSON), PayloadHeaders: r.Header, Source: eventSourceDockerRegistry})
	}

	ok(w)
}

func githubHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("New github event arrived...")

	validWebhook := validateGithubWebhook(r)
	if !validWebhook {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
	r.Body.Close()

	event := githubEvent(r)
	switch event {
	case PushEvent:
		var event github.PushEvent
		if err := json.Unmarshal(body, &event); err != nil {
			log.Printf("ERROR: Couldn't parse push event payload, %v\n", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		// TODO: Validate/Check the pusher (ALM account?) or secretkey/url can be unique to user?

		// TODO: We should better give responsibility of finding docker file and sending a message to icarium
		// to another goproc because github has request timeout set to 10 seconds. It seems better to
		// return a response asap.
		// (see: https://developer.github.com/changes/2017-09-12-changes-to-maximum-webhook-timeout-period/)
		githubToken := "3733101557ce4f040918e052db6370ab44b63b92"
		repositoryFullname := *event.Repo.FullName
		commitHash := *event.HeadCommit.TreeID
		dockerfileUrl := fmt.Sprintf("https://%s:x-oauth-basic@raw.githubusercontent.com/%s/%s/Dockerfile", githubToken, repositoryFullname, commitHash)
		response, err := http.Get(dockerfileUrl)
		if err != nil {
			log.Printf("Failed to check Dockerfile for the repository %v, %v\n", repositoryFullname, err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		body, err := ioutil.ReadAll(response.Body)
		if err == nil {
			log.Printf("Dockerfile check response: (%v) %v", response.StatusCode, string(body))
		}

		dockerFileExists := response.StatusCode >= 200 && response.StatusCode < 300
		if dockerFileExists {
			log.Printf("Dispatching build action for %s...\n", repositoryFullname)
			// TODO: Dispatch build action on queue
		}

		log.Println("No Dockerfile found...")
	default:
		log.Printf("Unknown github event: \"%v\"...\n", event)
	}

	events = append(events, Event{CreatedAt: time.Now(), Payload: string(body), PayloadHeaders: r.Header, Source: eventSourceGithub})
	ok(w)
}

func index(w http.ResponseWriter, r *http.Request) {
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(eventsJSON)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func version(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{\"version\": \"%s\"}", Version)
}

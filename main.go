package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
)

//go:generate go run gen_version.go

func main() {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/github", LogRequest("githubHandler", githubHandler)).Methods("POST")
	api.HandleFunc("/version", LogRequest("version", version))

	http.Handle("/", r)

	hostport := "0.0.0.0:80"
	log.Printf("Krul start listening at %v...", hostport)
	log.Fatal(http.ListenAndServe(hostport, nil))
}

func githubHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("New github event arrived...")

	isGithubRequest := strings.HasPrefix(r.Header.Get("User-Agent"), "GitHub-Hookshot")
	if !isGithubRequest {
		log.Printf("Invalid github webhook request")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	event := r.Header.Get("X-Github-Event")
	switch event {
	case "push":
		var event github.PushEvent
		if err := json.Unmarshal(body, &event); err != nil {
			log.Printf("ERROR: Couldn't parse push event payload, %v\n", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		repositoryFullname := *event.Repo.FullName
		username, err := getUsernameByRepository("github", repositoryFullname)
		if err != nil {
			log.Printf("Failed to get username from the repository %s, ERROR: %v", repositoryFullname, err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		if username == "" {
			log.Printf("Repository %s is not on watch list, skipping building...", repositoryFullname)
			triggerPullrCallback("push", "payload")
			ok(w)
			return
		}

		githubToken, err := getGithubTokenByUsername(username)
		if err != nil {
			log.Printf("Failed to get github token for user: %s, skipping building... ERROR: %v", username, err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		if githubToken == "" {
			log.Printf("Github token for user \"%s\" not found.", username)
			triggerPullrCallback("push", "payload")
			ok(w)
			return
		}

		commitHash := *event.After
		dockerfileExists, err := checkFileExists(repositoryFullname, "Dockerfile", commitHash, githubToken)
		if err != nil {
			log.Printf("Failed to check Dockerfile for the repository %v, %v\n", repositoryFullname, err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		if dockerfileExists {
			log.Printf("Dispatching build action for %s...\n", repositoryFullname)
			// TODO: Dispatch build action on queue
		}
	default:
		log.Printf("Unknown github event: \"%v\"...\n", event)
	}

	// TODO: Do we have callbacks other than github pushes? If so we should trigger it here too
	ok(w)
}

func triggerPullrCallback(action, payload string) {
	log.Println("Will trigger pullr callback if exists")
}

func version(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{\"version\": \"%s\"}", Version)
}

func ok(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\": 200}")
}

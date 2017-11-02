package main

import (
	"net/http"
	"strings"
)

const (
	PushEvent       = "push"
	UserAgentPrefix = "GitHub-Hookshot"
)

func githubEvent(r *http.Request) string {
	event, ok := r.Header["X-Github-Event"]
	if !ok {
		return ""
	}

	return event[0]
}

func validateGithubWebhook(r *http.Request) bool {
	userAgent, ok := r.Header["User-Agent"]

	if !ok || !strings.HasPrefix(userAgent[0], UserAgentPrefix) {
		return false
	}

	return true
}

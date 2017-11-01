package main

import (
	"net/http"
	"strings"
)

const (
	PushEvent       = "push"
	UserAgentPrefix = "GitHub-Hookshot"
)

func headerValue(r *http.Request, key string) string {
	header, ok := r.Header[key]
	if !ok {
		return ""
	}

	first := header[0]
	if !ok {
		return ""
	}

	return first
}

func githubEvent(r *http.Request) string {
	return headerValue(r, "X-GitHub-Event")
}

func validateGithubWebhook(r *http.Request) bool {
	userAgent := headerValue(r, "User-Agent")
	if !strings.HasPrefix(userAgent, UserAgentPrefix) {
		return false
	}

	return true
}

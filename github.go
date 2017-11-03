package main

import (
	"fmt"
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

func checkFileExists(repositoryFullname, path, ref, token string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s?ref=%s", repositoryFullname, path, ref)
	request, err := http.NewRequest("HEAD", url, nil)
	request.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return false, err
	}

	if res.StatusCode == 200 {
		return true, nil
	}

	return false, nil
}

package main

import (
	"fmt"
	"log"
	"net/http"
)

// checkFileExists checks if the given path exist in the given repository
// by the given commit ref
func checkFileExists(repositoryFullname, path, ref, token string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s?ref=%s", repositoryFullname, path, ref)
	request, err := http.NewRequest("HEAD", url, nil)
	request.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	log.Printf("Checking dockerfile at %s\n", url)
	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return false, err
	}
	res.Body.Close()

	log.Printf("Check return status code: %v", res.StatusCode)
	if res.StatusCode == 200 {
		return true, nil
	}

	return false, nil
}

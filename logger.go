package main

import (
	"log"
	"net/http"
	"os"
)

// LogRequest logs incoming requests
// TODO: Use http.Handler maybe
func LogRequest(handlerName string, handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	// TODO: We don't need a new Logger for every handler(?)
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("[req] %v %v => %v\n", r.Method, r.URL, handlerName)
		handler(w, r)
	}
}

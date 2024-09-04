package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s\n", r.Method, r.URL.Path)

	switch r.URL.Path {
	case "/index.html", "/":
		http.ServeFile(w, r, "/static/index.html")
	default:
		log.Printf("%s %s\n", r.URL.Path, http.StatusNotFound)
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	// Set HTTP listening port with HTTP_PORT environment variable
	port, ok := os.LookupEnv("HTTP_PORT")
	if !ok { panic("ERROR: No HTTP_PORT environment variable set.") }

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

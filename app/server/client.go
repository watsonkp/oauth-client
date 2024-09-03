package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Set HTTP listening port with HTTP_PORT environment variable
	port, ok := os.LookupEnv("HTTP_PORT")
	if !ok { panic("ERROR: No HTTP_PORT environment variable set.") }

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

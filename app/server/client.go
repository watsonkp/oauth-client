package main

import (
	"encoding/base64"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func randomState() string {
	c := 32
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("ERROR: Failed to generate random state (CSRF). %v", err)
	}

	return base64.StdEncoding.EncodeToString(b)
}

func notFoundHandlerFunc(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v %v\n", r.URL.Path, http.StatusNotFound)
	w.WriteHeader(http.StatusNotFound)
}

func makeCreateAuthorizationLinkHandler() http.HandlerFunc {
	// Configure the API's authorization endpoint URI
	endpoint, ok := os.LookupEnv("AUTHORIZATION_ENDPOINT")
	if !ok {
		log.Println("WARNING: No authorization endpoint link generation at /state. Missing required AUTHORIZATION_ENDPOINT environment variable.")
		return notFoundHandlerFunc
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Printf("WARNING: No authorization endpoint link generation at /state. Misconfigured AUTHORIZATION_ENDPOINT environment variable. %v for %v\n", err, endpoint)
		return notFoundHandlerFunc
	}
	q := u.Query()

	// Configure the application's registered client ID
	id, ok := os.LookupEnv("CLIENT_ID")
	if !ok {
		log.Println("WARNING: No authorization endpoint link generation at /state. Missing required CLIENT_ID environment variable.")
		return notFoundHandlerFunc
	}
	q.Set("client_id", id)

	// Configure the application's requested API resource scope
	scope, ok := os.LookupEnv("SCOPE")
	if !ok {
		log.Println("WARNING: No authorization endpoint link generation at /state. Missing required SCOPE environment variable.")
		return notFoundHandlerFunc
	}
	q.Set("scope", scope)

	// Configure the application's registered redirect URI that the API server will
	//  redirect the resource owner to when they authorize the application.
	// Static for security. Keep it simple. Don't be clever.
	redirect, ok := os.LookupEnv("REDIRECT_URI")
	if !ok {
		log.Println("WARNING: No authorization endpoint link generation at /state. Missing required REDIRECT_URI environment variable.")
		return notFoundHandlerFunc
	}
	redirectURI, err := url.Parse(redirect)
	if err != nil {
		log.Printf("WARNING: No authorization endpoint link generation at /state. Misconfigured REDIRECT_URI environment variable. %v for %v\n", err, redirect)
		return notFoundHandlerFunc
	}
	q.Set("redirect_uri", redirectURI.String())

	q.Set("response_type", "code")
	u.RawQuery = q.Encode()

	return func(w http.ResponseWriter, r *http.Request) {
		q := u.Query()
		q.Set("state", randomState())
		u.RawQuery = q.Encode()
		fmt.Fprintf(w, "<a href=\"%s\">Authorize Application</a>", u.String())
	}
}

func makeRequestResourceOwnerAccessTokenHandler() http.HandlerFunc {
	// Configure the API's token endpoint URI
	endpoint, ok := os.LookupEnv("TOKEN_ENDPOINT")
	if !ok {
		log.Println("WARNING: No resource owner access token request endpoint at /authorized. Missing required TOKEN_ENDPOINT environment variable.")
		return notFoundHandlerFunc
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Printf("WARNING: No resource owner access token request endpoint at /authorized. Misconfigured TOKEN_ENDPOINT environment variable. %v for %v\n", err, endpoint)
		return notFoundHandlerFunc
	}

	// Configure the application's registered client ID
	id, ok := os.LookupEnv("CLIENT_ID")
	if !ok {
		log.Println("WARNING: No resource owner access token request endpoint at /authorized. Missing required CLIENT_ID environment variable.")
		return notFoundHandlerFunc
	}

	// Configure the application's registered client secret
	secret, ok := os.LookupEnv("CLIENT_SECRET")
	if !ok {
		log.Println("WARNING: No resource owner access token request endpoint at /authorized. Missing required CLIENT_SECRET environment variable.")
		return notFoundHandlerFunc
	}

	// Configure the application's registered redirect URI that the API server will
	//  redirect the resource owner to when they authorize the application.
	// Static for security. Keep it simple. Don't be clever.
	redirect, ok := os.LookupEnv("REDIRECT_URI")
	if !ok {
		log.Println("WARNING: No resource owner access token request endpoint at /authorized. Missing required REDIRECT_URI environment variable.")
		return notFoundHandlerFunc
	}
	redirectURI, err := url.Parse(redirect)
	if err != nil {
		log.Printf("WARNING: No resource owner access token request endpoint at /state. Misconfigured REDIRECT_URI environment variable. %v for %v\n", err, redirect)
		return notFoundHandlerFunc
	}

	form := url.Values { "redirect_uri": { redirectURI.String() }, "grant_type": { "authorization_code" } }

	return func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query()
		if ok := v.Has("state"); !ok {
			w.WriteHeader(http.StatusUnauthorized)
		}
		code := v.Get("code")

		form.Set("code", code)

		req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(id, secret)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("ERROR: While requesting resource owner access token. %v\n", err)
			w.WriteHeader(http.StatusUnauthorized)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ERROR: While reading resource owner access token response body. %v\n", err)
			w.WriteHeader(http.StatusUnauthorized)
		}

		fmt.Fprint(w, string(body))
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v %v\n", r.Method, r.URL.Path)

	switch r.URL.Path {
	case "/index.html", "/":
		http.ServeFile(w, r, "/static/index.html")
	case "/authorized":
		http.ServeFile(w, r, "/static/index.html")
	default:
		log.Printf("%v %v\n", r.URL.Path, http.StatusNotFound)
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	// Set HTTP listening port with HTTP_PORT environment variable
	port, ok := os.LookupEnv("HTTP_PORT")
	if !ok { panic("ERROR: No HTTP_PORT environment variable set.") }

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/state", makeCreateAuthorizationLinkHandler())
	http.HandleFunc("/token", makeRequestResourceOwnerAccessTokenHandler())
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

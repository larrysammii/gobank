package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APISever struct {
	listenAddr string
	store      Storage
}

func NewAPISever(listenAddr string, store Storage) *APISever {
	return &APISever{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APISever) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", makeHttpHandleFunc(s.handleGetAccount))

	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APISever) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APISever) handleGetAccount(w http.ResponseWriter, r *http.Request) error {

	id := mux.Vars(r)["id"]
	// .Vars take the request and return a map of the variables that are in the request
	// for example: /account/1234
	// id = 1234

	fmt.Println(id)

	return WriteJSON(w, http.StatusOK, &Account{})
}

func (s *APISever) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APISever) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APISever) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Output the response
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// I dont get it
	return json.NewEncoder(w).Encode(v)
}

// Define the apiFunc type
// Function alias is a way to give a new name to an existing type or function
// The apiFunc type is a function that takes an http.ResponseWriter and an http.Request and returns an error
// Honestly I dont get it...
type apiFunc func(http.ResponseWriter, *http.Request) error

// Define the ApiError struct
type ApiError struct {
	Error string
}

// Define the makeHttpHandleFunc function
func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	// Why does it have to take apiFunc?
	// Because the http.HandlerFunc type is a function that takes an http.ResponseWriter and an http.Request and returns nothing
	// The makeHttpHandleFunc function takes an apiFunc and returns an http.HandlerFunc
	// The returned http.HandlerFunc calls the apiFunc and handles the error
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			// handle the error
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}

	}
}

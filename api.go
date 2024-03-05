package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type APISever struct {
	//
	// listenAddr is the address the server will listen on
	// store is the storage interface the server will use to store data
	//
	listenAddr string
	store      Storage
}

func NewAPISever(listenAddr string, store Storage) *APISever {
	return &APISever{
		// return a pointer to the APISever struct
		// & means the address of the struct
		// * means the value of the address
		// Points to the address of the struct, then the function chanages the value of the struct,
		// then returns the address to get the modified struct

		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APISever) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHttpHandleFunc(s.handleGetAccountByID), s.store))
	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer))
	// Could use subrouter to group routes?

	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
	// http.ListenAndServe creates a TCP listener on the specified address and port and then starts a loop that accepts incoming connections. For each incoming connection, it creates a new goroutine that handles the connection. The goroutine reads the request from the connection, passes it to the handler, and then writes the response back to the connection.
}

// -------------------------------------------------------------------------------------
// The Big Function that handles all the requests
func (s *APISever) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

// -------------------------------------------------------------------------------------

// GET /account/{id}
func (s *APISever) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := getID(r)
		if err != nil {
			return err
		}
		// .Vars take the request and return a map of the variables that are in the request
		// for example: /account/1234
		// id = 1234

		account, err := s.store.GetAccountByID(id)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, account)
		// it will only write the id on the response
	}
	// fmt.Println(id)

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)

}

// GET /account
// GET /accounts
func (s *APISever) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	// if GetAccounts() doesn't return an error,
	// then it writes the response to the client.
	return WriteJSON(w, http.StatusOK, accounts)
}

// POST /account
func (s *APISever) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	// createAccountReq := CreateAccountRequest{}
	// 	if err := json.NewDecoder(r.Body).Decode(&createAccountReq); err != nil {
	// 	return err
	// }
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}

	account := NewAccount(createAccountReq.FirstName, createAccountReq.LastName)

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	tokenString, err := createJWT(account)
	if err != nil {

		return err

	}
	fmt.Println("JWT token: ", tokenString)
	return WriteJSON(w, http.StatusOK, account)
}

// DELETE /account/{id}
func (s *APISever) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

// POST /transfer
func (s *APISever) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close() // Do I need this?

	return WriteJSON(w, http.StatusOK, transferReq)
}

// Output the response
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// I dont get it
	return json.NewEncoder(w).Encode(v)
}

// Define the ApiError struct
type ApiError struct {
	Error string `json:"error"`
}

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")

		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)

		if err != nil {
			permissionDenied(w)
			return
		}

		if !token.Valid {
			permissionDenied(w)
			return
		}

		userID, err := getID(r)
		if err != nil {
			permissionDenied(w)
			return
		}

		account, err := s.GetAccountByID(userID)
		if err != nil {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.Number != int64(claims["accountNumber"].(float64)) { // lmao
			permissionDenied(w)
			return
		}

		if err != nil {
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token"})
			return
		}

		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

// Define the apiFunc type
// Function alias is a way to give a new name to an existing type or function
// The apiFunc type is a function that takes an http.ResponseWriter and an http.Request and returns an error
type apiFunc func(http.ResponseWriter, *http.Request) error

// I guess when you check the type for the handler functions it will be used to check if the function is a valid handler function?

// Turns the apiFunc into a http.HandlerFunc,
// so that apiFunc can use the http.ResponseWriter and http.Request
func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}

	}
}

func getID(r *http.Request) (int, error) {

	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id %s", idStr)
	}
	return id, nil

}

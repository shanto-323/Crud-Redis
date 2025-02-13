package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Api struct {
	ipAddr string
	store  Store
}

func NewApi(ipAddr string, store Store) *Api {
	return &Api{
		ipAddr: ipAddr,
		store:  store,
	}
}

func (a *Api) Start() {
	mux := mux.NewRouter()

	mux.HandleFunc("/user", makeHttpHandleFunc(a.handleAccount))

	fmt.Println("Api Starting....")
	if err := http.ListenAndServe(a.ipAddr, mux); err != nil {
		log.Fatal(err)
	}
}

func (a *Api) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodPost {
		return a.CreateAcount(w, r)
	}
	if r.Method == http.MethodGet {
		return a.GetAccounts(w, r)
	}
	return fmt.Errorf("mathod not allowed %s", r.Method)
}

func (a *Api) CreateAcount(w http.ResponseWriter, r *http.Request) error {
	userReq := &NewUserReq{}
	if err := json.NewDecoder(r.Body).Decode(userReq); err != nil {
		return err
	}

	user := NewUser(userReq.Name)
	if err := a.store.CreateAccount(user); err != nil {
		return err
	}
	writeJson(w, http.StatusOK, user)
	return nil
}

func (a *Api) GetAccounts(w http.ResponseWriter, r *http.Request) error {
	respose, err := a.store.GetAccounts()
	if err != nil {
		return err
	}
	return writeJson(w, http.StatusOK, respose)
}

func writeJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func makeHttpHandleFunc(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJson(w, http.StatusBadRequest, err)
		}
	}
}

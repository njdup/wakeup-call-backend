package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/njdup/wakeup-call-backend/conf"

	"./models/users"
)

func HomeHandler(res http.ResponseWriter, req *http.Request) {
	user.TestInsert()
	fmt.Fprintf(res, "This is a test!")
}

// ConfigureRoutes sets all API routes
func configureRoutes(router *mux.Router) {
	router.HandleFunc("/test/", HomeHandler)
}

// Main launches the API server
func main() {
	router := mux.NewRouter()
	configureRoutes(router)

	http.Handle("/", router)
	http.ListenAndServe(config.Settings.Port, nil)
}
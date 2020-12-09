package statusbar

import (
	"encoding/json"
	"errors"
	"net/http"
)

type failure struct {
	Message string `json:"error"`
}

// listen starts the REST API microservice and handles requests.
func listen(rs []routine) {
	// First, let's build out a map of routines for easy routing later.
	routines := make(map[string]routine, len(rs))
	for _, v := range rs {
		routines[v.pkgName] = v
	}

	// Add a handler for the REST API.
	http.HandleFunc("/rest/v1", handleRest)

	// Start listening on our REST port.
	err := http.ListenAndServe(":3777", nil)
	if err == nil {
		err = errors.New("REST API down")
	}
	logError(err.Error())
}

func handleRest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGet(w, r)
	case "PUT":
		handlePut(w, r)
	case "DELETE":
		handleDelete(w, r)
	default:
		e := json.NewEncoder(w)
		e.Encode(failure{"Invalid method"})
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
}

func handlePut(w http.ResponseWriter, r *http.Request) {
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
}

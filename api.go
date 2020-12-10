package statusbar

import (
	"encoding/json"
	"errors"
	"io"
	"log"
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
	log.Printf(err.Error())
}

func handleRest(w http.ResponseWriter, r *http.Request) {
	var resp string
	switch r.Method {
	case "GET":
		resp = handleGet(r)
	case "PUT":
		resp = handlePut(r)
	case "DELETE":
		resp = handleDelete(r)
	default:
		b, err := json.Marshal(failure{"Invalid method"})
		if err != nil {
			resp = "Error: " + err.Error()
		} else {
			resp = string(b)
		}
	}

	io.WriteString(w, resp)
}

func handleGet(r *http.Request) string {
	return "GET"
}

func handlePut(r *http.Request) string {
	return "PUT"
}

func handleDelete(r *http.Request) string {
	return "DELETE"
}

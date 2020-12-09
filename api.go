package statusbar

import (
	"fmt"
	"net/http"
	"log"
)

// listen starts the REST API microservice and handles requests.
func listen(rs []routine) {
	// First, let's build out a map of routines for easy routing later.
	routines := make(map[string]routine, len(rs))
	for _, v := range rs {
		routines[v.pkgName] = v
	}

	http.HandleFunc("/api/v1", test)
	log.Fatal(http.ListenAndServe(":6777", nil))
}

func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "It works")
}

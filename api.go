package statusbar

import (
)

// listen starts the REST API microservice and handles requests.
func listen(rs []routine) {
	// First, let's build out a map of routines for easy selection later.
	routines := make(map[string]routine, len(rs))
	for _, v := range rs {
		routines[v.pkgName] = v
	}
}


// This file contains the callbacks that handle each of the REST API endpoints.

package statusbar

import (
	"encoding/json"
	"fmt"
	"github.com/snhilde/statusbar/v5/restapi"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// apiHandler is a wrapper object for convenience reasons: in order for the restapi package to be able to use the
// handlers belonging to the object passed to it, all handler methods must be exported. However, we don't want them
// showing up in the auto-docs, so we'll wrap up the main statusbar object with an apiHandler to retain all
// functionality but not flood the docs with a bunch of handler methods.
type apiHandler struct {
	*Statusbar
}

// routineInfo holds the information that is returned for each routine query.
type routineInfo struct {
	// Routine's display name.
	Name string `json:"name"`

	// How long the routine has been active, in seconds. If the routine is inactive, then this is 0.
	Uptime int `json:"uptime"`

	// Interval time between update runs, in seconds.
	Interval int `json:"interval"`

	// Whether or not the routine is currently active.
	Active bool `json:"active"`
}

// HandleGetPing responds to a ping request with "pong".
// endpoint: GET /ping
func (a apiHandler) HandleGetPing(endpoint restapi.Endpoint, params restapi.Params, request *http.Request) (int, string) {
	return 200, "pong"
}

// HandleGetEndpoints returns a JSON object of all possible v1 endpoints and their descriptions.
// endpoint: GET /endpoints
func (a apiHandler) HandleGetEndpoints(endpoint restapi.Endpoint, params restapi.Params, request *http.Request) (int, string) {
	// Open up our spec file for reading.
	file, err := os.Open("api_specs/restv1.json")
	if err != nil {
		return 500, encodePair("error", err.Error())
	}

	// Unmarshal JSON in file.
	spec := restapi.RestSpec{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&spec); err != nil && err != io.EOF {
		return 500, encodePair("error", err.Error())
	}

	// Go through the spec and read all the specified endpoints.
	endpoints := make([]map[string]string, 0)
	for _, table := range spec.Tables {
		for _, endpoint := range table.Endpoints {
			e := map[string]string{
				"method":      endpoint.Method,
				"url":         endpoint.URL,
				"description": endpoint.Desc,
			}
			endpoints = append(endpoints, e)
		}
	}

	return 200, encodePair("endpoints", endpoints)
}

// HandleGetRoutineAll responds with information about all the routines (active and inactive).
// endpoint: GET /routines
func (a apiHandler) HandleGetRoutineAll(endpoint restapi.Endpoint, params restapi.Params, request *http.Request) (int, string) {
	infos := make(map[string]routineInfo, 0)
	for _, routine := range a.routines {
		name := routine.moduleName()
		info := getRoutineInfo(routine)
		infos[name] = info
	}

	return 200, encodePair("routines", infos)
}

// HandleGetRoutine responds with information about the specified routine.
// endpoint: GET /routines/:routine
func (a apiHandler) HandleGetRoutine(endpoint restapi.Endpoint, params restapi.Params, request *http.Request) (int, string) {
	routine, err := getRoutine(a.routines, params["routine"])
	if err != nil {
		return 400, encodePair("error", err.Error())
	}

	return 200, encodePair(routine.moduleName(), getRoutineInfo(routine))
}

// HandlePutRoutineAll restarts all active routines.
// endpoint: PUT /routines
func (a apiHandler) HandlePutRoutineAll(endpoint restapi.Endpoint, params restapi.Params, request *http.Request) (int, string) {
	for _, routine := range a.routines {
		if routine.isActive() {
			select {
			case routine.updateChan <- struct{}{}:
			default:
				return 500, encodePair("error", "failure")
			}
		}
	}

	return 204, ""
}

// HandlePutRoutine restarts the specified routine.
// endpoint: PUT /routines/:routine
func (a apiHandler) HandlePutRoutine(endpoint restapi.Endpoint, params restapi.Params, request *http.Request) (int, string) {
	routine, err := getRoutine(a.routines, params["routine"])
	if err != nil {
		return 400, encodePair("error", err.Error())
	}

	if routine.isActive() {
		select {
		case routine.updateChan <- struct{}{}:
		default:
			return 500, encodePair("error", "failure")
		}
	}

	return 204, ""
}

// HandlePatchRoutine updates the specified routine's data. Currently, this only updates the interval time.
// endpoint: PATCH /routines/:routine
func (a apiHandler) HandlePatchRoutine(endpoint restapi.Endpoint, params restapi.Params, request *http.Request) (int, string) {
	routine, err := getRoutine(a.routines, params["routine"])
	if err != nil {
		return 400, encodePair("error", err.Error())
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return 400, encodePair("error", err.Error())
	}

	if len(body) == 0 {
		return 400, encodePair("error", "missing request body")
	}

	// Set the default interval to -1 so we know if a new interval was passed in or not.
	info := routineInfo{Interval: -1}
	if err := json.Unmarshal(body, &info); err != nil {
		return 400, encodePair("error", err.Error())
	}

	if info.Interval >= 0 {
		routine.setInterval(info.Interval)
	}

	// Let's also trigger an update in case the interval time is now up.
	if routine.isActive() {
		select {
		case routine.updateChan <- struct{}{}:
		default:
			return 500, encodePair("error", "failure")
		}
	}

	return 202, ""
}

// HandleDeleteRoutineAll stops all routines (and therefore the statusbar and API engine).
// endpoint: DELETE /routines
func (a apiHandler) HandleDeleteRoutineAll(endpoint restapi.Endpoint, params restapi.Params, request *http.Request) (int, string) {
	for _, routine := range a.routines {
		if routine.isActive() {
			if !routine.stop(5) {
				return 500, encodePair("error", "failure")
			}
		}
	}

	return 204, ""
}

// HandleDeleteRoutine stops the specified routine.
// endpoint: DELETE /routines/:routine
func (a apiHandler) HandleDeleteRoutine(endpoint restapi.Endpoint, params restapi.Params, request *http.Request) (int, string) {
	routine, err := getRoutine(a.routines, params["routine"])
	if err != nil {
		return 400, encodePair("error", err.Error())
	}

	if routine.isActive() {
		if !routine.stop(5) {
			return 500, encodePair("error", "failure")
		}
	}

	return 204, ""
}

// getRoutine is a helper function that gets the specified routine from the list of routines.
func getRoutine(routines []*routine, name string) (*routine, error) {
	for _, routine := range routines {
		if name == routine.moduleName() {
			return routine, nil
		}
	}

	return nil, fmt.Errorf("invalid routine")
}

// encodePair is a helper function that JSON-encodes a key/value pair.
func encodePair(key string, value interface{}) string {
	pair := map[string]interface{}{
		key: value,
	}

	b, _ := json.Marshal(pair)
	return string(b)
}

// getRoutineInfo returns the routine's information.
func getRoutineInfo(r *routine) routineInfo {
	if r != nil {
		return routineInfo{
			Name:     r.displayName(),
			Uptime:   r.uptime(),
			Interval: r.interval(),
			Active:   r.isActive(),
		}
	}
	return routineInfo{}
}

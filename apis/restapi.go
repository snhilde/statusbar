// This file holds the logic concerning the structuring and running of the REST API using the gin framework.
package statusbar

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// restApi holds the information about the REST API instance.
type restApi struct {
	engine   *gin.Engine
	port     int
	routines []*routine
}

// endpoint represents an endpoint on the REST API.
type endpoint struct {
	method  string
	url     string
	handler func(*restApi, *gin.Context)
	desc    string
}

// endpointMap is a mapping of endpoints for a specific group.
type endpointMap []endpoint

var generalEndpoints = endpointMap{
	// GET routes
	{ http.MethodGet,    "/ping",                      (*restApi).handleGetPing,
		"DESC" },
	{ http.MethodGet,    "/endpoints",                 (*restApi).handleGetEndpoints,
		"DESC" },
}

// Map of endpoints for actions involving routines
var routineEndpoints = endpointMap{
	// GET routes
	{ http.MethodGet,    "/routines",                  (*restApi).handleGetRoutineAll,
		"DESC" },
	{ http.MethodGet,    "/routines/:routine",         (*restApi).handleGetRoutine,
		"DESC" },

	// PUT routes
	{ http.MethodPut,    "/routines/refresh",          (*restApi).handlePutRefreshAll,
		"DESC" },
	{ http.MethodPut,    "/routines/refresh/:routine", (*restApi).handlePutRefresh,
		"DESC" },

	// PATCH routes
	{ http.MethodPatch,  "/routines/:routine",         (*restApi).handlePatchRoutine,
		"DESC" },

	// DELETE routes
	{ http.MethodDelete, "/routines",                  (*restApi).handleDeleteRoutineAll,
		"DESC" },
	{ http.MethodDelete, "/routines/:routine",         (*restApi).handleDeleteRoutine,
		"DESC" },
}

// New builds out a new REST API instance that is ready to be run. By default, the REST API listens on port 3991.
// You can change this value with setPort. The default gin engine has recovery logic and a logger baked in.
func newRestApi() *restApi {
	rest := new(restApi)

	// Use a default port of 3991.
	rest.setPort(3991)

	// Set up a new gin engine.
	rest.engine = gin.Default()

	// Build the mappings for v1.
	rest.buildV1()

	return rest
}

// run runs the REST API engine.
func (r *restApi) run() {
	if r != nil && r.engine != nil {
		port := fmt.Sprintf(":%d", r.port)
		r.engine.Run(port)
	}
}

// stop stops the REST API engine.
func (r *restApi) stop() {
	// https://github.com/gin-gonic/gin#graceful-shutdown-or-restart
}

// setPort sets the port. If not specified before calling Run, the port defaults to 3991.
func (r *restApi) setPort(port int) {
	if r != nil {
		r.port = port
	}
}

// setRoutines sets the routines that the REST API is layered on top of.
func (r *restApi) setRoutines(routines ...*routine) {
	if r != nil {
		r.routines = routines
	}
}

// buildV1 builds out the mappings for REST API v1 with this prefix: /rest/v1
func (r *restApi) buildV1() {
	if r != nil && r.engine != nil {
		v1 := r.engine.Group("/rest/v1")
		maps := []endpointMap{}

		// Build the mapping for the REST API endpoints.
		maps = append(maps, generalEndpoints)

		// Build the mapping for the routine endpoints.
		maps = append(maps, routineEndpoints)

		for _, m := range maps {
			for _, route := range m {
				// We have to copy the function pointer to make sure the closures below use the correct *restApi method.
				handler := route.handler
				switch route.method {
				case http.MethodGet:
					v1.GET(route.url, func(c *gin.Context) {
						handler(r, c)
					})
				case http.MethodHead:
					v1.HEAD(route.url, func(c *gin.Context) {
						handler(r, c)
					})
				case http.MethodPost:
					v1.POST(route.url, func(c *gin.Context) {
						handler(r, c)
					})
				case http.MethodPut:
					v1.PUT(route.url, func(c *gin.Context) {
						handler(r, c)
					})
				case http.MethodPatch:
					v1.PATCH(route.url, func(c *gin.Context) {
						handler(r, c)
					})
				case http.MethodDelete:
					v1.DELETE(route.url, func(c *gin.Context) {
						handler(r, c)
					})
				case http.MethodOptions:
					v1.OPTIONS(route.url, func(c *gin.Context) {
						handler(r, c)
					})
				}
			}
		}
	}
}


// GET /ping
// handleGetPing responds to a ping request with "pong".
func (r *restApi) handleGetPing(c *gin.Context) {
	c.String(200, "pong")
}

// GET /endpoints
// handleGetEndpoints returns a JSON object of all possible v1 endpoints and their descriptions.
func (r *restApi) handleGetEndpoints(c *gin.Context) {
	type endpoints struct {
		Method string `json:"method"`
		URL    string `json:"url"`
		Desc   string `json:"description"`
	}
	maps := []endpointMap{generalEndpoints, routineEndpoints}
	c.JSON(200, maps)
}


// GET /routines
// handleGetRoutineAll responds with information about the statusbar and all the routines (active and inactive).
func (r *restApi) handleGetRoutineAll(c *gin.Context) {
}

// GET /routines/:routine
// handleGetRoutine responds with information about all the specified routine.
func (r *restApi) handleGetRoutine(c *gin.Context) {
	_, err := getRoutine(r.routines, c.Param("routine"))
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
}

// PUT /routines/refresh
// handlePutRefreshAll refreshes all active routines.
func (r *restApi) handlePutRefreshAll(c *gin.Context) {
}

// PUT /routines/refresh/:routine
// handlePutRefresh refreshes the specified routine.
func (r *restApi) handlePutRefresh(c *gin.Context) {
	_, err := getRoutine(r.routines, c.Param("routine"))
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
}

// PATCH /routines/:routine
// handlePatchRoutine updates the specified routine's data. Currently, this only updates the interval time.
func (r *restApi) handlePatchRoutine(c *gin.Context) {
	_, err := getRoutine(r.routines, c.Param("routine"))
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
}

// DELETE /routines
// handleDeleteRoutineAll stops the stasusbar.
func (r *restApi) handleDeleteRoutineAll(c *gin.Context) {
}

// DELETE /routines/:routine
// deleteRoutine stops the specified routine.
func (r *restApi) handleDeleteRoutine(c *gin.Context) {
	_, err := getRoutine(r.routines, c.Param("routine"))
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
}

// getRoutine gets the specified routine from the list of routines.
func getRoutine(routines []*routine, name string) (*routine, error) {
	for _, routine := range routines {
		if name == routine.moduleName() {
			return routine, nil
		}
	}

	return nil, errors.New("Invalid routine")
}

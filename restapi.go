package statusbar

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

// RestApi holds the information about the REST API instance.
type RestApi struct {
	engine   *gin.Engine
	port     int
	routines []routine
}

// NewRestApi builds out a new REST API instance that is ready to be run. By default, the REST API listens on port 3991.
// You can change this value with SetPort. The default gin engine has a logger and recovery logic baked in.
func NewRestApi() *RestApi {
	rest := new(RestApi)

	// Use a default port of 3991.
	rest.SetPort(3991)

	// Set up a new gin engine.
	rest.engine = gin.Default()

	// Build the mappings for v1.
	rest.buildV1()

	return rest
}

func (r *RestApi) Run() {
	if r != nil && r.engine != nil {
		port := fmt.Sprintf(":%d", r.port)
		r.engine.Run(port)
	}
}

// SetPort sets the port. If not specified before calling Run, the port defaults to 3991.
func (r *RestApi) SetPort(port int) {
	if r != nil {
		r.port = port
	}
}

// SetRoutines sets the routines that the REST API is layered on top of.
func (r *RestApi) SetRoutines(routines ...routine) {
	if r != nil {
		r.routines = routines
	}
}

// buildV1 builds out the mappings for REST API v1 with this prefix: /rest/v1
func (r *RestApi) buildV1() {
	if r != nil && r.engine != nil {
		v1 := r.engine.Group("/rest/v1")

		// GET routes
		v1.GET("/routines", func(c *gin.Context) { r.getRoutineAll(c) })
		v1.GET("/routines/:routine", func(c *gin.Context) { r.getRoutine(c) })

		// PUT routes
		v1.PUT("/routines/refresh", func(c *gin.Context) { r.putRefreshAll(c) })
		v1.PUT("/routines/refresh/:routine", func(c *gin.Context) { r.putRefresh(c) })

		// PATCH routes
		v1.PATCH("/routines/:routine", func(c *gin.Context) { r.patchRoutine(c) })

		// DELETE routes
		v1.DELETE("/routines", func(c *gin.Context) { r.deleteRoutineAll(c) })
		v1.DELETE("/routines/:routine", func(c *gin.Context) { r.deleteRoutine(c) })
	}
}


// GET /routines
// getRoutineAll responds with information about the statusbar and all the routines (active and inactive).
func (r *RestApi) getRoutineAll(c *gin.Context) {
	log.Printf("GET /routines")
}

// GET /routines/:routine
// getRoutine responds with information about all the specified routine.
func (r *RestApi) getRoutine(c *gin.Context) {
	log.Printf("GET /routines/:routine")
	log.Printf("routine: %s", c.Param("routine"))
}

// PUT /routines/refresh
// putRefreshAll refreshes all active routines.
func (r *RestApi) putRefreshAll(c *gin.Context) {
	log.Printf("PUT /routines/refresh")
}

// PUT /routines/refresh/:routine
// putRefresh refreshes the specified routine.
func (r *RestApi) putRefresh(c *gin.Context) {
	log.Printf("PUT /routines/refresh/:routine")
}

// PATCH /routines/:routine
// patchRoutine updates the specified routine's data. Currently, this only updates the interval time.
func (r *RestApi) patchRoutine(c *gin.Context) {
	log.Printf("PATCH /routines/:routine")
}

// DELETE /routines
// deleteRoutineAll stops the stasusbar.
func (r *RestApi) deleteRoutineAll(c *gin.Context) {
	log.Printf("DELETE /routines")
}

// DELETE /routines/:routine
// deleteRoutine stops the specified routine.
func (r *RestApi) deleteRoutine(c *gin.Context) {
	log.Printf("DELETE /routines/:routine")
}

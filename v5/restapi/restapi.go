// +build go1.8

// Package restapi implements a REST API engine using the Gin routing framework.
//
// The basic premise of this package is that you set up the routing table with the provided specification and then build
// the endpoints directly with a single call. With everything ready to go, you then run the engine.
//
// The most important step is building the REST API spec. You can do this by structuring up a JSON object and importing
// that directly or by using an unmarshaled structure.
//
// To begin, review the notes on RestSpec and craft your implementation, making sure that every Endpoint's Callback
// implements HandlerFunc. Next, create a new Engine and import the specification using the appropriate AddSpec* helper
// (AddSpec to add a RestSpec structure directly, AddSpecFile to import a file containing the JSON representation of the
// spec, or AddSpecReader to use an io.Reader that wraps the JSON object for the spec). Finally, run the engine with
// Run. The Gin engine now handles the REST API routing, mapping URLs to Callbacks.
package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"reflect"
	"time"
)

// Engine is the main type for this package. It handles building and running all versions of the REST API according to
// the specifications provided.
type Engine struct {
	engine *gin.Engine
	server *http.Server
}

// Params is a map of REST path parameters to their values. For example, if a path is specified as "/weather/:day" in
// the specs and a client hits the endpoint "/weather/sunday", then params["day"] = "sunday".
type Params map[string]string

// HandlerFunc is the function definition that handlers must use to define a route's callback. For example, let's say
// the RestSpec specification has an Endpoint with a URL of "/users/list" and a Callback of "ListAllUsers". The handler
// object passed in with AddSpec* would then need to implement ListAllUsers(Endpoint, Params, *http.Request) (int, string).
// AddSpec* returns an error if the handler object does not implement a method with that exact name and definition. The
// method must be exported (capitalized) for it to be visible to the engine.
//
// Endpoint is the endpoint as defined in the specification. Params is a map of Gin URL params. *http.Request is the
// data as sent in the request. The method returns the HTTP response code and a string of response data.
type HandlerFunc func(Endpoint, Params, *http.Request) (int, string)

// RestSpec is the data model for the REST API specification. To implement the REST API, you build out a JSON object
// following this model and import it using an AddSpec* helper.
type RestSpec struct {
	// Name of the API.
	Name string `json:"name"`

	// Prefix to use for endpoint routing, e.g. /api/v1.
	Prefix string `json:"prefix"`

	// Description of this specification.
	Desc string `json:"description"`

	// Version of this specification.
	Version float64 `json:"version"`

	// List of endpoint tables.
	Tables []Table `json:"tables"`
}

// Table is used to conceptually group together similar endpoints.
type Table struct {
	// Name of this table.
	Name string `json:"name"`

	// Description of this table.
	Desc string `json:"description"`

	// List of endpoints in this table.
	Endpoints []Endpoint `json:"endpoints"`
}

// Endpoint holds the information needed to build an endpoint, including its URL, description, and request/response
// parameters.
type Endpoint struct {
	// HTTP method for this endpoint, e.g. "GET" or "POST".
	Method string `json:"method"`

	// URL for this method, e.g. /users/list. For formatting options, see Gin's example documentation:
	// https://github.com/gin-gonic/gin#api-examples.
	URL string `json:"url"`

	// Description of this endpoint.
	Desc string `json:"description"`

	// Map of key/value pairs for request data.
	Request map[string]interface{} `json:"request":`

	// Map of key/value pairs in response data.
	Response map[string]interface{} `json:"response":`

	// Handler callback that is called to handle this endpoint's implementation. See the HandlerFunc type for more
	// information on this.
	Callback string `json:"callback"`
}

// NewEngine creates a new Engine using Gin's default engine, which includes fault handling and logging.
func NewEngine() *Engine {
	e := new(Engine)
	e.engine = gin.Default()

	return e
}

// AddSpec adds the enpoints in the specification to Engine's routes.
func (e *Engine) AddSpec(spec RestSpec, handler interface{}) error {
	if e == nil || e.engine == nil {
		return fmt.Errorf("invalid Engine")
	}

	engine := e.engine

	// Set up this API's group.
	group := engine.Group(spec.Prefix)

	// Get the underlying type of the handler. We use this to find the appropriate methods later.
	handlerType := reflect.ValueOf(handler)

	// Map the endpoints into the engine.
	for _, table := range spec.Tables {
		for _, endpoint := range table.Endpoints {
			if endpoint.Callback == "" {
				return fmt.Errorf("missing callback for %s", endpoint.URL)
			}

			// Figure out which of the handler's methods we need to set as this endpoint's handler.
			handler := handlerType.MethodByName(endpoint.Callback)
			if handler == (reflect.Value{}) {
				return fmt.Errorf("handler does not implement %s", endpoint.Callback)
			}

			// Type assert the callback method back to the correct definition.
			iface := handler.Interface()
			f, ok := iface.(func(Endpoint, Params, *http.Request) (int, string))
			if !ok {
				return fmt.Errorf("%s does not satisfy HandlerFunc", endpoint.Callback)
			}

			group.Handle(endpoint.Method, endpoint.URL, func(c *gin.Context) {
				params := make(Params)
				for _, p := range c.Params {
					params[p.Key] = p.Value
				}

				code, output := f(endpoint, params, c.Request)
				if code >= 400 && code < 600 {
					c.Error(fmt.Errorf(output))
				}
				if output == "" {
					c.Status(code)
				} else {
					if json.Valid([]byte(output)) {
						c.Header("Content-Type", "application/json")
					}
					c.String(int(code), output)
				}
			})
		}
	}

	return nil
}

// AddSpecFile reads the REST API specification in the file at path and adds it to Engine's routes. The
// specification must be JSON-encoded using the template defined in RestSpec.
func (e *Engine) AddSpecFile(path string, handler interface{}) error {
	if e == nil || e.engine == nil {
		return fmt.Errorf("invalid Engine")
	}

	// Open file at path.
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	// Unmarshal JSON in file and add the endpoints.
	return e.AddSpecReader(file, handler)
}

// AddSpecReader reads the REST API specification in r and adds it to Engine's routes. The specification must be
// JSON-encoded using the template defined in RestSpec.
func (e *Engine) AddSpecReader(r io.Reader, handler interface{}) error {
	// Unmarshal JSON in reader.
	spec := RestSpec{}
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&spec); err != nil && err != io.EOF {
		return err
	}

	return e.AddSpec(spec, handler)
}

// Run runs the API engine in a new goroutine and listens on the designated port.
func (e *Engine) Run(port int) {
	if e != nil && e.engine != nil && e.server == nil {
		// We need to create a new server on every call to Run because a server cannot be reused after a call to Stop.
		e.server = new(http.Server)
		e.server.Handler = e.engine
		e.server.Addr = fmt.Sprintf(":%d", port)

		go e.server.ListenAndServe()
	}
}

// Stop stops the API engine in timeout seconds.
func (e *Engine) Stop(timeout int) error {
	if e == nil || e.server == nil {
		return fmt.Errorf("invalid server")
	}

	// Give the engine 5 seconds to close out any connections.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	err := e.server.Shutdown(ctx)
	e.server = nil
	return err
}

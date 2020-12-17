// Package restapi implements a REST API engine using the Gin routing framework.
package restapi

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"reflect"
)

// Engine is the main type for this package. It handles building and running all versions of the REST API according to
// the specifications provided.
type Engine struct {
	engine *gin.Engine
}

// Params is a map of REST path parameters to their values. For example, if a path is specified as "/weather/:day" in the
// specs and a client hits the endpoint "/weather/sunday", then Params["day"] = "sunday".
type Params map[string]string

// HandlerFunc is the function definition that handlers must use to define a route's callback. For example, let's say the JSON
// spec has an endpoint for "/users/list" with a registered callback of "ListAllUsers". Whatever handler object is
// passed in with AddSpec would then need to implement ListAllUsers(Params, *http.Request) (int, string). AddSpec will
// return an error if the handler object does not implement a method with that exact name and definition.
type HandlerFunc func(Endpoint, Params, *http.Request) (int, string)

// RestSpec is the data model for the REST API specification. To implement the REST API, you build out a JSON object
// following this model and import it using AddSpec or AddSpecFile.
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

// Endpoint holds the information needed to build an endpoint, including its url, description, and request/response
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
func NewEngine() Engine {
	return Engine{gin.Default()}
}

// AddSpecFile reads the REST API specification in the file at path and adds it to Engine's routes. The
// specification must be JSON-encoded using the template defined in RestSpec.
func (e *Engine) AddSpecFile(path string, handler interface{}) error {
	if e == nil || e.engine == nil {
		return fmt.Errorf("Invalid Engine")
	}

	// Open file at path.
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	// Unmarshal JSON in file.
	spec := RestSpec{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&spec); err != nil && err != io.EOF {
		return err
	}

	return e.AddSpec(spec, handler)
}

// AddSpec adds the enpoints in the specification to Engine's routes.
func (e *Engine) AddSpec(spec RestSpec, handler interface{}) error {
	if e == nil || e.engine == nil {
		return fmt.Errorf("Invalid Engine")
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
				return fmt.Errorf("Missing callback for %s", endpoint.URL)
			}

			// Figure out which of the handler's methods we need to set as this endpoint's handler.
			handler := handlerType.MethodByName(endpoint.Callback)
			if handler == (reflect.Value{}) {
				return fmt.Errorf("Handler does not implement %s", endpoint.Callback)
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
				c.String(int(code), output)
			})
		}
	}

	return nil
}

// Run runs the API engine and listens on port.
func (e *Engine) Run(port int) {
	if e != nil && e.engine != nil {
		p := fmt.Sprintf(":%d", port)
		e.engine.Run(p)
	}
}

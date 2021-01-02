// Package sbtravisci displays the current build status of the given repository.
package sbtravisci

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var colorEnd = "^d^"

// Routine is the main object for this package. It contains the information needed to query the build status.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Client used to make the requests.
	client *http.Client

	// Request to get the most recent build status.
	request *http.Request

	// Latest build.
	build build

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// build holds the information that Travis returns for the latest build.
type build struct {
	Repo struct {
		Name string `json:"name"`
	} `json:"repository"`
	Number  string `json:"number"`
	State   string `json:"state"`
	BeginTS string `json:"started_at"`
	EndTS   string `json:"finished_at"`
}

// New makes a new routine object. owner is the username of the repository's owner. repo is the name of the repository.
// colors is an optional triplet of hex color codes for colorizing the output based on these rules:
//   1. Normal color, used for enqueued/passing builds.
//   2. Warning color, used for canceled/failed builds.
//   3. Error color, used for error messages.
func New(owner, repo string, colors ...[3]string) *Routine {
	r := new(Routine)

	// Set up our client with a timeout of 30 seconds (the default client does not have a timeout).
	r.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// The API requires that the owner and repo are joined with "%2F" instead of a literal slash. url.URL escapes the
	// percent sign in the path, so we have to unescape it. Additionally, we have to show both the literal path and the
	// escaped path when creating the URL object so that url.String() knows to use the literal version.
	rawPath := fmt.Sprintf("repo/%s%%2F%s/builds", owner, repo)
	escPath, _ := url.PathUnescape(rawPath)

	// We only want to request the most recent build.
	query := url.Values{}
	query.Set("limit", "1")

	// Set up the URL. We don't need to validate any parameters, because Travis will do the error checking for us.
	u := url.URL{
		Scheme:   "https",
		Host:     "api.travis-ci.com",
		RawPath:  rawPath,
		Path:     escPath,
		RawQuery: query.Encode(),
	}

	// Set up the request.
	r.request, _ = http.NewRequest("GET", u.String(), nil)
	r.request.Header.Add("Travis-API-Version", "3")

	// Store the color codes. Don't do any validation.
	if len(colors) > 0 {
		r.colors.normal = "^c" + colors[0][0] + "^"
		r.colors.warning = "^c" + colors[0][1] + "^"
		r.colors.error = "^c" + colors[0][2] + "^"
	} else {
		// If a color array wasn't passed in, then we don't want to print this.
		colorEnd = ""
	}

	return r
}

// Update gets the current clone count.
func (r *Routine) Update() (bool, error) {
	if r == nil {
		return false, fmt.Errorf("bad routine")
	}

	build, err := r.getBuild()
	if err != nil {
		r.err = err
		return true, err
	}

	r.build = build

	return true, nil
}

// String prints the latest build status.
func (r *Routine) String() string {
	if r == nil {
		return "bad routine"
	}

	// Figure out which color and timestamp we need to use for this state.
	var color string
	var ts string
	switch r.build.State {
	case "created":
		// This state doesn't have a timestamp yet.
		color = r.colors.normal
	case "started":
		color = r.colors.normal
		ts = r.build.BeginTS
	case "passed":
		color = r.colors.normal
		ts = r.build.EndTS
	case "failed":
		color = r.colors.warning
		ts = r.build.EndTS
	case "canceled":
		color = r.colors.warning
		ts = r.build.EndTS
	default:
		color = r.colors.error
		if r.build.EndTS != "" {
			ts = r.build.EndTS
		} else {
			ts = r.build.BeginTS
		}
	}

	// Try to parse out the timestamp for this state.
	if t, err := time.Parse("2006-01-02T15:04:05Z", ts); err == nil {
		// The timestamp gets returned in UTC time. Convert it over to the local timezone.
		t = t.Local()

		// This gets added to the end of the build status message.
		ts = " on " + t.Format("01/02 at 15:04")
	} else {
		ts = ""
	}

	return fmt.Sprintf("%s%s: %s%s%s", color, r.build.Repo.Name, strings.Title(r.build.State), ts, colorEnd)
}

// Error formats and returns an error message.
func (r *Routine) Error() string {
	if r == nil {
		return "bad routine"
	}

	if r.err == nil {
		r.err = fmt.Errorf("unknown error")
	}

	return r.colors.error + r.err.Error() + colorEnd
}

// Name returns the display name of this module.
func (r *Routine) Name() string {
	return "Travis CI Build Status"
}

// getBuild gets the latest build.
func (r *Routine) getBuild() (build, error) {
	type Response struct {
		Error  string  `json:"error_message"`
		Builds []build `json:"builds"`
	}

	resp, err := r.client.Do(r.request)
	if err != nil {
		return build{}, err
	}
	defer resp.Body.Close()

	// Pull out the response data.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return build{}, err
	}

	// Parse the JSON doc.
	response := Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		return build{}, err
	}

	if response.Error != "" {
		return build{}, fmt.Errorf(response.Error)
	}

	if len(response.Builds) < 1 {
		return build{}, fmt.Errorf("missing latest build")
	}

	return response.Builds[0], nil
}

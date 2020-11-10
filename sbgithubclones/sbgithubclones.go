package sbgithubclones

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var colorEnd = "^d^"

// Routine is the main object for this package. It contains the objects needed to query the current clone count.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Request with user-supplied information.
	req *http.Request

	// Client for sending GET requests.
	client *http.Client

	// Total number of clones today.
	total int

	// Number of unique clones today.
	unique int

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New makes a new routine object. owner is the username of the repository's owner. repo is the name of the repository.
// authUser is the username for authentication (must have push permissions to repo). authToken is the token for
// authentication. colors is an optional triplet of hex color codes for colorizing the output based on these rules:
//   1. Normal color, used for normal printing.
//   2. Warning color, currently unused.
//   3. Error color, used for printing error messages.
func New(owner, repo, authUser, authToken string, colors ...[3]string) *Routine {
	var r Routine

	// Set up the query.
	q := url.Values{}
	q.Set("per", "day")

	// Set up the URL. We don't need to validate any parameters, because Github will do the error checking for us.
	url := url.URL{
		Scheme:   "https",
		Host:     "api.github.com",
		Path:     fmt.Sprintf("repos/%s/%s/traffic/clones", url.PathEscape(owner), url.PathEscape(repo)),
		RawQuery: q.Encode(),
	}

	// Set up the request.
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		r.err = err
		return &r
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.SetBasicAuth(authUser, authToken)
	r.req = req

	// Initialize our client.
	r.client = &http.Client{}

	// Store the color codes. Don't do any validation.
	if len(colors) > 0 {
		r.colors.normal = "^c" + colors[0][0] + "^"
		r.colors.warning = "^c" + colors[0][1] + "^"
		r.colors.error = "^c" + colors[0][2] + "^"
	} else {
		// If a color array wasn't passed in, then we don't want to print this.
		colorEnd = ""
	}

	return &r
}

// Update gets the current clone count.
func (r *Routine) Update() (bool, error) {
	if r == nil {
		return false, errors.New("Bad routine")
	}

	type CloneCount struct {
		Timestamp string `json:"timestamp"`
		Count     int    `json:"count"`
		Uniques   int    `json:"uniques"`
	}

	type CloneCounts struct {
		Counts []CloneCount `json:"clones"`
	}

	// Send the request.
	resp, err := r.client.Do(r.req)
	if err != nil {
		r.err = err
		return true, err
	}
	defer resp.Body.Close()

	// Pull out the response data.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		r.err = err
		return true, err
	}

	// Parse the json.
	c := CloneCounts{}
	if err := json.Unmarshal(body, &c); err != nil {
		r.err = err
		return true, err
	}

	// Grab the count for today. If there is no timestamp matching today, then it means there haven't been any clones
	// so far.
	now := time.Now().UTC()
	day := now.Day()
	for _, count := range c.Counts {
		if t, err := time.Parse("2006-01-02T00:00:00Z", count.Timestamp); err == nil {
			if t.Day() == day {
				r.total = count.Count
				r.unique = count.Uniques
				break
			}
		}
	}

	return true, nil
}

// String prints the current clone count.
func (r *Routine) String() string {
	if r == nil {
		return "Bad routine"
	}

	return fmt.Sprintf("%s%v Clones (%v Unique)%s", r.colors.normal, r.total, r.unique, colorEnd)
}

// Error formats and returns an error message.
func (r *Routine) Error() string {
	if r == nil {
		return "Bad routine"
	}

	if r.err == nil {
		r.err = errors.New("Unknown error")
	}

	return r.colors.error + r.err.Error() + colorEnd
}

// Name returns the display name of this module.
func (r *Routine) Name() string {
	return "Github Clone Count"
}

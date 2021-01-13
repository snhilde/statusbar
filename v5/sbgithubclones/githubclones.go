// Package sbgithubclones displays the daily and weekly clone count for a Github repository.
package sbgithubclones

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var colorEnd = "^d^"

// Routine is the main object for this package. It contains the objects needed to query the current clone count for the
// day and week.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Name of repository.
	repo string

	// Client used to make the requests.
	client *http.Client

	// Requests to get the daily and weekly counts.
	reqDay  *http.Request
	reqWeek *http.Request

	// Total number of clones today and this week.
	dayCount  string
	weekCount string

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

	r.repo = repo

	// Set up our client with a timeout of 30 seconds (the default client does not have a timeout).
	r.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	day, err := buildRequest(owner, repo, authUser, authToken, true)
	if err != nil {
		r.err = err
		return &r
	}
	week, err := buildRequest(owner, repo, authUser, authToken, false)
	if err != nil {
		r.err = err
		return &r
	}
	r.reqDay = day
	r.reqWeek = week

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
		return false, fmt.Errorf("bad routine")
	}

	day, err := getCount(r.client, r.reqDay, true)
	if err != nil {
		r.err = fmt.Errorf("error getting today's count")
		return true, err
	}
	r.dayCount = day

	week, err := getCount(r.client, r.reqWeek, false)
	if err != nil {
		r.err = fmt.Errorf("error getting this week's count")
		return true, err
	}
	r.weekCount = week

	return true, nil
}

// String prints the current clone count for the day and week. If no count is returned, this displays "-" for that time
// period.
func (r *Routine) String() string {
	if r == nil {
		return "bad routine"
	}

	if r.dayCount == "" {
		r.dayCount = "-"
	}
	if r.weekCount == "" {
		r.weekCount = "-"
	}

	return fmt.Sprintf("%s%s: %s/%s Clones%s", r.colors.normal, r.repo, r.dayCount, r.weekCount, colorEnd)
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
	return "Github Clone Count"
}

// buildRequest builds the request that will be used to get either the daily or weekly clone counts.
func buildRequest(owner, repo, authUser, authToken string, daily bool) (*http.Request, error) {
	// Set up the query.
	q := url.Values{}
	if daily {
		q.Set("per", "day")
	} else {
		q.Set("per", "week")
	}

	// Set up the URL. We don't need to validate any parameters, because Github will do the error checking for us.
	u := url.URL{
		Scheme:   "https",
		Host:     "api.github.com",
		Path:     fmt.Sprintf("repos/%s/%s/traffic/clones", url.PathEscape(owner), url.PathEscape(repo)),
		RawQuery: q.Encode(),
	}

	// Set up the request.
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.SetBasicAuth(authUser, authToken)

	return req, nil
}

// getCount queries Github for the current clone count for either the day or week.
func getCount(client *http.Client, req *http.Request, daily bool) (string, error) {
	type CloneCount struct {
		Timestamp string `json:"timestamp"`
		Count     int    `json:"count"`
	}

	type CloneCounts struct {
		Message string       `json:"message"`
		Counts  []CloneCount `json:"clones"`
	}

	// Get the count.
	resp, err := client.Do(req)
	if err != nil {
		return "-", err
	}
	defer resp.Body.Close()

	// Pull out the response data.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "-", err
	}

	// Parse the JSON doc.
	c := CloneCounts{}
	if err := json.Unmarshal(body, &c); err != nil {
		return "-", err
	}

	// If there's an error message in the response, then something went wrong.
	if c.Message != "" {
		if c.Message == "not found" {
			// Let's make this error message a little more obvious.
			c.Message = "repository not found"
		}
		return "-", fmt.Errorf(c.Message)
	}

	// Find the current count for this reporting period.
	day := getDay(daily)
	for _, count := range c.Counts {
		if t, err := time.Parse("2006-01-02T00:00:00Z", count.Timestamp); err == nil {
			if t.Day() == day {
				return fmt.Sprintf("%v", count.Count), nil
			}
		}
	}

	// If we didn't find a matching timestamp, then that means there haven't been any clones for this time period.
	return "-", nil
}

// getDay determines which day we need to use when looking for the current clone count. For the daily count, we use the
// current day. For the weekly count, we go back to the nearest Monday and use that.
func getDay(daily bool) int {
	now := time.Now()
	if !daily {
		// Set the current day to the most recent Monday. Note: The first weekday is Sunday, which is indexed at 0.
		dayOfWeek := int(now.Weekday())
		if dayOfWeek == 0 {
			// For Sunday, go back six days.
			now = now.AddDate(0, 0, -6)
		} else {
			// For all other days, this goes back the correct number of days to get to Monday.
			now = now.AddDate(0, 0, 1-dayOfWeek)
		}
	}

	// We are now on the day that Github will use to report the current count for this reporting period. Let's grab the
	// string of this time to match later.
	return now.Day()
}

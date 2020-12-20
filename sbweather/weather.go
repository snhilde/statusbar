// Package sbweather displays the current weather in the provided zip code.
// Currently supported for US only.
package sbweather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var colorEnd = "^d^"

// Routine is the main object for this package.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Whether or not the routine has been initialized yet.
	initialized bool

	// zip code for localizing the weather, if provided
	zip string

	// NWS-provided URL for getting the temperature, as found during the init.
	url string

	// HTTP client to reuse for all requests out.
	client *http.Client

	// Current temperature for the provided zip code.
	currTemp string

	// Forecast high.
	highTemp string

	// Forecast low.
	lowTemp string

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New makes a new routine object. zip is the zip code to use for localizing the weather. If no zip code is provided
// (i.e., zip is ""), then the module shows the weather at the IP address's location. colors is an optional triplet of hex
// color codes for colorizing the output based on these rules:
//   1. Normal color, used for printing the current temperature and forecast.
//   2. Warning color, currently unused.
//   3. Error color, used for error messages.
func New(zip string, colors ...[3]string) *Routine {
	var r Routine

	// Set up our client with a timeout of 30 seconds (the default client does not have a timeout).
	r.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	r.zip = zip

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

// Update gets the current hourly temperature.
func (r *Routine) Update() (bool, error) {
	if r == nil {
		return false, fmt.Errorf("bad routine")
	}

	// See if we need to initialize the routine still. We're doing this here instead of in New so as to not block the
	// start-up process of other routines.
	if !r.initialized {
		if err := r.init(); err != nil {
			r.err = fmt.Errorf("failed to start up")
			// We're going to return true so we can try the process again when the connection is back online (assuming
			// that's the problem).
			return true, err
		}
		r.initialized = true
	}

	// Reset all readings.
	r.currTemp = ""
	r.highTemp = ""
	r.lowTemp = ""

	// Get hourly temperature.
	temp, err := getTemp(r.client, r.url+"/hourly")
	if err != nil {
		r.err = err
		return true, err
	}
	r.currTemp = temp

	high, low, err := getForecast(r.client, r.url)
	if err != nil {
		r.err = err
		return true, err
	}
	r.highTemp = high
	r.lowTemp = low

	// Sanity check the high and low temps against the current temp.
	r.checkTemps()

	return true, nil
}

// String formats and prints the current temperature.
func (r *Routine) String() string {
	if r == nil {
		return "bad routine"
	}

	// Grab some info on which day's forecast we're reporting.
	day := "today"
	if time.Now().Hour() >= 15 {
		day = "tom"
	}

	// Let's work through the different scenarios where we might or might not have high and low temperature forecasts.
	haveCurr := r.currTemp != ""
	haveHigh := r.highTemp != ""
	haveLow := r.lowTemp != ""

	if haveCurr {
		if haveHigh {
			if haveLow {
				// We know the current temperature as well as the high and low forecasts.
				return fmt.Sprintf("%s%v °F (%s: %v/%v)%s", r.colors.normal, r.currTemp, day, r.highTemp, r.lowTemp, colorEnd)
			}
			// We know the current temperature and the high forecast.
			return fmt.Sprintf("%s%v °F (%s's high: %v)%s", r.colors.normal, r.currTemp, day, r.highTemp, colorEnd)
		} else if haveLow {
			// We know the current temperature and the low forecast.
			return fmt.Sprintf("%s%v °F (%s's low: %v)%s", r.colors.normal, r.currTemp, day, r.lowTemp, colorEnd)
		}
		// We know only the current temperature.
		return fmt.Sprintf("%s%v °F now%s", r.colors.normal, r.currTemp, colorEnd)
	} else if haveHigh {
		if haveLow {
			// We have the high and low forecasts but no current reading.
			return fmt.Sprintf("%s%s: %v/%v °F%s", r.colors.normal, day, r.highTemp, r.lowTemp, colorEnd)
		}
		// We have only the forecasted high.
		return fmt.Sprintf("%s%s's high: %v °F%s", r.colors.normal, day, r.highTemp, colorEnd)
	} else if haveLow {
		// We have only the forecasted low.
		return fmt.Sprintf("%s%s's low: %v °F%s", r.colors.normal, day, r.lowTemp, colorEnd)
	}

	// If we're here, then we don't have any weather information at all.
	return r.colors.warning + "No weather data" + colorEnd
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
	return "Weather"
}

// init initializes the weather data. If a zip code was specified, then we'll use the geographic coordinates for that
// area. Otherwise, we'll use the current coordinates of the IP address.
func (r *Routine) init() error {
	lat, long, err := getCoords(r.client, r.zip)
	if err != nil {
		return err
	}

	// Reduce to 4 decimal places of precision.
	lat, long = reduceCoords(lat, long)

	// Get the URL for the forecast at the geographic coordinates.
	url, err := getURL(r.client, lat, long)
	if err != nil {
		return err
	}
	r.url = url

	return nil
}

// checkTemps checks that the current temperature is not outside the bounds of the high and low for the day.
func (r *Routine) checkTemps() {
	// We only want to check this for the current day's forecast.
	t := time.Now()
	if t.Hour() >= 15 {
		return
	}

	if r.currTemp == "" {
		return
	}

	curr, err := strconv.Atoi(r.currTemp)
	if err != nil {
		return
	}

	if r.highTemp != "" {
		if high, err := strconv.Atoi(r.highTemp); err == nil {
			if curr > high {
				r.highTemp = r.currTemp
				return
			}
		}
	}

	if r.lowTemp != "" {
		if low, err := strconv.Atoi(r.lowTemp); err == nil {
			if curr < low {
				r.lowTemp = r.currTemp
			}
		}
	}
}

// getCoords is a jumping point for getting the geographic coordinates based on either the provided zip or an IP address.
func getCoords(client *http.Client, zip string) (string, string, error) {
	if zip == "" {
		// Get the coordinates of the IP address.
		return ipToCoords(client)
	}

	// Convert the provided zip code into geographic coordinates.
	return zipToCoords(client, zip)
}

// ipToCoords gets the geographic coordinates centered around the IP address. The request returns ASCII data that is not
// wrapped in any protocol layer. The coordinates will look like this: lat.1234,long.1234
func ipToCoords(client *http.Client) (string, string, error) {
	type coords struct {
		Lat float32 `json:"lat"`
		Lon float32 `json:"lon"`
	}

	url := "http://ip-api.com/json?fields=lat,lon"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	c := coords{}
	if err := json.Unmarshal(body, &c); err != nil {
		return "", "", err
	}

	if c.Lat == 0 && c.Lon == 0 {
		return "", "", fmt.Errorf("failed to find coordinates")
	}

	return fmt.Sprintf("%v", c.Lat), fmt.Sprintf("%v", c.Lon), nil
}

// zipToCoords gets the geographic coordinates for the provided zip code. It should receive a response in this format:
// {"status":1,"output":[{"zip":"90210","latitude":"34.103131","longitude":"-118.416253"}]}
func zipToCoords(client *http.Client, zip string) (string, string, error) {
	type coords struct {
		Status int                 `json:"status"`
		Output []map[string]string `json:"output"`
	}

	url := "https://api.promaptools.com/service/us/zip-lat-lng/get/?zip=" + zip + "&key=17o8dysaCDrgv1c"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	c := coords{}
	if err := json.Unmarshal(body, &c); err != nil {
		return "", "", err
	}

	// Make sure the status is good.
	if c.Status != 1 {
		return "", "", fmt.Errorf("coordinates request failed")
	}

	// Make sure we got back just one dictionary.
	if len(c.Output) != 1 {
		return "", "", fmt.Errorf("received invalid coordinates array")
	}

	lat := c.Output[0]["latitude"]
	long := c.Output[0]["longitude"]
	if lat == "" || long == "" {
		return "", "", fmt.Errorf("missing coordinates in response")
	}

	return lat, long, nil
}

// reduceCoords reduces the provided coordinates to 4 decimal places of precision.
func reduceCoords(lat, long string) (string, string) {
	if strings.Count(lat, ".") == 1 {
		i := strings.Index(lat, ".")
		l := i + 1 + 4 // +1 to include the decimal, +4 to have up to 4 decimal places of precision
		if len(lat) < l {
			l = len(lat)
		}

		lat = lat[:l]
	}

	if strings.Count(long, ".") == 1 {
		i := strings.Index(long, ".")
		l := i + 1 + 4 // +1 to include the decimal, +4 to have up to 4 decimal places of precision
		if len(long) < l {
			l = len(long)
		}

		long = long[:l]
	}

	return lat, long
}

// getURL queries the NWS to determine which URL we should be using for getting the weather forecast.
// Our value should be here: properties -> forecast.
func getURL(client *http.Client, lat string, long string) (string, error) {
	type props struct {
		Status     int    `json:"status"`
		Detail     string `json:"detail"`
		Properties struct {
			Forecast string `json:"forecast"`
		} `json:"properties"`
	}

	url := "https://api.weather.gov/points/" + lat + "," + long
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	p := props{}
	if err := json.Unmarshal(body, &p); err != nil {
		return "", err
	}

	// Catch some error codes.
	switch p.Status {
	// Add other codes here as they come up.
	case 301:
		return "", fmt.Errorf("max 4 digits of precision")
	case 404:
		return "", fmt.Errorf("invalid location")
	}

	url = p.Properties.Forecast
	if url == "" {
		return "", fmt.Errorf("bad temperature URL")
	}

	return url, nil
}

// getTemp gets the current temperature from the NWS database.
// Our value should be here: properties -> periods -> (latest period) -> temperature.
// If there's an error in the system, it will usually return a "status" element with a value of 500 and an error
// verbiage in a "title" element. We'll check for that error first and then look for the temperature.
func getTemp(client *http.Client, url string) (string, error) {
	type temp struct {
		Status     int    `json:"status"`
		Title      string `json:"title"`
		Properties struct {
			Periods []map[string]interface{} `json:"periods"`
		} `json:"properties"`
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("temp: bad request")
	}
	req.Header.Set("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("temp: bad client")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("temp: bad read")
	}

	t := temp{}
	if err := json.Unmarshal(body, &t); err != nil {
		return "", fmt.Errorf("temp: bad data")
	}

	if t.Status == 500 {
		if t.Title != "" {
			return "", fmt.Errorf(t.Title)
		}
		return "", fmt.Errorf("temp: server error")
	}

	// Get the list of weather readings.
	periods := t.Properties.Periods
	if len(periods) == 0 {
		return "", fmt.Errorf("missing hourly temperature periods")
	}

	// Use the most recent reading.
	latest := periods[0]
	if len(latest) == 0 {
		return "", fmt.Errorf("missing current temperature")
	}

	// Get just the temperature reading.
	return fmt.Sprintf("%v", latest["temperature"]), nil
}

// getForecast gets the forecasted temperatures from the NWS database.
// Our values should be here: properties -> periods -> (chosen periods) -> temperature.
// We're going to use these rules to determine which day's forecast we want:
//   1. If it's before 3 pm, we'll use the current day.
//   2. If it's after 3 pm, we'll display the high/low for the next day.
func getForecast(client *http.Client, url string) (string, string, error) {
	type forecast struct {
		Title      string `json:"title"`
		Properties struct {
			Periods []map[string]interface{} `json:"periods"`
		} `json:"properties"`
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("forecast: bad request")
	}
	req.Header.Set("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("forecast: bad client")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("forecast: bad read")
	}
	// TODO: handle expired grid.

	f := forecast{}
	if err := json.Unmarshal(body, &f); err != nil {
		return "", "", fmt.Errorf("forecast: bad JSON")
	}

	// Get the list of forecasts.
	periods := f.Properties.Periods
	if len(periods) == 0 {
		if f.Title != "" {
			return "", "", fmt.Errorf(f.Title)
		}
		return "", "", fmt.Errorf("missing forecast periods")
	}

	// If it's before 3pm, we'll use the forecast of the current day.
	// After that, we'll use tomorrow's forecast.
	t := time.Now()
	if t.Hour() >= 15 {
		t = t.Add(time.Hour * 12)
	}

	// For the day's high, we want to always look at the first time period that ends at 6:00 pm. If it's after 3:00 pm
	// for the day already, then we'll look at that time period for the following day.
	highEnd := t.Format("2006-01-02T") + "18:00:00"

	// For the day's low, we want to look at the time perioud that ends at 6:00 am on the following day. Like before,
	// this will be shifted back by a day if the current time is already past 3:00 pm.
	t = t.AddDate(0, 0, 1)
	lowEnd := t.Format("2006-01-02T") + "06:00:00"

	// Iterate through the list until we find the forecast for today/tomorrow.
	var high string
	var low string
	for _, f := range periods {
		// This is when this time period ends. The beginning of the time period will advance as the day advances, but
		// the end will always stay the same.
		endTime := f["endTime"].(string)
		if strings.Contains(endTime, highEnd) {
			// We'll get the high from here.
			high = fmt.Sprintf("%v", f["temperature"])
		} else if strings.Contains(endTime, lowEnd) {
			// We'll get the low from here.
			low = fmt.Sprintf("%v", f["temperature"])
		}

		if high != "" && low != "" {
			// This is all we need from the forecast, so we can exit now.
			break
		}
	}

	return high, low, nil
}

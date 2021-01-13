// Package sbweather displays the current weather and forecast for the provided coordinates.
package sbweather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var colorEnd = "^d^"

// noData is used to reset floats so we can tell whether or not they contain useful data.
const noData = -1234.5678

// Routine is the main object for this package.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Client used to make the requests.
	client *http.Client

	// Request to get the most recent build status.
	request *http.Request

	// Whether or not to use metric units.
	metric bool

	// Current temperature for the provided location.
	currTemp float32

	// Forecast high.
	highTemp float32

	// Forecast low.
	lowTemp float32

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// weather holds the weather data for today and the 7-day forecast.
type weather struct {
	Status  interface{} `json:"cod"`
	Message string      `json:"message"`
	Current struct {
		Temp float32 `json:"temp"`
	} `json:"current"`
	Forecasts []forecast `json:"daily"`
}

// forecast represents the forecast for one day.
type forecast struct {
	Timestamp int64 `json:"dt"`
	Range     struct {
		Low  float32 `json:"min"`
		High float32 `json:"max"`
	} `json:"temp"`
}

// New makes a new routine object with the specified latitude/longitude and formatting. key is the API key provided by
// OpenWeather. You can get a free key here: https://home.openweathermap.org/users/sign_up. The metric boolean denotes
// whether or not you want the temperature displayed in celsius. colors is an optional triplet of hex color codes for
// colorizing the output based on these rules:
//   1. Normal color, used for printing the current temperature and forecast.
//   2. Warning color, currently unused.
//   3. Error color, used for error messages.
func New(lat, lon float32, key string, metric bool, colors ...[3]string) *Routine {
	r := new(Routine)

	// Set up our client with a timeout of 30 seconds (the default client does not have a timeout).
	r.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Set the location and key query parameters. To get the forecast with a free API key, we have to use
	// latitude/longitude. There doesn't appear to be a maximum precision. There are 5 optional parts in the response
	// data: current weather, minutely forecast, hourly forecast, daily forecast, and weather alerts. We want to exclude
	// everything except the current weather and the daily forecasts.
	query := url.Values{}
	query.Set("lat", fmt.Sprintf("%f", lat))
	query.Set("lon", fmt.Sprintf("%f", lon))
	query.Set("appid", key)
	query.Set("exclude", "minutely,hourly,alerts")
	if metric {
		query.Set("units", "metric")
	} else {
		query.Set("units", "imperial")
	}
	r.metric = metric

	// Set up the URL. We don't need to validate any parameters, because OpenWeather does that for us.
	u := url.URL{
		Scheme:   "https",
		Host:     "api.openweathermap.org",
		Path:     "/data/2.5/onecall",
		RawQuery: query.Encode(),
	}

	// Set up the request.
	r.request, _ = http.NewRequest("GET", u.String(), nil)

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

// Update gets the current hourly temperature.
func (r *Routine) Update() (bool, error) {
	if r == nil {
		return false, fmt.Errorf("bad routine")
	}

	// Get weather data.
	weather, err := getWeather(r.client, r.request)
	if err != nil {
		r.err = fmt.Errorf("error getting weather data")
		return true, err
	}

	// Set the current temperature.
	r.currTemp = weather.Current.Temp

	// Reset the high and low temperatures so we know if we have an updated reading or not.
	r.highTemp = noData
	r.lowTemp = noData

	// Set the high and low temperatures. The forecast for each day starts at noon local time and is displayed in UTC
	// time. We'll start by getting a time.Time object for noon today, and then we'll skip ahead a day if we want the
	// high/low for tomorrow instead.
	now := time.Now()
	noon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	if !onToday() {
		noon = noon.AddDate(0, 0, 1)
	}
	ts := noon.Unix()
	for _, f := range weather.Forecasts {
		if f.Timestamp == ts {
			r.highTemp = f.Range.High
			r.lowTemp = f.Range.Low
		}
	}

	return true, nil
}

// String formats and prints the current temperature.
func (r *Routine) String() string {
	if r == nil {
		return "bad routine"
	}

	// Grab some info on which day's forecast we're reporting.
	day := ""
	if onToday() {
		day = "today"
	} else {
		day = "tom"
	}

	unit := ""
	if r.metric {
		unit = "°C"
	} else {
		unit = "°F"
	}

	// Let's work through the different scenarios where we might or might not have certain temperatures.
	haveHigh := r.highTemp != noData
	haveLow := r.lowTemp != noData

	// We will always have the current temperature. Start off with this.
	s := fmt.Sprintf("%.0f %s", r.currTemp, unit)
	if haveHigh {
		if haveLow {
			// We also know the high and low forecasts.
			s += fmt.Sprintf(" (%s: %.0f/%.0f)", day, r.highTemp, r.lowTemp)
		} else {
			// We also know the high forecast but not the low forecast.
			s += fmt.Sprintf(" (%s's high: %.0f)", day, r.highTemp)
		}
	} else if haveLow {
		// We also know the low forecast but not the high forecast.
		s += fmt.Sprintf(" (%s's low: %.0f)", day, r.lowTemp)
	} else {
		// We know only the current temperature.
		s += " now"
	}

	// Add the colors and return the entire string.
	return r.colors.normal + s + colorEnd
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

// getWeather gets the current weather data from OpenWeather.
func getWeather(client *http.Client, request *http.Request) (weather, error) {
	resp, err := client.Do(request)
	if err != nil {
		return weather{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return weather{}, err
	}

	w := weather{}
	if err := json.Unmarshal(body, &w); err != nil {
		return weather{}, err
	}

	// Check for an error in the response data. For bad latitude/longitude, the status code is returned as a string. For
	// bad API key, it's returned as a number.
	if status, ok := w.Status.(float64); ok {
		if status == 401 {
			return weather{}, fmt.Errorf("invalid API key")
		}
	}
	if w.Message != "" {
		return weather{}, fmt.Errorf(w.Message)
	}

	return w, nil
}

// onToday checks whether or not the forecast is for today or tomorrow. If the current time is before 3pm, then the
// forecast is still for today. Otherwise, it's for tomorrow.
func onToday() bool {
	return time.Now().Hour() < 15
}

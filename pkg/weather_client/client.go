package weather_client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Response struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Pressure float64 `json:"pressure"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	}
	Sys struct {
		Sunrise int64 `json:"sunrise"`
		Sunset  int64 `json:"sunset"`
	} `json:"sys"`
}

type Result struct {
	Temp          float64
	Pressure      float64
	Humidity      int
	WindSpeed     float64
	WindDirection float64
	Sunrise       time.Time
	Sunset        time.Time
}

const Kelvin = 273.15

var TestMode = false
var TestURL string
var HTTPClient = &http.Client{
	Timeout: time.Second * 5,
}

func enableTestMode(client *http.Client, url string) {
	TestMode = true
	TestURL = url
	HTTPClient = client
}

func getWeather(lat, lon float64, appID string) (Result, error) {
	weather := Result{}

	url := fmt.Sprintf(
		"http://api.openweathermap.org/data/2.5/weather?lat=%v&lon=%v&appid=%v",
		lat,
		lon,
		appID,
	)

	log.Printf("getting weather from %v", url)

	if TestMode {
		url = TestURL
	}

	resp, err := HTTPClient.Get(url)
	if err != nil {
		return weather, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return weather, err
	}

	log.Printf("body is %v", string(body))

	weatherResponse := Response{}

	err = json.Unmarshal(body, &weatherResponse)
	if err != nil {
		return weather, err
	}

	log.Printf("response is %+v", weatherResponse)

	weather = Result{
		Temp:          weatherResponse.Main.Temp - Kelvin,
		Pressure:      weatherResponse.Main.Pressure,
		Humidity:      weatherResponse.Main.Humidity,
		WindSpeed:     weatherResponse.Wind.Speed,
		WindDirection: weatherResponse.Wind.Deg,
		Sunrise:       time.Unix(weatherResponse.Sys.Sunrise, 0),
		Sunset:        time.Unix(weatherResponse.Sys.Sunset, 0),
	}

	log.Printf("weather is %+v", weather)

	return weather, nil
}

type Client struct {
	lat                  float64
	lon                  float64
	appID                string
	callsPerSecond       float64
	firstInteraction     bool
	lastCall             time.Time
	holdingWeatherResult bool
	lastWeatherResult    Result
	permissibleAge       time.Duration
}

func New(lat, lon float64, appID string, callsPerMinute int, permissibleAge time.Duration) Client {
	w := Client{
		lat:              lat,
		lon:              lon,
		appID:            appID,
		callsPerSecond:   float64(callsPerMinute) / 60,
		firstInteraction: true,
		permissibleAge:   permissibleAge,
	}

	log.Printf("created %+v", w)

	return w
}

func (c *Client) GetWeather() (Result, error) {
	log.Printf("weather requested")

	now := time.Now()

	if !c.firstInteraction {
		weatherResultAge := now.Sub(c.lastCall).Seconds()
		if weatherResultAge/(1/c.callsPerSecond) < 1 {
			log.Printf("rate limited; returning cached weather which is %+v", c.lastWeatherResult)

			return c.lastWeatherResult, nil
		}
	} else {
		c.firstInteraction = false
	}

	c.lastCall = now
	weatherResult, err := getWeather(c.lat, c.lon, c.appID)
	if err != nil {
		return c.lastWeatherResult, err
	}

	log.Printf("return latest weather which is %+v", weatherResult)

	c.lastWeatherResult = weatherResult
	c.holdingWeatherResult = true
	return c.lastWeatherResult, err
}

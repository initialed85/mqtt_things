package wunderground_weather_client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	httpClient *http.Client
)

func init() {
	httpClient = &http.Client{
		Timeout: time.Second * 60,
	}
}

type WeatherResponse struct {
	Observations []struct {
		StationID         string  `json:"stationID"`
		ObsTimeUtc        string  `json:"obsTimeUtc"`
		ObsTimeLocal      string  `json:"obsTimeLocal"`
		Neighborhood      string  `json:"neighborhood"`
		SoftwareType      any     `json:"softwareType"`
		Country           string  `json:"country"`
		SolarRadiation    any     `json:"solarRadiation"`
		Lon               float64 `json:"lon"`
		RealtimeFrequency any     `json:"realtimeFrequency"`
		Epoch             int64   `json:"epoch"`
		Lat               float64 `json:"lat"`
		Uv                any     `json:"uv"`
		Winddir           float64 `json:"winddir"`
		Humidity          float64 `json:"humidity"`
		QcStatus          int     `json:"qcStatus"`
		Metric            struct {
			Temp        float64 `json:"temp"`
			HeatIndex   float64 `json:"heatIndex"`
			Dewpt       float64 `json:"dewpt"`
			WindChill   float64 `json:"windChill"`
			WindSpeed   float64 `json:"windSpeed"`
			WindGust    float64 `json:"windGust"`
			Pressure    float64 `json:"pressure"`
			PrecipRate  float64 `json:"precipRate"`
			PrecipTotal float64 `json:"precipTotal"`
			Elev        float64 `json:"elev"`
		} `json:"metric"`
	} `json:"observations"`
}

type Weather struct {
	StationID            string
	Neighborhood         string
	Country              string
	Latitude             float64
	Longitude            float64
	Elevation            float64
	Time                 time.Time
	Temperature          float64
	ApparentTemperature  float64
	WindChillTemperature float64
	Humidity             float64
	WindSpeed            float64
	WindDirection        float64
	WindGust             float64
	AirPressure          float64
	PrecipitationRate    float64
	PrecipitationTotal   float64
}

type Forecast struct{}

func GetWeather(stationID string, apiKey string) (*Weather, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"https://api.weather.com/v2/pws/observations/current?stationId=%v&format=json&units=m&apiKey=%v",
			stationID,
			apiKey,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	// req.Header.Set("Accept-Encoding", "gzip")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wanted %v, got %v (body: %v)",
			http.StatusOK, resp.StatusCode, string(b),
		)
	}

	var response WeatherResponse

	err = json.Unmarshal(b, &response)
	if err != nil {
		return nil, err
	}

	if len(response.Observations) != 1 {
		return nil, fmt.Errorf(
			"expected exactly 1 observation, got %v; (body: %v)",
			len(response.Observations), string(b),
		)
	}

	observation := response.Observations[0]

	observationTime := time.Unix(observation.Epoch, 0)

	weather := Weather{
		StationID:            observation.StationID,
		Neighborhood:         observation.Neighborhood,
		Country:              observation.Country,
		Latitude:             observation.Lat,
		Longitude:            observation.Lon,
		Elevation:            observation.Metric.Elev,
		Time:                 observationTime,
		Temperature:          observation.Metric.Temp,
		ApparentTemperature:  observation.Metric.HeatIndex,
		WindChillTemperature: observation.Metric.WindChill,
		Humidity:             observation.Humidity,
		WindSpeed:            observation.Metric.WindSpeed,
		WindDirection:        observation.Winddir,
		WindGust:             observation.Metric.WindGust,
		AirPressure:          observation.Metric.Pressure,
		PrecipitationRate:    observation.Metric.PrecipRate,
		PrecipitationTotal:   observation.Metric.PrecipTotal,
	}

	return &weather, nil
}

func GetForecast(latitude float64, longitude float64, apiKey string) (*Forecast, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"https://api.weather.com/v3/wx/forecast/daily/5day?geocode=%v,%v&format=json&units=m&language=en-US&apiKey=%v",
			latitude,
			longitude,
			apiKey,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	// req.Header.Set("Accept-Encoding", "gzip")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wanted %v, got %v (body: %v)",
			http.StatusOK, resp.StatusCode, resp.Body,
		)
	}

	log.Printf("b: %v", string(b))

	return nil, nil
}

type Client struct {
	stationID        string
	apiKey           string
	latitude         float64
	longitude        float64
	permissibleAge   time.Duration
	mu               *sync.Mutex
	lastWeatherCall  time.Time
	lastWeather      *Weather
	lastForecastCall time.Time
	lastForecast     *Forecast
}

func New(stationID string, latitude float64, longitude float64, apiKey string, permissibleAge time.Duration) *Client {
	c := Client{
		stationID:      stationID,
		apiKey:         apiKey,
		latitude:       latitude,
		longitude:      longitude,
		permissibleAge: permissibleAge,
		mu:             new(sync.Mutex),
	}

	return &c
}

func (c *Client) GetWeather() (*Weather, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.lastWeatherCall) < c.permissibleAge {
		return c.lastWeather, nil
	}

	call := time.Now()

	weather, err := GetWeather(c.stationID, c.apiKey)
	if err != nil {
		return nil, err
	}

	c.lastWeatherCall = call
	c.lastWeather = weather

	return weather, nil
}

func (c *Client) GetForecast() (*Forecast, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.lastForecastCall) < c.permissibleAge {
		return c.lastForecast, nil
	}

	call := time.Now()

	forecast, err := GetForecast(c.latitude, c.longitude, c.apiKey)
	if err != nil {
		return nil, err
	}

	c.lastForecastCall = call
	c.lastForecast = forecast

	return forecast, nil
}

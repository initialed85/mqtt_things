package open_meteo_weather_client

import (
	"encoding/json"
	"net/http"
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
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	GenerationtimeMs     float64 `json:"generationtime_ms"`
	UtcOffsetSeconds     int     `json:"utc_offset_seconds"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Elevation            int     `json:"elevation"`
	DailyUnits           struct {
		Time                   string `json:"time"`
		WeatherCode            string `json:"weather_code"`
		Temperature2MMax       string `json:"temperature_2m_max"`
		Temperature2MMin       string `json:"temperature_2m_min"`
		ApparentTemperatureMax string `json:"apparent_temperature_max"`
		ApparentTemperatureMin string `json:"apparent_temperature_min"`
		Sunrise                string `json:"sunrise"`
		Sunset                 string `json:"sunset"`
		UvIndexMax             string `json:"uv_index_max"`
		UvIndexClearSkyMax     string `json:"uv_index_clear_sky_max"`
		PrecipitationSum       string `json:"precipitation_sum"`
		WindSpeed10MMax        string `json:"wind_speed_10m_max"`
	} `json:"daily_units"`
	Daily struct {
		Time                   []string  `json:"time"`
		WeatherCode            []int     `json:"weather_code"`
		Temperature2MMax       []float64 `json:"temperature_2m_max"`
		Temperature2MMin       []float64 `json:"temperature_2m_min"`
		ApparentTemperatureMax []float64 `json:"apparent_temperature_max"`
		ApparentTemperatureMin []float64 `json:"apparent_temperature_min"`
		Sunrise                []string  `json:"sunrise"`
		Sunset                 []string  `json:"sunset"`
		UvIndexMax             []float64 `json:"uv_index_max"`
		UvIndexClearSkyMax     []float64 `json:"uv_index_clear_sky_max"`
		PrecipitationSum       []int     `json:"precipitation_sum"`
		WindSpeed10MMax        []float64 `json:"wind_speed_10m_max"`
	} `json:"daily"`
}

type Weather struct {
	Latitude       float64
	Longitude      float64
	Time           time.Time
	TemperatureMin float64
	TemperatureMax float64
	Sunrise        time.Time
	Sunset         time.Time
	UVIndex        float64
	Precipitation  float64
	WindSpeed      float64
}

func Parse(b []byte) (*Weather, error) {
	var resp WeatherResponse
	err := json.Unmarshal(b, &resp)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

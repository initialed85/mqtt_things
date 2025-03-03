package open_meteo_weather_client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testData = []byte(`{
  "latitude": -31.875,
  "longitude": 115.875,
  "generationtime_ms": 0.0749826431274414,
  "utc_offset_seconds": 28800,
  "timezone": "Asia/Singapore",
  "timezone_abbreviation": "+08",
  "elevation": 26,
  "daily_units": {
    "time": "iso8601",
    "weather_code": "wmo code",
    "temperature_2m_max": "°C",
    "temperature_2m_min": "°C",
    "apparent_temperature_max": "°C",
    "apparent_temperature_min": "°C",
    "sunrise": "iso8601",
    "sunset": "iso8601",
    "uv_index_max": "",
    "uv_index_clear_sky_max": "",
    "precipitation_sum": "mm",
    "wind_speed_10m_max": "m/s"
  },
  "daily": {
    "time": [
      "2024-12-08"
    ],
    "weather_code": [
      2
    ],
    "temperature_2m_max": [
      29.5
    ],
    "temperature_2m_min": [
      18.4
    ],
    "apparent_temperature_max": [
      31.2
    ],
    "apparent_temperature_min": [
      16.5
    ],
    "sunrise": [
      "2024-12-08T05:03"
    ],
    "sunset": [
      "2024-12-08T19:13"
    ],
    "uv_index_max": [
      9.4
    ],
    "uv_index_clear_sky_max": [
      9.4
    ],
    "precipitation_sum": [
      0
    ],
    "wind_speed_10m_max": [
      5.59
    ]
  }
}`)

func TestOpenMeteoWeatherClient(t *testing.T) {
	t.Run("Parse", func(t *testing.T) {
		weather, err := Parse(testData)
		require.NoError(t, err)
	})
}

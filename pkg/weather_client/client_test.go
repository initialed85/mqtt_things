package weather_client

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type TestHTTPServer struct {
	CallCount int
	Client    *http.Client
	Close     func()
	URL       string
}

func (t *TestHTTPServer) Handle(w http.ResponseWriter, r *http.Request) {
	responseBody := `{"coord":{"lon":115.89,"lat":-31.92},"weather_client":[{"id":800,"main":"Clear","description":"clear sky","icon":"01n"}],"base":"stations","main":{"temp":296.13,"pressure":1012,"humidity":50,"temp_min":295.15,"temp_max":297.15},"visibility":10000,"wind":{"speed":4.1,"deg":220},"clouds":{"all":0},"dt":1547037000,"sys":{"type":1,"id":9586,"message":0.005,"country":"AU","sunrise":1546982423,"sunset":1547033185},"id":2066756,"name":"Maylands","cod":200}`

	w.Header().Set("Content-Type", "text/plain")
	_, err := fmt.Fprintln(w, responseBody)
	if err != nil {
		log.Fatal(err)
	}

	t.CallCount++
}

func getTestHTTPServer() *TestHTTPServer {
	t := TestHTTPServer{}

	s := httptest.NewServer(http.HandlerFunc(t.Handle))
	t.Client = s.Client()
	t.Close = s.Close
	t.URL = s.URL

	return &t
}

func TestGetWeather(t *testing.T) {
	testHTTPServer := getTestHTTPServer()
	defer testHTTPServer.Close()

	enableTestMode(testHTTPServer.Client, testHTTPServer.URL)

	weatherResult, err := getWeather(-31.923148, 115.894575, "d59f6762bf26d648a301f363cd84405f")
	if err != nil {
		log.Fatal(err)
		assert.Nil(t, err)
	}

	assert.Equal(t, 22.980000000000018, weatherResult.Temp)
	assert.Equal(t, 1012.0, weatherResult.Pressure)
	assert.Equal(t, 50, weatherResult.Humidity)
	assert.Equal(t, 4.1, weatherResult.WindSpeed)
	assert.Equal(t, 220.0, weatherResult.WindDirection)
	assert.Equal(t, time.Unix(1546982423, 0), weatherResult.Sunrise)
	assert.Equal(t, time.Unix(1547033185, 0), weatherResult.Sunset)
}

func TestNewWeather(t *testing.T) {
	testHTTPServer := getTestHTTPServer()
	defer testHTTPServer.Close()

	enableTestMode(testHTTPServer.Client, testHTTPServer.URL)

	weather := New(
		-31.923148,
		115.894575,
		"d59f6762bf26d648a301f363cd84405f",
		60,
		time.Second,
	)

	var weatherResult Result
	var err error

	for i := 0; i < 16; i++ {
		weatherResult, err = weather.GetWeather()
		if err != nil {
			log.Fatal(err)
			assert.Nil(t, err)
		}
	}

	time.Sleep(time.Second)

	for i := 0; i < 16; i++ {
		weatherResult, err = weather.GetWeather()
		if err != nil {
			log.Fatal(err)
			assert.Nil(t, err)
		}
	}

	assert.Equal(t, 22.980000000000018, weatherResult.Temp)
	assert.Equal(t, 1012.0, weatherResult.Pressure)
	assert.Equal(t, 50, weatherResult.Humidity)
	assert.Equal(t, 4.1, weatherResult.WindSpeed)
	assert.Equal(t, 220.0, weatherResult.WindDirection)
	assert.Equal(t, time.Unix(1546982423, 0), weatherResult.Sunrise)
	assert.Equal(t, time.Unix(1547033185, 0), weatherResult.Sunset)

	assert.Equal(t, 2, testHTTPServer.CallCount)
}

package wunderground_weather_server

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/initialed85/mqtt_things/internal/hack"
)

/*
dig rtupdate.wunderground.com @1.1.1.1
...
rtupdate.wunderground.com. 39	IN	CNAME	pws-ingest-use1-01.sun.weather.com.
pws-ingest-use1-01.sun.weather.com. 279	IN CNAME a60f001b0bdcd4f64a49719eb2307270-ee5cf64f94dbf2ed.elb.us-east-1.amazonaws.com.
a60f001b0bdcd4f64a49719eb2307270-ee5cf64f94dbf2ed.elb.us-east-1.amazonaws.com. 39 IN A 52.22.134.222
a60f001b0bdcd4f64a49719eb2307270-ee5cf64f94dbf2ed.elb.us-east-1.amazonaws.com. 39 IN A 54.159.105.134
a60f001b0bdcd4f64a49719eb2307270-ee5cf64f94dbf2ed.elb.us-east-1.amazonaws.com. 39 IN A 34.232.250.77
*/

const upstreamURL = "http://pws-ingest-use1-01.sun.weather.com" // because we've spoofed rtupdate.wunderground.com
const upstreamHost = "rtupdate.wunderground.com"                // in case they insist on this host header at some point

/*
method: GET

url: /weatherstation/updateweatherstation.php?ID=IPERTH3490&PASSWORD=123abc&action=updateraww&realtime=1&rtfreq=5&dateutc=now&baromin=29.82&tempf=67.1&dewptf=61.5&humidity=82&windspeedmph=1.3&windgustmph=1.3&winddir=225&rainin=0.0&dailyrainin=0.09&indoortempf=76.2&indoorhumidity=64

header: Connection = "keep-alive"

query: ID = "IPERTH3490"
query: PASSWORD = "123abc"

query: action = "updateraww"
query: dateutc = "now"
query: realtime = "1"
query: rtfreq = "5"

query: baromin = "29.82"
query: dailyrainin = "0.09"
query: dewptf = "61.5"
query: humidity = "82"
query: indoorhumidity = "64"
query: indoortempf = "76.2"
query: rainin = "0.0"
query: tempf = "67.1"
query: winddir = "225"
query: windgustmph = "1.3"
query: windspeedmph = "1.3"
*/

type Weather struct {
	Timestamp         time.Time // UTC
	StationID         string    // as given to Weather Underground
	Latitude          float64   // injected, doesn't come from weather station
	Longitude         float64   // injected, doesn't come from weather station
	Temperature       float64   // deg C
	DewPoint          float64   // deg C
	Humidity          float64   // %
	WindSpeed         float64   // m/s instantaneous
	WindDirection     float64   // deg
	WindGust          float64   // m/s max over a vendor-determined period
	AirPressure       float64   // hPa
	RainLast60Mins    float64   // mm last 60 minutes
	RainToday         float64   // mm since 00:00 local time
	TemperatureIndoor float64   // deg C
	HumidityIndoor    float64   // %
}

func Run(
	ctx context.Context,
	port uint16,
	latitude float64,
	longitude float64,
	handleWeather func(Weather),
) error {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Second * 10,
	}

	httpServer := &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%v", port),
		Handler: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				var err error
				var status int = http.StatusOK
				var payload = "success"

				defer func() {
					if err != nil {
						payload = err.Error()
						if status == http.StatusOK {
							status = http.StatusInternalServerError
						}
					}

					w.WriteHeader(status)
					w.Write([]byte(payload))
				}()

				if r.Method != http.MethodGet {
					err = fmt.Errorf("error: method was %v (wanted %v)", r.Method, http.MethodGet)
					status = http.StatusMethodNotAllowed
					return
				}

				if r.URL.Path != "/weatherstation/updateweatherstation.php" {
					err = fmt.Errorf("error: unknown path %v", r.URL.String())
					status = http.StatusNotFound
					return
				}

				log.Printf(">>> %v %v", r.Method, r.URL.String())

				var req *http.Request
				req, err = http.NewRequest(
					http.MethodGet,
					fmt.Sprintf("%v%v", upstreamURL, r.URL.String()),
					nil,
				)
				if err != nil {
					log.Printf("failed to build request for upstream URL %v; %v", upstreamURL, err)
					err = nil
				} else {
					req.Header.Add("Host", upstreamHost)
					log.Printf("<<< %v %v", req.Method, req.URL.String())
					var resp *http.Response
					resp, err = httpClient.Do(req)
					var b []byte
					if resp != nil && resp.Body != nil {
						b, _ = io.ReadAll(resp.Body)
						defer func() {
							_ = resp.Body.Close()
						}()
					}
					if err == nil && resp.StatusCode != http.StatusOK {
						err = fmt.Errorf("status was %v (wanted %v); body: %v", resp.Status, http.StatusOK, string(b))
					}
					if err != nil {
						log.Printf("failed to do request to upstream URL %v; %v", upstreamURL, err)
						err = nil
					}
				}

				rawItems := make(map[string]any)

				for k, vs := range r.URL.Query() {
					for _, stringV := range vs {
						floatV, err := strconv.ParseFloat(stringV, 64)
						if err != nil {
							rawItems[k] = stringV
							continue
						}

						rawItems[k] = floatV
					}
				}

				log.Printf("rawItems: %v", hack.UnsafeJSONPrettyFormat(rawItems))

				weather := Weather{
					Timestamp: time.Now().UTC(),
					Latitude:  latitude,
					Longitude: longitude,
				}

				var ok bool

				weather.StationID, _ = rawItems["ID"].(string)

				weather.Temperature, ok = rawItems["tempf"].(float64)
				if ok {
					weather.Temperature -= 32
					weather.Temperature *= (5.0 / 9.0)
				}

				weather.DewPoint, ok = rawItems["dewptf"].(float64)
				if ok {
					weather.DewPoint -= 32
					weather.DewPoint *= (5.0 / 9.0)
				}

				weather.Humidity, _ = rawItems["humidity"].(float64)

				weather.WindSpeed, ok = rawItems["windspeedmph"].(float64)
				if ok {
					weather.WindSpeed *= 1.6093444979
				}

				weather.WindDirection, _ = rawItems["winddir"].(float64)

				weather.WindGust, ok = rawItems["windgustmph"].(float64)
				if ok {
					weather.WindGust *= 1.6093444979
				}

				weather.AirPressure, ok = rawItems["baromin"].(float64)
				if ok {
					weather.AirPressure *= 33.863889532610884
				}

				weather.RainLast60Mins, ok = rawItems["rainin"].(float64)
				if ok {
					weather.RainLast60Mins *= 25.4
				}

				weather.RainToday, ok = rawItems["dailyrainin"].(float64)
				if ok {
					weather.RainToday *= 25.4
				}

				weather.TemperatureIndoor, ok = rawItems["indoortempf"].(float64)
				if ok {
					weather.TemperatureIndoor -= 32
					weather.TemperatureIndoor *= (5.0 / 9.0)
				}

				weather.HumidityIndoor, _ = rawItems["indoorhumidity"].(float64)

				log.Printf("weather: %v", hack.UnsafeJSONPrettyFormat(weather))

				handleWeather(weather)
			},
		),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*1)
		defer shutdownCancel()
		_ = httpServer.Shutdown(shutdownCtx)
		_ = httpServer.Close()
	}()

	log.Printf("listening on %v", httpServer.Addr)

	err := httpServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

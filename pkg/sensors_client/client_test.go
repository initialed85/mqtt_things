package sensors_client

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testResponseBody = `{"1":{"state":{"daylight":false,"lastupdated":"2021-01-04T11:11:00"},"config":{"on":true,"configured":true,"sunriseoffset":15,"sunsetoffset":-15},"name":"Daylight","type":"Daylight","modelid":"PHDL00","manufacturername":"Signify Netherlands B.V.","swversion":"1.0"},"4":{"state":{"presence":false,"lastupdated":"2019-04-17T22:54:00"},"config":{"on":true,"reachable":true},"name":"HomeAway","type":"CLIPPresence","modelid":"HOMEAWAY","manufacturername":"1472245522adec65ac9ac526b811fae","swversion":"A_1","uniqueid":"L_01_Hswwo","recycle":false},"7":{"state":{"status":0,"lastupdated":"2019-04-17T22:54:00"},"config":{"on":true,"reachable":true},"name":"Play and stop","type":"CLIPGenericStatus","modelid":"HUELABSVDIMMER","manufacturername":"Philips","swversion":"1.0","uniqueid":"9222-e683-4725-835a","recycle":true},"8":{"state":{"status":1,"lastupdated":"2019-04-18T13:00:27"},"config":{"on":true,"reachable":true},"name":"player","type":"CLIPGenericStatus","modelid":"HUELABSSTOGGLE","manufacturername":"Philips","swversion":"1.0","uniqueid":"2:435c-5c50-4ddd-9b40","recycle":true},"14":{"state":{"presence":null,"lastupdated":"none"},"config":{"on":false,"reachable":true},"name":"Google Pixel 3","type":"Geofence","modelid":"HA_GEOFENCE","manufacturername":"Philips","swversion":"A_1","uniqueid":"L_02_MC1v1","recycle":true},"15":{"state":{"presence":false,"lastupdated":"2021-01-04T14:49:56"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:49"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","sensitivity":2,"sensitivitymax":2,"ledindication":false,"usertest":false,"pending":[]},"name":"Bedroom","type":"ZLLPresence","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue motion sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:e6:3e-02-0406","capabilities":{"certified":true,"primary":true}},"16":{"state":{"lightlevel":0,"dark":true,"daylight":false,"lastupdated":"2021-01-04T14:49:35"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:49"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","tholddark":16000,"tholdoffset":7000,"ledindication":false,"usertest":false,"pending":[]},"name":"Hue ambient light sensor 1","type":"ZLLLightLevel","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue ambient light sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:e6:3e-02-0400","capabilities":{"certified":true,"primary":false}},"17":{"state":{"temperature":2224,"lastupdated":"2021-01-04T14:49:36"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:49"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","ledindication":false,"usertest":false,"pending":[]},"name":"Hue temperature sensor 1","type":"ZLLTemperature","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue temperature sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:e6:3e-02-0402","capabilities":{"certified":true,"primary":false}},"18":{"state":{"presence":false,"lastupdated":"2021-01-04T14:36:38"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:43"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","sensitivity":2,"sensitivitymax":2,"ledindication":false,"usertest":false,"pending":[]},"name":"Living room","type":"ZLLPresence","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue motion sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:fb:2e-02-0406","capabilities":{"certified":true,"primary":true}},"19":{"state":{"lightlevel":0,"dark":true,"daylight":false,"lastupdated":"2021-01-04T14:49:32"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:43"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","tholddark":16000,"tholdoffset":7000,"ledindication":false,"usertest":false,"pending":[]},"name":"Hue ambient light sensor 2","type":"ZLLLightLevel","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue ambient light sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:fb:2e-02-0400","capabilities":{"certified":true,"primary":false}},"20":{"state":{"temperature":2452,"lastupdated":"2021-01-04T14:49:32"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:43"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","ledindication":false,"usertest":false,"pending":[]},"name":"Hue temperature sensor 2","type":"ZLLTemperature","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue temperature sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:fb:2e-02-0402","capabilities":{"certified":true,"primary":false}},"21":{"state":{"presence":false,"lastupdated":"2021-01-04T14:43:03"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:36"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","sensitivity":2,"sensitivitymax":2,"ledindication":false,"usertest":false,"pending":[]},"name":"Kitchen","type":"ZLLPresence","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue motion sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:f4:19-02-0406","capabilities":{"certified":true,"primary":true}},"22":{"state":{"lightlevel":0,"dark":true,"daylight":false,"lastupdated":"2021-01-04T14:48:36"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:36"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","tholddark":16000,"tholdoffset":7000,"ledindication":false,"usertest":false,"pending":[]},"name":"Hue ambient light sensor 3","type":"ZLLLightLevel","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue ambient light sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:f4:19-02-0400","capabilities":{"certified":true,"primary":false}},"23":{"state":{"temperature":2607,"lastupdated":"2021-01-04T14:49:28"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:36"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","ledindication":false,"usertest":false,"pending":[]},"name":"Hue temperature sensor 3","type":"ZLLTemperature","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue temperature sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:f4:19-02-0402","capabilities":{"certified":true,"primary":false}},"26":{"state":{"presence":false,"lastupdated":"2021-01-04T10:59:44"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:29"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","sensitivity":2,"sensitivitymax":2,"ledindication":false,"usertest":false,"pending":[]},"name":"Kid's room","type":"ZLLPresence","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue motion sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:e9:1d-02-0406","capabilities":{"certified":true,"primary":true}},"27":{"state":{"lightlevel":0,"dark":true,"daylight":false,"lastupdated":"2021-01-04T14:49:22"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:29"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","tholddark":16000,"tholdoffset":7000,"ledindication":false,"usertest":false,"pending":[]},"name":"Hue ambient light sensor 4","type":"ZLLLightLevel","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue ambient light sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:e9:1d-02-0400","capabilities":{"certified":true,"primary":false}},"28":{"state":{"temperature":2509,"lastupdated":"2021-01-04T14:49:23"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:29"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","ledindication":false,"usertest":false,"pending":[]},"name":"Hue temperature sensor 4","type":"ZLLTemperature","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue temperature sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:e9:1d-02-0402","capabilities":{"certified":true,"primary":false}},"29":{"state":{"presence":true,"lastupdated":"2021-01-04T14:51:35"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:23"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","sensitivity":2,"sensitivitymax":2,"ledindication":false,"usertest":false,"pending":[]},"name":"Office","type":"ZLLPresence","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue motion sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:fc:3c-02-0406","capabilities":{"certified":true,"primary":true}},"30":{"state":{"lightlevel":1215,"dark":true,"daylight":false,"lastupdated":"2021-01-04T14:49:28"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:23"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","tholddark":16000,"tholdoffset":7000,"ledindication":false,"usertest":false,"pending":[]},"name":"Hue ambient light sensor 5","type":"ZLLLightLevel","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue ambient light sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:fc:3c-02-0400","capabilities":{"certified":true,"primary":false}},"31":{"state":{"temperature":2668,"lastupdated":"2021-01-04T14:49:15"},"swupdate":{"state":"noupdates","lastinstall":"2021-01-04T13:54:23"},"config":{"on":true,"battery":100,"reachable":true,"alert":"none","ledindication":false,"usertest":false,"pending":[]},"name":"Hue temperature sensor 5","type":"ZLLTemperature","modelid":"SML001","manufacturername":"Signify Netherlands B.V.","productname":"Hue temperature sensor","swversion":"6.1.1.27575","uniqueid":"00:17:88:01:09:15:fc:3c-02-0402","capabilities":{"certified":true,"primary":false}},"34":{"state":{"status":0,"lastupdated":"2021-01-04T14:49:56"},"config":{"on":true,"reachable":true},"name":"presenceState","type":"CLIPGenericStatus","modelid":"HUELABSENUM","manufacturername":"Philips","swversion":"1.0","uniqueid":"5:12:cdfb-68a1-488e-82f0","recycle":true},"35":{"state":{"status":0,"lastupdated":"2021-01-04T14:49:46"},"config":{"on":true,"reachable":true},"name":"textState","type":"CLIPGenericStatus","modelid":"BEH_STATE","manufacturername":"Philips","swversion":"1.0","uniqueid":"2:13:d6b8-54b0-47e4-bc2f","recycle":true},"36":{"state":{"status":0,"lastupdated":"2021-01-04T14:36:38"},"config":{"on":true,"reachable":true},"name":"presenceState","type":"CLIPGenericStatus","modelid":"HUELABSENUM","manufacturername":"Philips","swversion":"1.0","uniqueid":"5:12:5278-8689-4bd8-b5cb","recycle":true},"37":{"state":{"status":0,"lastupdated":"2021-01-04T14:36:11"},"config":{"on":true,"reachable":true},"name":"textState","type":"CLIPGenericStatus","modelid":"BEH_STATE","manufacturername":"Philips","swversion":"1.0","uniqueid":"2:13:197f-afcb-4591-a945","recycle":true},"38":{"state":{"presence":true,"lastupdated":"none"},"config":{"on":true,"reachable":true},"name":"Google Pixel 4a","type":"Geofence","modelid":"HA_GEOFENCE","manufacturername":"Philips","swversion":"A_1","uniqueid":"L_02_VXDZx","recycle":true}}`

type TestHTTPServer struct {
	CallCount int
	Client    *http.Client
	Close     func()
	URL       string
}

func (t *TestHTTPServer) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := fmt.Fprintln(w, testResponseBody)
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

func Test_getSensors(t *testing.T) {
	testHTTPServer := getTestHTTPServer()
	defer testHTTPServer.Close()

	enableTestMode(testHTTPServer.Client, testHTTPServer.URL)

	sensors, err := getSensors(
		"192.168.137.252",
		"some_app_id",
	)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(
		t,
		[]BlendedSensor{
			{ID: 15, Name: "Bedroom", Presence: false, LightLevel: 0, Dark: true, Daylight: false, Temperature: 22.24},
			{ID: 18, Name: "Living room", Presence: false, LightLevel: 0, Dark: true, Daylight: false, Temperature: 24.52},
			{ID: 21, Name: "Kitchen", Presence: false, LightLevel: 0, Dark: true, Daylight: false, Temperature: 26.07},
			{ID: 26, Name: "Kid's room", Presence: false, LightLevel: 0, Dark: true, Daylight: false, Temperature: 25.09},
			{ID: 29, Name: "Office", Presence: true, LightLevel: 1215, Dark: true, Daylight: false, Temperature: 26.68},
		},
		sensors,
	)
}

func TestNew(t *testing.T) {
	testHTTPServer := getTestHTTPServer()
	defer testHTTPServer.Close()

	enableTestMode(testHTTPServer.Client, testHTTPServer.URL)

	c := New(
		"192.168.137.252",
		"some_app_id",
	)

	sensors, err := c.GetSensors()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(
		t,
		[]BlendedSensor{
			{ID: 15, Name: "Bedroom", Presence: false, LightLevel: 0, Dark: true, Daylight: false, Temperature: 22.24},
			{ID: 18, Name: "Living room", Presence: false, LightLevel: 0, Dark: true, Daylight: false, Temperature: 24.52},
			{ID: 21, Name: "Kitchen", Presence: false, LightLevel: 0, Dark: true, Daylight: false, Temperature: 26.07},
			{ID: 26, Name: "Kid's room", Presence: false, LightLevel: 0, Dark: true, Daylight: false, Temperature: 25.09},
			{ID: 29, Name: "Office", Presence: true, LightLevel: 1215, Dark: true, Daylight: false, Temperature: 26.68},
		},
		sensors,
	)
}

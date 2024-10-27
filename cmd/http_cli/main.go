package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
)

type TopicPayload struct {
	LastUpdated time.Time `json:"last_updated"`
	Payload     string    `json:"payload"`
}

type ErrorResponse struct {
	StatusCode int    `json:"status_code"`
	Reason     string `json:"reason"`
}

const (
	topic = "+/+/#"
)

var (
	substitutes       = []string{"home/", "outside/", "inside/", "globe/", "bank/", "state/", "/get"}
	lock              sync.RWMutex
	topicPayloadByUrl = map[string]TopicPayload{}
)

func callback(message mqtt.Message) {
	if !strings.HasSuffix(message.Topic, "/get") {
		return
	}

	url := message.Topic
	for _, substitute := range substitutes {
		url = strings.ReplaceAll(url, substitute, "")
	}
	url = fmt.Sprintf("/%v", url)

	lock.Lock()
	topicPayloadByUrl[url] = TopicPayload{time.Now(), message.Payload}
	lock.Unlock()
}

func writeError(statusCode int, w http.ResponseWriter, reason string) {
	buf, err := json.MarshalIndent(ErrorResponse{statusCode, reason}, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	_, err = fmt.Fprintf(w, "%v\n", string(buf))
	if err != nil {
		log.Fatal(err)
	}
}

func writeAll(w http.ResponseWriter) {

	lock.RLock()
	buf, err := json.MarshalIndent(topicPayloadByUrl, "", "    ")
	lock.RUnlock()
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprintf(w, "%v\n", string(buf))
	if err != nil {
		log.Fatal(err)
	}
}

func writeTopicPayload(w http.ResponseWriter, topicPayload TopicPayload) {
	buf, err := json.MarshalIndent(topicPayload, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprintf(w, "%v\n", string(buf))
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if req.Method != http.MethodGet {
		writeError(http.StatusMethodNotAllowed, w, fmt.Sprintf("Method %v not allowed", req.Method))

		return
	}

	url := strings.TrimRight(req.URL.String(), "/")
	if url == "/__all__" || strings.TrimSpace(url) == "" {
		writeAll(w)

		return
	}

	lock.RLock()
	topicPayload, ok := topicPayloadByUrl[url]
	lock.RUnlock()
	if !ok {
		writeError(http.StatusBadRequest, w, fmt.Sprintf("TopicPayload for %v not known", url))

		return
	}

	writeTopicPayload(w, topicPayload)
}

func main() {
	hostPtr := flag.String("host", "", "mqtt broker host")
	usernamePtr := flag.String("username", "", "mqtt username")
	passwordPtr := flag.String("password", "", "mqtt password")
	portPtr := flag.Int("port", -1, "http port")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if *hostPtr == "" {
		log.Fatal("host flag empty")
	}

	if *portPtr == -1 {
		log.Fatal("port flag empty")
	}

	mqttClient := mqtt.GetMQTTClient(*hostPtr, *usernamePtr, *passwordPtr)
	err := mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 16)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err = mqttClient.Disconnect()
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	err = mqttClient.Subscribe(topic, mqtt.ExactlyOnce, callback)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handler)

	err = http.ListenAndServe(fmt.Sprintf(":%v", *portPtr), nil)
	if err != nil {
		log.Fatal(err)
	}
}

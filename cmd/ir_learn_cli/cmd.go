package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/initialed85/mqtt_things/pkg/smart_aircons_client"
)

type Code struct {
	Name        string
	StringValue string
	ByteValue   []byte
}

var names = map[string]struct{}{
	"off":      {},
	"fan_only": {},
	"cool_18":  {},
	"cool_19":  {},
	"cool_20":  {},
	"cool_21":  {},
	"cool_22":  {},
	"cool_23":  {},
	"cool_24":  {},
	"cool_25":  {},
	"cool_26":  {},
	"cool_27":  {},
	"cool_28":  {},
	"cool_29":  {},
	"cool_30":  {},
	"heat_18":  {},
	"heat_19":  {},
	"heat_20":  {},
	"heat_21":  {},
	"heat_22":  {},
	"heat_23":  {},
	"heat_24":  {},
	"heat_25":  {},
	"heat_26":  {},
	"heat_27":  {},
	"heat_28":  {},
	"heat_29":  {},
	"heat_30":  {},
}

func main() {
	irType := strings.TrimSpace(os.Getenv("IR_TYPE"))
	irHost := strings.TrimSpace(os.Getenv("IR_HOST"))

	if irType == "" {
		log.Fatal("IR_TYPE env var empty or unset")
	}

	if irHost == "" {
		log.Fatal("IR_HOST env var empty or unset")
	}

	var learnIR func(string) ([]byte, error)

	switch irType {
	case "zmote":
		learnIR = smart_aircons_client.ZmoteLearnIR
	case "broadlink":
		learnIR = smart_aircons_client.BroadlinkLearnIR
	default:
		log.Fatalf("IR_TYPE env var unknown: %#+v", irType)
	}

	r := bufio.NewReader(os.Stdin)

	codeByName := make(map[string][]byte)

	dump := func() {
		fmt.Printf("var CodeByNameForXYZ = map[string][]byte{\n")
		for name, code := range codeByName {
			fmt.Printf("    %#+v: %v,\n", name, strings.ReplaceAll(fmt.Sprintf("%#+v", code), "[]byte", ""))
		}
		fmt.Printf("}\n")
	}

	defer func() {
		dump()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		<-c
		fmt.Printf("\n")
		dump()
		os.Exit(0)
	}()
	runtime.Gosched()

	// loop forever (to save multiple codes)
outer:
	for {
		log.Printf("waiting to learn a code...")

		var code []byte
		var err error

		// loop until a code is received
	receive:
		for {
			deferUntil := time.Now().Add(time.Millisecond * 100)
			code, err = learnIR(irHost)
			if err != nil {
				time.Sleep(time.Until(deferUntil))
				continue receive
			}

			log.Printf("learned code: %v", code)

			break receive
		}

		if code == nil || len(code) == 0 {
			log.Fatalf("code unexpectedly empty; last error was %v", err)
		}

		// loop until a valid name is given
	record:
		for {
			fmt.Printf("enter a name: ")
			b, _, err := r.ReadLine()
			if err != nil {
				log.Fatalf("failed to read user input: %v", err)
			}

			name := strings.TrimSpace(string(b))

			if name == "" {
				log.Printf("ignored %v", name)
				continue outer
			}

			if name == "exit" {
				log.Printf("done.")
				break outer
			}

			_, ok := names[name]
			if !ok {
				log.Printf("error: name %#+v unknown", name)
				continue record
			}

			_, overwritten := codeByName[name]
			codeByName[name] = code

			if overwritten {
				log.Printf("replaced %v", name)
			} else {
				log.Printf("saved %v", name)
			}

			break record
		}
	}
}

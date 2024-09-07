package smart_aircons_client

import (
	"fmt"
	"log"
)

var allCodes = map[string]map[string][]byte{}

func GetCode(name string, on bool, mode string, temperature int64) ([]byte, error) {
	var ok bool
	var codes map[string][]byte
	var code []byte

	codes, ok = allCodes[name]
	if !ok {
		return nil, fmt.Errorf("%#+v not a recognized name", name)
	}

	var codeName = ""

	if !on || mode == "off" {
		mode = "off"
		codeName = mode
	} else {
		if mode == "fan_only" {
			mode = "fan_only"
			codeName = mode
		} else {
			codeName = fmt.Sprintf("%v_%v", mode, temperature)
		}
	}

	code, ok = codes[codeName]
	if !ok {
		return nil, fmt.Errorf("%#+v not a recognized code for %#+v", codeName, name)
	}

	log.Printf("name=%#+v, codeName=%#+v, on=%#+v, mode=%#+v, temperature=%#+v, code=%#+v", name, codeName, on, mode, temperature, code)

	return code, nil
}

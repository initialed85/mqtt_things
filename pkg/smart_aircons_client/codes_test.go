package smart_aircons_client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCode(t *testing.T) {
	var code string
	var err error

	code, err = GetCode("fujitsu", false, "cool", 24)
	handleError(err)
	assert.Equal(t, allCodes["fujitsu"]["off"], code)

	code, err = GetCode("fujitsu", true, "fan_only", 24)
	handleError(err)
	assert.Equal(t, allCodes["fujitsu"]["fan_only"], code)

	code, err = GetCode("fujitsu", true, "cool", 24)
	handleError(err)
	assert.Equal(t, allCodes["fujitsu"]["cool_24"], code)

	code, err = GetCode("fujitsu", true, "heat", 28)
	handleError(err)
	assert.Equal(t, allCodes["fujitsu"]["heat_28"], code)

	_, err = GetCode("fujitsu", true, "ham", 28)
	assert.NotNil(t, err)

	_, err = GetCode("HAM", true, "ham", 28)
	assert.NotNil(t, err)
}

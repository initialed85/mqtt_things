package mqtt_client

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func getClientID(provider string) string {
	identifier := fmt.Sprintf("unknown_%+v", time.Now().UnixNano())

	uuid4, err := uuid.NewRandom()
	if err == nil {
		identifier = uuid4.String()
	}

	return fmt.Sprintf("%v_%v", provider, identifier)
}

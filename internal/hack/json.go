package hack

import "encoding/json"

func UnsafeJSONPrettyFormat(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")

	return string(b)
}

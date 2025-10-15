package types

import "time"

// ProgressMessage is the JSON-friendly representation of a deploy progress event.
type ProgressMessage struct {
	Phase   string            `json:"phase"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
	Time    time.Time         `json:"time"`
}

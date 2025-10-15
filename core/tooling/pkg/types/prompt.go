package types

import "context"

// Prompter handles interactive authentication prompts.
type Prompter interface {
	Notify(context.Context, PromptMessage) error
	PromptSecret(context.Context, PromptMessage) (string, error)
}

// PromptKind identifies the type of prompt interaction requested.
type PromptKind string

const (
	PromptKindInfo  PromptKind = "info"
	PromptKindToken PromptKind = "token"
	PromptKindLink  PromptKind = "link"
)

// PromptMessage describes a prompt interaction for authentication flows.
type PromptMessage struct {
	ID               string     `json:"id"`
	Provider         string     `json:"provider"`
	Kind             PromptKind `json:"kind"`
	Message          string     `json:"message,omitempty"`
	URL              string     `json:"url,omitempty"`
	Scopes           []string   `json:"scopes,omitempty"`
	RequiresResponse bool       `json:"requires_response"`
}

// PromptResponse carries the response to a PromptMessage.
type PromptResponse struct {
	ID     string `json:"id"`
	Secret string `json:"secret,omitempty"`
	Error  string `json:"error,omitempty"`
}

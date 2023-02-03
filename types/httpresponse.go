package types

type HttpReturn struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

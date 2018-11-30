package payloads

type ApiError struct {
	Message string `json:"message"`
	Slug    string `json:"slug,omitempty"`
}

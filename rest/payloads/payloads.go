package payloads

import "fmt"

type ApiError struct {
	Message string `json:"message"`
	Slug    string `json:"slug,omitempty"`
}

func (a *ApiError) Error() string {
	return fmt.Sprintf("Got error of type: [%s], with message [%s]", a.Slug, a.Message)
}

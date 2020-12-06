package model

import "time"

type PrincipalType string

const (
	UserPrincipalType         PrincipalType = "user"
	OrganizationPrincipalType PrincipalType = "organization"
)

type User struct {
	ID AccountID `json:"id"`

	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`

	CreatedAt time.Time `json:"-"`
}

type Organization struct {
	PrincipalID AccountID `json:"id"`

	Name string `json:"name"`
	Logo string `json:"logo,omitempty"`
}

type Principal struct {
	ID AccountID `json:"id"`

	Username     string        `json:"username"`
	Organization *Organization `json:"organization,omitempty" `
	User         *User         `json:"user,omitempty"`

	Type PrincipalType `json:"type"`
}

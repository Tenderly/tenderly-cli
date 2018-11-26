package model

import "time"

type User struct {
	ID AccountID `json:"id"`

	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`

	CreatedAt time.Time `json:"-"`
}

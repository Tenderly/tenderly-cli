package model

import "time"

type AccountID string

type ProjectID struct {
	AccountID AccountID
	Name      string
}

type Project struct {
	ID        AccountID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Owner     AccountID
	CreatedAt time.Time `json:"created_at"`
}

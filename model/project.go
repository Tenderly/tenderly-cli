package model

import "time"

type AccountID string

func (a AccountID) String() string {
	return string(a)
}

type ProjectID struct {
	AccountID AccountID
	Name      string
}

type ProjectPermissions struct {
	AddContract bool `json:"add_contract"`
}

type Project struct {
	ID        AccountID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Owner     AccountID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`

	Permissions *ProjectPermissions `json:"permissions,omitempty"`

	IsShared bool `json:"-"`
}

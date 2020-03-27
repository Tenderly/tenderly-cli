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

type OwnerInfo struct {
	ID        AccountID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
}

type Project struct {
	ID        AccountID  `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	Owner     AccountID  `json:"owner_id"`
	OwnerInfo *OwnerInfo `json:"owner"`
	CreatedAt time.Time  `json:"created_at"`

	Permissions *ProjectPermissions `json:"permissions,omitempty"`

	IsShared bool `json:"-"`
}

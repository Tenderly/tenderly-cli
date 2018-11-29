package rest

import (
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/call"
)

type AuthRoutes interface {
	Register(request call.RegisterRequest) (*call.TokenResponse, error)
	Login(request call.LoginRequest) (*call.TokenResponse, error)
}

type UserRoutes interface {
	User() (*model.User, error)
}

type ProjectRoutes interface {
	CreateProject(request call.ProjectRequest) (*model.Project, error)
	GetProject(accountId, id string) (*model.Project, error)
	GetProjects(accountId string) ([]*model.Project, error)
}

type ContractRoutes interface {
	UploadContracts(request call.UploadContractsRequest) ([]*model.Contract, error)
}

type Rest struct {
	Auth     AuthRoutes
	User     UserRoutes
	Project  ProjectRoutes
	Contract ContractRoutes
}

func NewRest(
	auth AuthRoutes,
	user UserRoutes,
	project ProjectRoutes,
	contract ContractRoutes) *Rest {
	return &Rest{
		Auth:     auth,
		User:     user,
		Project:  project,
		Contract: contract,
	}
}

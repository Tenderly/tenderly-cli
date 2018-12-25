package rest

import (
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/payloads"
)

type AuthRoutes interface {
	Register(request payloads.RegisterRequest) (*payloads.TokenResponse, error)
	Login(request payloads.LoginRequest) (*payloads.TokenResponse, error)
}

type UserRoutes interface {
	User() (*model.User, error)
}

type ProjectRoutes interface {
	CreateProject(request payloads.ProjectRequest) (*payloads.ProjectResponse, error)
	GetProject(accountId, id string) (*model.Project, error)
	GetProjects(accountId string) (*payloads.GetProjectsResponse, error)
}

type ContractRoutes interface {
	UploadContracts(request payloads.UploadContractsRequest) (*payloads.UploadContractsResponse, error)
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

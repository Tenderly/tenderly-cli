package rest

import (
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/payloads"
)

type AuthRoutes interface {
	Register(request payloads.RegisterRequest) (*payloads.TokenResponse, error)
	Login(request payloads.LoginRequest) (*payloads.TokenResponse, error)
	Logout(accountId string, tokenId string) error
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
	UploadContracts(request payloads.UploadContractsRequest, projectSlug string) (*payloads.UploadContractsResponse, error)
	VerifyContracts(request payloads.UploadContractsRequest) (*payloads.UploadContractsResponse, error)
}

type ExportRoutes interface {
	ExportTransaction(request payloads.ExportTransactionRequest, projectSlug string) (*payloads.ExportTransactionResponse, error)
}

type Rest struct {
	Auth     AuthRoutes
	User     UserRoutes
	Project  ProjectRoutes
	Contract ContractRoutes
	Export   ExportRoutes
}

func NewRest(
	auth AuthRoutes,
	user UserRoutes,
	project ProjectRoutes,
	contract ContractRoutes,
	export ExportRoutes,
) *Rest {
	return &Rest{
		Auth:     auth,
		User:     user,
		Project:  project,
		Contract: contract,
		Export:   export,
	}
}

package rest

import (
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	generatedActions "github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"
)

type AuthRoutes interface {
	Register(request payloads.RegisterRequest) (*payloads.TokenResponse, error)
	Login(request payloads.LoginRequest) (*payloads.TokenResponse, error)
	Logout(accountId string, tokenId string) error
}

type UserRoutes interface {
	Principal() (*model.Principal, error)
}

type ProjectRoutes interface {
	CreateProject(request payloads.ProjectRequest) (*payloads.ProjectResponse, error)
	GetProject(accountId, id string) (*model.Project, error)
	GetProjects(accountId string) (*payloads.GetProjectsResponse, error)
}

type ContractRoutes interface {
	GetContracts(projectSlug string) (*payloads.GetContractsResponse, error)
	UploadContracts(request payloads.UploadContractsRequest, projectSlug string) (*payloads.UploadContractsResponse, error)
	VerifyContracts(request payloads.UploadContractsRequest) (*payloads.UploadContractsResponse, error)
	RemoveContracts(request payloads.RemoveContractsRequest, projectSlug string) (*payloads.RemoveContractsResponse, error)
}

type ExportRoutes interface {
	ExportTransaction(request payloads.ExportTransactionRequest, projectSlug string) (*payloads.ExportTransactionResponse, error)
}

type NetworkRoutes interface {
	GetPublicNetworks() (*payloads.NetworksResponse, error)
}

type ActionRoutes interface {
	Validate(request generatedActions.ValidateRequest, projectSlug string) (*generatedActions.ValidateResponse, error)
	Publish(request generatedActions.PublishRequest, projectSlug string) (*generatedActions.PublishResponse, error)
}

type Rest struct {
	Auth     AuthRoutes
	User     UserRoutes
	Project  ProjectRoutes
	Contract ContractRoutes
	Export   ExportRoutes
	Networks NetworkRoutes
	Actions  ActionRoutes
}

func NewRest(
	auth AuthRoutes,
	user UserRoutes,
	project ProjectRoutes,
	contract ContractRoutes,
	export ExportRoutes,
	networks NetworkRoutes,
	actions ActionRoutes,
) *Rest {
	return &Rest{
		Auth:     auth,
		User:     user,
		Project:  project,
		Contract: contract,
		Export:   export,
		Networks: networks,
		Actions:  actions,
	}
}

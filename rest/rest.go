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
	RenameContract(request payloads.RenameContractRequest, projectSlug, networkID, address string) (*payloads.RenameContractResponse, error)
}

type ExportRoutes interface {
	ExportTransaction(request payloads.ExportTransactionRequest, projectSlug string) (*payloads.ExportTransactionResponse, error)
}

type NetworkRoutes interface {
	GetPublicNetworks() (*payloads.NetworksResponse, error)
}

type ActionRoutes interface {
	GetActions(accountSlugOrID string, projectSlugOrID string) (*payloads.GetActionsResponse, error)
	Validate(request generatedActions.ValidateRequest, projectSlug string) (*generatedActions.ValidateResponse, error)
	Publish(request generatedActions.PublishRequest, projectSlug string) (*generatedActions.PublishResponse, error)
}

type DevNetRoutes interface {
	SpawnRPC(accountID string, projectID string, templateSlug string, accessKey string, token string) (string, error)
}

type ExtensionRoutes interface {
	DeployExtension(accountSlugOrID string, projectSlugOrID string, actionID string, gatewayID string, extensionName string, extensionMethodName string) (*payloads.DeployExtensionResponse, error)
}

type GatewayRoutes interface {
	GetGateways(accountID string, projectID string) (*payloads.GetGatewaysResponse, error)
}

type Rest struct {
	Auth       AuthRoutes
	User       UserRoutes
	Project    ProjectRoutes
	Contract   ContractRoutes
	Export     ExportRoutes
	Networks   NetworkRoutes
	Actions    ActionRoutes
	DevNet     DevNetRoutes
	Gateways   GatewayRoutes
	Extensions ExtensionRoutes
}

func NewRest(
	auth AuthRoutes,
	user UserRoutes,
	project ProjectRoutes,
	contract ContractRoutes,
	export ExportRoutes,
	networks NetworkRoutes,
	actions ActionRoutes,
	devnet DevNetRoutes,
	gateways GatewayRoutes,
	extensions ExtensionRoutes,
) *Rest {
	return &Rest{
		Auth:       auth,
		User:       user,
		Project:    project,
		Contract:   contract,
		Export:     export,
		Networks:   networks,
		Actions:    actions,
		DevNet:     devnet,
		Gateways:   gateways,
		Extensions: extensions,
	}
}

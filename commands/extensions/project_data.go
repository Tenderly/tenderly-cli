package extensions

import (
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	extensionsModel "github.com/tenderly/tenderly-cli/model/extensions"
	gatewaysModel "github.com/tenderly/tenderly-cli/model/gateways"
)

var _ ProjectData = (*projectData)(nil)

type ProjectData interface {
	GetProjectSlug() string
	GetAccountSlug() string
	GetGateway() *gatewaysModel.Gateway
	GetActions() []actionsModel.Action
	GetExtensions() []extensionsModel.BackendExtension
	FindActionByName(name string) *actionsModel.Action
	FindActionByID(id string) *actionsModel.Action
	FindExtensionByName(name string) *extensionsModel.BackendExtension
}

type projectData struct {
	accountSlug string
	projectSlug string
	gateway     *gatewaysModel.Gateway
	actions     []actionsModel.Action
	extensions  []extensionsModel.BackendExtension
}

func NewProjectData(
	accountSlug string,
	projectSlug string,
	gateway *gatewaysModel.Gateway,
	actions []actionsModel.Action,
	extensions []extensionsModel.BackendExtension,
) ProjectData {
	if actions == nil {
		actions = []actionsModel.Action{}
	}
	if extensions == nil {
		extensions = []extensionsModel.BackendExtension{}
	}
	pd := &projectData{
		accountSlug: accountSlug,
		projectSlug: projectSlug,
		gateway:     gateway,
		actions:     actions,
		extensions:  extensions,
	}

	return pd
}

func (pd *projectData) GetProjectSlug() string {
	return pd.projectSlug
}

func (pd *projectData) GetAccountSlug() string {
	return pd.accountSlug
}

func (pd *projectData) GetGateway() *gatewaysModel.Gateway {
	return pd.gateway
}

func (pd *projectData) GetActions() []actionsModel.Action {
	return pd.actions
}

func (pd *projectData) GetExtensions() []extensionsModel.BackendExtension {
	return pd.extensions
}

func (pd *projectData) FindActionByName(name string) *actionsModel.Action {
	for _, action := range pd.actions {
		if action.Name == name {
			return &action
		}
	}

	return nil
}

func (pd *projectData) FindActionByID(ID string) *actionsModel.Action {
	for _, action := range pd.actions {
		if action.ID == ID {
			return &action
		}
	}

	return nil
}

func (pd *projectData) FindExtensionByName(name string) *extensionsModel.BackendExtension {
	for _, extension := range pd.extensions {
		if extension.Name == name {
			return &extension
		}
	}

	return nil
}

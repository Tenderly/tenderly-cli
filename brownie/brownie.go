package brownie

import (
	"github.com/tenderly/tenderly-cli/providers"
)

type DeploymentProvider struct {
}

func NewDeploymentProvider() *DeploymentProvider {
	return &DeploymentProvider{}
}

var _ providers.DeploymentProvider = (*DeploymentProvider)(nil)

func (*DeploymentProvider) GetProviderName() providers.DeploymentProviderName {
	return providers.BrownieDeploymentProvider
}

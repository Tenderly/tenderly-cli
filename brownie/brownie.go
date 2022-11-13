package brownie

import (
	"github.com/tenderly/tenderly-cli/providers"
)

var directoryStructure = []string{
	"build",
}

type Provider struct {
}

func NewBrownieProvider() Provider {
	return Provider{}
}

func (p Provider) GetProviderName() providers.DeploymentProviderName {
	return providers.BrownieDeploymentProvider
}

func (p Provider) GetDirectoryStructure() []string {
	return directoryStructure
}

package hardhat

import (
	"strconv"
	"strings"

	"github.com/tenderly/tenderly-cli/providers"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/call"
)

type DeploymentProvider struct {
	NetworkIdMap map[string]int
}

func NewDeploymentProvider() *DeploymentProvider {
	rest := rest.NewRest(
		call.NewAuthCalls(),
		call.NewUserCalls(),
		call.NewProjectCalls(),
		call.NewContractCalls(),
		call.NewExportCalls(),
		call.NewNetworkCalls(),
		call.NewActionCalls(),
	)

	networks, err := rest.Networks.GetPublicNetworks()

	if err != nil {
		return nil
	}

	idMap := make(map[string]int)

	for _, v := range *networks {
		val, err := strconv.Atoi(v.ID)
		if err != nil {
			continue
		}
		idMap[strings.ToLower(v.Name)] = val
	}

	return &DeploymentProvider{
		NetworkIdMap: idMap,
	}
}

var _ providers.DeploymentProvider = (*DeploymentProvider)(nil)

func (*DeploymentProvider) GetProviderName() providers.DeploymentProviderName {
	return providers.HardhatDeploymentProvider
}

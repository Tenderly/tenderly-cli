package brownie

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/providers"
)

const (
	contractDirectoryPath  = "contracts"
	contractDeploymentPath = "deployments"
	contractMapFile        = "map.json"

	dependencySeparator = "packages"
)

type networkIDFilter map[string]bool

// newNetworkIDFilter creates a new network ID filter map
func newNetworkIDFilter(networkIDs []string) networkIDFilter {
	networkIDFilterMap := make(map[string]bool)

	for _, networkID := range networkIDs {
		networkIDFilterMap[networkID] = true
	}

	return networkIDFilterMap
}

// hasNetworkID checks if the network ID filter map has a network ID
func (n networkIDFilter) hasNetworkID(networkID string) bool {
	_, exists := n[networkID]

	return exists
}

type contractsMap map[string]*providers.Contract

// mergeContractsMaps returns the merged contracts maps
func mergeContractsMaps(maps ...contractsMap) contractsMap {
	mergedMap := make(contractsMap)

	// Iterate over each map, and merge the keys.
	// This should be substituted with contractsMap.Copy(dst, src) when
	// the go version is bumped to at least 1.18
	for _, contractMap := range maps {
		for k, v := range contractMap {
			mergedMap[k] = v
		}
	}

	return mergedMap
}

// getContracts converts the contracts map to an array
func (c contractsMap) getContracts() []providers.Contract {
	var (
		index     = 0
		contracts = make([]providers.Contract, len(c))
	)

	for _, contract := range c {
		contracts[index] = *contract
		index++
	}

	return contracts
}

// filterContracts filters the contracts map based on the deployment map and network ID filters
func (c contractsMap) filterContracts(
	deploymentMap deploymentsMap,
	networkIDFilterMap networkIDFilter,
) int {
	var numberOfContractsWithANetwork int

	for networkID, contractDeployments := range deploymentMap {
		for contractName, deploymentAddresses := range contractDeployments {
			if _, ok := c[contractName]; !ok {
				continue
			}

			if !networkIDFilterMap.hasNetworkID(networkID) {
				continue
			}

			if c[contractName].Networks == nil {
				c[contractName].Networks = make(map[string]providers.ContractNetwork)
			}

			// We only take the latest deployment to some network
			c[contractName].Networks[networkID] = providers.ContractNetwork{
				Address: deploymentAddresses[0],
			}

			numberOfContractsWithANetwork += 1
		}
	}

	return numberOfContractsWithANetwork
}

type deploymentsMap map[string]map[string][]string

func (p Provider) GetContracts(
	buildDir string,
	networkIDs []string,
	_ ...*model.StateObject,
) ([]providers.Contract, int, error) {
	// Create the filter ID map
	networkIDFilterMap := newNetworkIDFilter(networkIDs)

	// Get the contracts directory listing
	contractMap, err := getContractsMap(filepath.Join(buildDir, contractDirectoryPath))
	if err != nil {
		return nil, 0, err
	}

	// Get the deployments map
	deploymentMap, err := readDeploymentsMap(
		filepath.Join(buildDir, contractDeploymentPath, contractMapFile),
	)
	if err != nil {
		return nil, 0, err
	}

	// Filter the contracts map
	numberOfContractsWithANetwork := contractMap.filterContracts(deploymentMap, networkIDFilterMap)

	return contractMap.getContracts(), numberOfContractsWithANetwork, nil
}

// getContractsMap recursively gathers contract files from the specified directory
// and aggregates them into a contracts map
func getContractsMap(contractsPath string) (contractsMap, error) {
	// Read the directory listing
	directoryEntry, err := os.ReadDir(contractsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get directory build files at %s, %w", contractsPath, err)
	}

	contractMap := make(contractsMap)

	for _, contractFile := range directoryEntry {
		fileName := contractFile.Name()

		// Check if there is an underlying contract file directory
		if contractFile.IsDir() {
			dependencyPath := filepath.Join(contractsPath, fileName)

			// Recursively fetch the contract files
			newMap, err := getContractsMap(dependencyPath)
			if err != nil {
				logrus.Warn(fmt.Sprintf("Failed resolving dependencies at %s with error: %s", dependencyPath, err))

				break
			}

			// Merge the maps
			contractMap = mergeContractsMaps(contractMap, newMap)

			continue
		}

		// The directory entry is a file, verify and read it
		if !strings.HasSuffix(fileName, ".json") {
			// Non-JSON files are ignored
			continue
		}

		// Read the contract data
		contractData, err := readProviderContract(filepath.Join(contractsPath, fileName))
		if err != nil {
			logrus.Warn("unable to read contract file, %v", err)

			break
		}

		// Set the source path
		sourcePath := strings.Split(contractData.SourcePath, dependencySeparator)
		if len(sourcePath) > 1 {
			contractData.SourcePath = strings.TrimPrefix(sourcePath[1], string(os.PathSeparator))
		}

		contractMap[contractData.Name] = contractData
	}

	return contractMap, nil
}

// readProviderContract reads the provider contract file from the specified path
func readProviderContract(contractFilePath string) (*providers.Contract, error) {
	contractRaw, err := os.ReadFile(contractFilePath)

	if err != nil {
		return nil, fmt.Errorf("unable to read contract file at %s, %w", contractFilePath, err)
	}

	var contractData providers.Contract

	if err := json.Unmarshal(contractRaw, &contractData); err != nil {
		return nil, fmt.Errorf("unable to parse contract file at %s, %w", contractFilePath, err)
	}

	return &contractData, err
}

// readDeploymentsMap reads the deployments map from the specified location
func readDeploymentsMap(deploymentsPath string) (deploymentsMap, error) {
	data, err := os.ReadFile(deploymentsPath)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to read deployment map file at %s, %w",
			deploymentsPath,
			err,
		)
	}

	var deploymentMap deploymentsMap

	if err := json.Unmarshal(data, &deploymentMap); err != nil {
		return nil, fmt.Errorf("unable to parse deployments map file, %w", err)
	}

	return deploymentMap, nil
}

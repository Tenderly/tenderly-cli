package brownie

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/providers"
)

const (
	BrownieContractDirectoryPath  = "contracts"
	BrownieContractDeploymentPath = "deployments"
	BrownieContractMapFile        = "map.json"

	BrownieDependencySeparator = "packages"
)

func (dp *DeploymentProvider) GetContracts(
	buildDir string,
	networkIDs []string,
	objects ...*model.StateObject,
) ([]providers.Contract, int, error) {
	contractsPath := filepath.Join(buildDir, BrownieContractDirectoryPath)
	files, err := os.ReadDir(contractsPath)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed listing build files")
	}

	networkIDFilterMap := make(map[string]bool)
	for _, networkID := range networkIDs {
		networkIDFilterMap[networkID] = true
	}
	objectMap := make(map[string]*model.StateObject)
	for _, object := range objects {
		if object.Code == nil || len(object.Code) == 0 {
			continue
		}
		objectMap[hexutil.Encode(object.Code)] = object
	}

	contractMap := make(map[string]*providers.Contract)
	var numberOfContractsWithANetwork int
	for _, contractFile := range files {
		if contractFile.IsDir() {
			dependencyPath := filepath.Join(contractsPath, contractFile.Name())
			err = dp.resolveDependencies(dependencyPath, contractMap)
			if err != nil {
				logrus.Debug(fmt.Sprintf("Failed resolving dependencies at %s with error: %s", dependencyPath, err))
				break
			}
			continue
		}
		if !strings.HasSuffix(contractFile.Name(), ".json") {
			continue
		}
		contractFilePath := filepath.Join(contractsPath, contractFile.Name())
		data, err := os.ReadFile(contractFilePath)
		if err != nil {
			logrus.Debug(fmt.Sprintf("Failed reading build file at %s with error: %s", contractFilePath, err))
			break
		}

		var contractData providers.Contract
		err = json.Unmarshal(data, &contractData)
		if err != nil {
			logrus.Debug(fmt.Sprintf("Failed parsing build file at %s with error: %s", contractFilePath, err))
			break
		}

		contractMap[contractData.Name] = &contractData
	}

	deploymentMapFile := filepath.Join(buildDir, BrownieContractDeploymentPath, BrownieContractMapFile)

	data, err := os.ReadFile(deploymentMapFile)
	if err != nil {
		logrus.Debug(fmt.Sprintf("Failed reading map file at %s with error: %s", deploymentMapFile, err))
		return nil, 0, errors.Wrap(err, "failed reading map file")
	}

	var deploymentMap map[string]map[string][]string
	err = json.Unmarshal(data, &deploymentMap)
	if err != nil {
		logrus.Debug(fmt.Sprintf("Failed parsing map file at %s with error: %s", deploymentMapFile, err))
		return nil, 0, errors.Wrap(err, "failed unmarshaling map file")
	}
	for networkID, contractDeployments := range deploymentMap {
		for contractName, deploymentAddresses := range contractDeployments {
			if _, ok := contractMap[contractName]; !ok {
				continue
			}

			if len(networkIDFilterMap) > 0 && !networkIDFilterMap[networkID] {
				continue
			}

			if contractMap[contractName].Networks == nil {
				contractMap[contractName].Networks = make(map[string]providers.ContractNetwork)
			}
			//We only take the latest deployment to some network
			contractMap[contractName].Networks[networkID] = providers.ContractNetwork{
				Address: deploymentAddresses[0],
			}
			numberOfContractsWithANetwork += 1
		}
	}

	var contracts []providers.Contract
	for _, contract := range contractMap {
		contracts = append(contracts, *contract)
	}

	return contracts, numberOfContractsWithANetwork, nil
}

func (dp *DeploymentProvider) resolveDependencies(path string, contractMap map[string]*providers.Contract) error {
	info, err := os.Stat(path)
	if err != nil {
		logrus.Debugf("Failed reading dependency at %s", path)
		return errors.Wrap(err, "failed reading dependency files")
	}
	if info.IsDir() {
		files, err := os.ReadDir(path)
		if err != nil {
			logrus.Debugf("Failed reading dependency at %s", path)
			return errors.Wrap(err, "failed reading dependency files")
		}

		for _, file := range files {
			newFilePath := filepath.Join(path, file.Name())
			err = dp.resolveDependencies(newFilePath, contractMap)
			if err != nil {
				return err
			}
		}
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logrus.Debug(fmt.Sprintf("Failed reading build file at %s with error: %s", path, err))
		return errors.Wrap(err, "failed reading contract")
	}

	var contractData providers.Contract
	err = json.Unmarshal(data, &contractData)
	if err != nil {
		logrus.Debug(fmt.Sprintf("Failed parsing build file at %s with error: %s", path, err))
		return errors.Wrap(err, "failed parsing contract")
	}

	sourcePath := strings.Split(contractData.SourcePath, BrownieDependencySeparator)
	contractData.SourcePath = strings.TrimPrefix(sourcePath[1], string(os.PathSeparator))
	contractMap[contractData.Name] = &contractData

	return nil
}

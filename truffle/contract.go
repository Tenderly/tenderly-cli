package truffle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/providers"
)

func (dp *DeploymentProvider) GetContracts(
	buildDir string,
	networkIDs []string,
	objects ...*model.StateObject,
) ([]providers.Contract, int, error) {
	files, err := os.ReadDir(buildDir)
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

	hasNetworkFilters := len(networkIDFilterMap) > 0

	sources := make(map[string]bool)
	var contracts []providers.Contract
	var numberOfContractsWithANetwork int
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(buildDir, file.Name())
		data, err := os.ReadFile(filePath)

		if err != nil {
			return nil, 0, errors.Wrap(err, "failed reading build file")
		}

		var contract providers.Contract
		err = json.Unmarshal(data, &contract)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed parsing build file")
		}

		if contract.Networks == nil {
			contract.Networks = make(map[string]providers.ContractNetwork)
		}

		sourcePath := contract.SourcePath
		if runtime.GOOS == "windows" && strings.HasPrefix(sourcePath, "/") {
			sourcePath = strings.ReplaceAll(sourcePath, "/", "\\")
			sourcePath = strings.TrimPrefix(sourcePath, "\\")
			sourcePath = strings.Replace(sourcePath, "\\", ":\\", 1)
		}
		sources[sourcePath] = true

		for _, node := range contract.Ast.Nodes {
			if node.NodeType != "ImportDirective" {
				continue
			}

			absPath := node.AbsolutePath

			if runtime.GOOS == "windows" && strings.HasPrefix(absPath, "/") {
				absPath = strings.ReplaceAll(absPath, "/", "\\")
				absPath = strings.TrimPrefix(absPath, "\\")
				absPath = strings.Replace(absPath, "\\", ":\\", 1)
			}

			if !sources[absPath] && node.AbsolutePath == node.File {
				sources[absPath] = false
			}
		}

		if object := objectMap[contract.DeployedBytecode]; object != nil && len(networkIDs) == 1 {
			contract.Networks[networkIDs[0]] = providers.ContractNetwork{
				Links:   nil, // @TODO: Libraries
				Address: object.Address,
			}
		}

		if hasNetworkFilters {
			for networkID := range contract.Networks {
				if !networkIDFilterMap[networkID] {
					delete(contract.Networks, networkID)
				}
			}
		}

		contracts = append(contracts, contract)
		numberOfContractsWithANetwork += len(contract.Networks)
	}

	for localPath, included := range sources {
		if !included {
			currentLocalPath := localPath
			if len(localPath) > 0 && localPath[0] == '@' {
				localPath, err = os.Getwd()
				if err != nil {
					return nil, 0, errors.Wrap(err, "failed getting working dir")
				}

				if runtime.GOOS == "windows" {
					currentLocalPath = strings.ReplaceAll(currentLocalPath, "/", "\\")
					currentLocalPath = strings.TrimPrefix(currentLocalPath, "\\")
				}

				localPath = filepath.Join(localPath, "node_modules", currentLocalPath)
				doesNotExist := providers.CheckIfFileDoesNotExist(localPath)
				if doesNotExist {
					localPath = providers.GetGlobalPathForModule(currentLocalPath)
				}
			}

			source, err := os.ReadFile(localPath)
			if err != nil {
				return nil, 0, errors.Wrap(err, "failed reading contract source file")
			}

			contracts = append(contracts, providers.Contract{
				Source:     string(source),
				SourcePath: currentLocalPath,
			})
		}
	}

	return contracts, numberOfContractsWithANetwork, nil
}

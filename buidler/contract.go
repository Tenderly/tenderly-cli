package buidler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/providers"
)

type BuidlerContract struct {
	*providers.Contract
	Address  string         `json:"address"`
	Receipt  buidlerReceipt `json:"receipt"`
	Metadata string         `json:"metadata"`
}

type buidlerReceipt struct {
	TransactionHash string `json:"transactionHash"`
}

type buidlerMetadata struct {
	Compiler providers.ContractCompiler           `json:"compiler"`
	Sources  map[string]providers.ContractSources `json:"sources"`
}

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
		if !file.IsDir() {
			continue
		}
		filePath := filepath.Join(buildDir, file.Name())
		contractFiles, err := os.ReadDir(filePath)

		for _, contractFile := range contractFiles {
			if contractFile.IsDir() || !strings.HasSuffix(contractFile.Name(), ".json") {
				continue
			}
			contractFilePath := filepath.Join(filePath, contractFile.Name())
			data, err := os.ReadFile(contractFilePath)

			if err != nil {
				return nil, 0, errors.Wrap(err, "failed reading build file")
			}

			var buidlerContract BuidlerContract
			var buidlerMeta buidlerMetadata
			err = json.Unmarshal(data, &buidlerContract)
			if err != nil {
				return nil, 0, errors.Wrap(err, "failed parsing build file")
			}

			err = json.Unmarshal([]byte(buidlerContract.Metadata), &buidlerMeta)
			if err != nil {
				return nil, 0, errors.Wrap(err, "failed parsing build file")
			}

			contract := providers.Contract{
				Abi:              buidlerContract.Abi,
				Bytecode:         buidlerContract.Bytecode,
				DeployedBytecode: buidlerContract.DeployedBytecode,
				SourcePath:       filePath,
				Compiler: providers.ContractCompiler{
					Name:    "",
					Version: buidlerMeta.Compiler.Version,
				},
			}

			contract.Name = fmt.Sprintf("%s", strings.Split(contractFile.Name(), ".")[0])

			networkData := strings.Split(file.Name(), "_")

			if contract.Networks == nil {
				contract.Networks = make(map[string]providers.ContractNetwork)
			}

			if len(networkData) == 1 {
				if val, ok := dp.NetworkIdMap[networkData[0]]; ok {
					contract.Networks[strconv.Itoa(val)] = providers.ContractNetwork{
						Address:         buidlerContract.Address,
						TransactionHash: buidlerContract.Receipt.TransactionHash,
					}
				}
			}

			if len(networkData) == 2 {
				contract.Networks[networkData[1]] = providers.ContractNetwork{
					Address:         buidlerContract.Address,
					TransactionHash: buidlerContract.Receipt.TransactionHash,
				}
			}

			contract.SourcePath = fmt.Sprintf("%s/%s.sol", contract.SourcePath, contract.Name)
			sourcePath := contract.SourcePath
			if runtime.GOOS == "windows" && strings.HasPrefix(sourcePath, "/") {
				sourcePath = strings.ReplaceAll(sourcePath, "/", "\\")
				sourcePath = strings.TrimPrefix(sourcePath, "\\")
				sourcePath = strings.Replace(sourcePath, "\\", ":\\", 1)
			}
			sources[sourcePath] = true

			for path, _ := range buidlerMeta.Sources {
				if !strings.Contains(path, "@") {
					continue
				}
				absPath := path

				if runtime.GOOS == "windows" && strings.HasPrefix(absPath, "/") {
					absPath = strings.ReplaceAll(absPath, "/", "\\")
					absPath = strings.TrimPrefix(absPath, "\\")
					absPath = strings.Replace(absPath, "\\", ":\\", 1)
				}

				if !sources[absPath] {
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

			sourceKey := fmt.Sprintf("contracts/%s.sol", contract.Name)

			if val, ok := buidlerMeta.Sources[sourceKey]; ok {
				contract.Source = val.Content
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
	}

	return contracts, numberOfContractsWithANetwork, nil
}

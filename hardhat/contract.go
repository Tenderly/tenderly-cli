package hardhat

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
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/providers"
)

type HardhatContract struct {
	*providers.Contract
	Address  string         `json:"address"`
	Receipt  hardhatReceipt `json:"receipt"`
	Metadata string         `json:"metadata"`
}

type hardhatReceipt struct {
	TransactionHash string `json:"transactionHash"`
}

type hardhatMetadata struct {
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
		contractFiles, _ := os.ReadDir(filePath)

		successfulRead := true
		for _, contractFile := range contractFiles {
			if contractFile.IsDir() || !strings.HasSuffix(contractFile.Name(), ".json") {
				continue
			}
			contractFilePath := filepath.Join(filePath, contractFile.Name())
			data, err := os.ReadFile(contractFilePath)

			if err != nil {
				logrus.Debug(fmt.Sprintf("Failed reading build file at %s with error: %s", contractFilePath, err))
				successfulRead = false
				break
			}

			var hardhatContract HardhatContract
			var hardhatMeta hardhatMetadata
			err = json.Unmarshal(data, &hardhatContract)
			if err != nil {
				logrus.Debug(fmt.Sprintf("Failed parsing build file at %s with error: %s", contractFilePath, err))
				successfulRead = false
				break
			}

			if hardhatContract.Metadata != "" {
				err = json.Unmarshal([]byte(hardhatContract.Metadata), &hardhatMeta)
				if err != nil {
					logrus.Debug(fmt.Sprintf("Failed parsing build file metadata at %s with error: %s", contractFilePath, err))
					successfulRead = false
					break
				}
			}

			contract := providers.Contract{
				Abi:              hardhatContract.Abi,
				Bytecode:         hardhatContract.Bytecode,
				DeployedBytecode: hardhatContract.DeployedBytecode,
				SourcePath:       filePath,
				Compiler: providers.ContractCompiler{
					Name:    "",
					Version: hardhatMeta.Compiler.Version,
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
						Address:         hardhatContract.Address,
						TransactionHash: hardhatContract.Receipt.TransactionHash,
					}
				} else {
					chainIdPath := filepath.Join(filePath, ".chainId")

					chainData, err := os.ReadFile(chainIdPath)
					if err != nil {
						logrus.Debug(fmt.Sprintf("Failed reading chainID file at %s with error: %s", chainIdPath, err))
						successfulRead = false
						break
					}

					var chainId int
					err = json.Unmarshal(chainData, &chainId)
					if err != nil {
						logrus.Debug(fmt.Sprintf("Failed parsing chainID file at %s with error: %s", chainIdPath, err))
						successfulRead = false
						break
					}

					contract.Networks[strconv.Itoa(chainId)] = providers.ContractNetwork{
						Address:         hardhatContract.Address,
						TransactionHash: hardhatContract.Receipt.TransactionHash,
					}
				}
			}

			if len(networkData) == 2 {
				contract.Networks[networkData[1]] = providers.ContractNetwork{
					Address:         hardhatContract.Address,
					TransactionHash: hardhatContract.Receipt.TransactionHash,
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

			for path, _ := range hardhatMeta.Sources {
				if strings.Contains(path, contract.Name) {
					contract.SourcePath = path
					continue
				}
				if !strings.Contains(path, "@") {
					sources[path] = false
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

			if val, ok := hardhatMeta.Sources[contract.SourcePath]; ok {
				contract.Source = val.Content
			}

			contracts = append(contracts, contract)
			numberOfContractsWithANetwork += len(contract.Networks)
		}

		if !successfulRead {
			continue
		}

		for localPath, included := range sources {
			if !included {
				currentLocalPath := localPath
				if len(localPath) > 0 && localPath[0] == '@' {
					localPath, err = os.Getwd()
					if err != nil {
						logrus.Debug(fmt.Sprintf("Failed getting working dir at %s with error: %s", currentLocalPath, err))
						successfulRead = false
						continue
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
					localPath = filepath.Join("node_modules", currentLocalPath)
					doesNotExist := providers.CheckIfFileDoesNotExist(localPath)
					if doesNotExist {
						localPath = providers.GetGlobalPathForModule(currentLocalPath)
					}

					source, err := os.ReadFile(localPath)
					if err != nil {
						logrus.Debug(fmt.Sprintf("Failed reading contract source file at %s with error: %s", localPath, err))
						successfulRead = false
						continue
					}

					contracts = append(contracts, providers.Contract{
						Source:     string(source),
						SourcePath: currentLocalPath,
					})

					continue
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

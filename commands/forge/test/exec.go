package test

import (
	"bufio"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"os"
	osexec "os/exec"
	"strings"
)

type exec struct{}

func (e *exec) forgeCompile() error {
	cmd := osexec.Command("forge", "compile")
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (e *exec) forgeVerify(address common.Address, contractName string, rpcURL string, accessKey string) error {
	// hack to enable it for canary
	baseURL := rpcURL
	if strings.Contains(rpcURL, "mainnet/") {
		// remove it from rpcURL
		start := strings.Index(rpcURL, "mainnet/")
		end := start + len("mainnet/")
		baseURL = rpcURL[:start] + rpcURL[end:]
	}

	verifierURL := baseURL + "/verify/etherscan"

	cmd := osexec.Command("forge", "verify-contract", address.Hex(), contractName, "--etherscan-api-key", accessKey, "--verifier-url", verifierURL)
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

func (e *exec) findFiles(folder string) ([]string, error) {
	cmd := osexec.Command("find", folder, "-type", "f")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	filesStr := strings.Trim(string(output), "\n")
	files := strings.Split(filesStr, "\n")

	return files, nil
}

func (e *exec) getContractTestName(filePath string) (string, bool) {
	// read file content
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// don't forget multiline contract definitions, fuzztest, commas, etc
		if strings.Contains(line, "contract") && strings.Contains(line, "{") && strings.Contains(line, " Test") {
			line = strings.Trim(line, " ")
			tokens := strings.Split(line, " ")
			// contract name is second token // contract MyTest is Test {
			return tokens[1], true
		}
	}

	return "", false
}

func (e *exec) getCompilerOutput(filePath string) (*CompilerOutput, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var output CompilerOutput
	err = json.Unmarshal(bytes, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

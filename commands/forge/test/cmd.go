package test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/commands/forge"
	"os"
	osexec "os/exec"
	"strings"
)

func init() {
	forge.Cmd.AddCommand(newCommand())
}

func newCommand() *cobra.Command {
	cmdTest := newCmdTest()

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test Commands",
		Run:   cmdTest.Run,
	}

	cmdTest.args = newArgs(cmd)

	return cmd
}

type cmdTest struct {
	args *args
	exec *exec
}

func newCmdTest() *cmdTest {
	return &cmdTest{
		exec: &exec{},
	}
}

func (c *cmdTest) Run(cmd *cobra.Command, args []string) {
	cmd.Println()

	wd, err := os.Getwd()
	if err != nil {
		cmd.PrintErrln(err)
		return
	}
	if len(args) > 0 {
		wd = args[0]
	}

	testFolder := wd + "/test"
	outFolder := wd + "/out"

	cmd.Println("Compiling...")
	err = c.exec.forgeCompile()
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	testFiles, err := c.exec.findFiles(testFolder)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	// right now we know that there is only single file, there should be a loop
	testFilePath := testFiles[0]
	contractName, ok := c.exec.getContractTestName(testFilePath)
	if !ok {
		cmd.PrintErrln("Test Contract Name not found")
		return
	}

	testFileName := fileName(testFilePath)
	compilationOutputFile := outFolder + "/" + testFileName + "/" + contractName + ".json"

	compilerOutput, err := c.exec.getCompilerOutput(compilationOutputFile)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	cmd.Println("Connecting to Virtual Testnet... ", c.args.rpcURL)
	rpcClient, err := newRpcClient(c.args.rpcURL)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	// set nonce (commented because of multiple test runs)
	//err = rpcClient.SetNonce(sender, hexutil.Uint64(1))
	//if err != nil {
	//	cmd.PrintErrln(err)
	//	return
	//}

	// set balance sender
	cmd.Println("Setting balances... ")
	err = rpcClient.SetBalance(sender, hexutil.Big(*abi.MaxUint256))
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	// set balance caller
	err = rpcClient.SetBalance(caller, hexutil.Big(*abi.MaxUint256))
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	// deploy contract
	cmd.Println(fmt.Sprintf("Deploying contract %s ...", contractName))
	bytecode := hexutil.MustDecode(compilerOutput.Bytecode.Object)
	ca := NewCallArgs(
		from(sender),
		data(bytecode),
	)
	txHash, err := rpcClient.SendTransaction(ca)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	receipt, err := rpcClient.GetTransactionReceipt(txHash)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	testAddressStr, ok := receipt["contractAddress"].(string)
	if !ok || testAddressStr == "" {
		cmd.PrintErrln("Contract Address not found", ok, testAddressStr)
		return
	}
	testAddress := common.HexToAddress(testAddressStr)

	// verify contract if api key
	if c.args.tdlyKey != "" {
		cmd.Println("Verifying contract...")
		err = c.exec.forgeVerify(testAddress, contractName, c.args.rpcURL, c.args.tdlyKey)
		if err != nil {
			cmd.Println("Error occurred during verification.")
			execErr, ok := err.(*osexec.ExitError)
			if ok {
				cmd.Println(string(execErr.Stderr))
			}
		}
	} else {
		cmd.Println("Skipping verification, no API key provided")
	}

	// call setUp function
	cmd.Println(fmt.Sprintf("Execute %s.setUp()", contractName))
	_, err = rpcClient.Execute(testAddress, "setUp()")
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	testFns := testFunctions(compilerOutput)
	testFn := testFns[0] // assume that there is only a single test function

	// call test function
	cmd.Println(fmt.Sprintf("Running test %s.%s", contractName, testFn))
	txHash, err = rpcClient.Execute(testAddress, testFn)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}
	cmd.Println()

	receipt, err = rpcClient.GetTransactionReceipt(txHash)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	status := receipt["status"]
	gasUsed := receipt["gasUsed"]
	cmd.Println(fmt.Sprintf("%s %s (gas: %d)", fmtStatus(status), testFn, fmtGas(gasUsed)))
}

func fileName(filePath string) string {
	tokens := strings.Split(filePath, "/")
	return tokens[len(tokens)-1]
}

func testFunctions(out *CompilerOutput) []string {
	functions := make([]string, 0)
	for _, abi := range out.Abi {
		if abi.Type == "function" && strings.Contains(abi.Name, "test_") { //check if this filtering is ok (test_)
			fnName := abi.Name + "()" // we don't have fuzzying right now, just add brackets
			functions = append(functions, fnName)
		}
	}

	return functions
}

func fmtStatus(hex any) string {
	hexStr, ok := hex.(string)
	if !ok {
		return commands.Colorizer.Yellow("[UNKNOWN]").String()
	}
	if hexStr == "0x0" {
		return commands.Colorizer.Red("[FAIL]").String()
	}
	return commands.Colorizer.Green("[PASS]").String()
}

func fmtGas(gasUsedHex any) uint64 {
	gasUsedHexStr, ok := gasUsedHex.(string)
	if !ok {
		return 0
	}
	return hexutil.MustDecodeBig(gasUsedHexStr).Uint64()
}

package commands

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/call"
	"github.com/tenderly/tenderly-cli/truffle"
	"github.com/tenderly/tenderly-cli/userError"
	"os"
	"path/filepath"
)

func newRest() *rest.Rest {
	return rest.NewRest(
		call.NewAuthCalls(),
		call.NewUserCalls(),
		call.NewProjectCalls(),
		call.NewContractCalls(),
	)
}

func MustGetTruffleConfig() (*truffle.Config, error) {
	projectDir, err := filepath.Abs(config.ProjectDirectory)
	truffleConfigFile := truffle.NewTruffleConfigFile

	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("get absolute project dir: %s", err),
			"Couldn't get absolute project path",
		)
	}

	truffleConfig, err := truffle.GetTruffleConfig(truffleConfigFile, projectDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Truffle config file",
		)
	}
	if os.IsNotExist(err) {
		logrus.Debugf("couldn't read new truffle config file: %s", err)
		truffleConfigFile = truffle.OldTruffleConfigFile
		truffleConfig, err = truffle.GetTruffleConfig(truffleConfigFile, projectDir)
	}

	if os.IsNotExist(err) {
		logrus.Debugf("couldn't read truffle config file: %s", err)
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't find Truffle config file",
		)
	}

	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("unable to fetch config: %s", err),
			"Couldn't read Truffle config file",
		)
	}

	return truffleConfig, nil
}

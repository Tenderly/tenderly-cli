package devnet

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/userError"
	"os"
)

var accountID string
var projectSlug string
var templateSlug string
var accessKey string
var token string

func init() {
	spawnRpcCommand.PersistentFlags().StringVar(
		&accountID,
		"account",
		"",
		"The Tenderly account. If not provided, the system will try to read 'account_id' from the 'tenderly.yaml' configuration file.",
	)
	spawnRpcCommand.PersistentFlags().StringVar(
		&projectSlug,
		"project",
		"",
		"The DevNet project slug. If not provided, the system will try to read 'project_slug' from the 'tenderly.yaml' configuration file.",
	)
	spawnRpcCommand.PersistentFlags().StringVar(
		&templateSlug,
		"template",
		"",
		"The DevNet template which is going to be applied when spawning the DevNet RPC.",
	)
	spawnRpcCommand.PersistentFlags().StringVar(
		&accessKey,
		"access_key",
		"",
		"The Tenderly access key. If not provided, the system will try to read 'access_key' from the 'tenderly.yaml' configuration file.",
	)
	spawnRpcCommand.PersistentFlags().StringVar(
		&token,
		"token",
		"",
		"The Tenderly token. If not provided, the system will try to read 'token' from the 'tenderly.yaml' configuration file.",
	)
	DevNetCmd.AddCommand(spawnRpcCommand)
}

var spawnRpcCommand = &cobra.Command{
	Use:   "spawn-rpc",
	Short: "Spawn DevNet RPC",
	Long:  `Spawn DevNet RPC that represents the network endpoint for your DevNet`,
	Run:   spawnRPCHandler,
}

func spawnRPCHandler(cmd *cobra.Command, args []string) {
	if accountID == "" {
		commands.CheckLogin()
		accountID = config.GetGlobalString(config.AccountID)
	}

	if accountID == "" {
		err := userError.NewUserError(errors.New("account not found"), "No account found. Please login with `tenderly login` or provide '--account' flag")
		userError.LogError(err)
		os.Exit(1)
	}

	if accessKey == "" {
		accessKey = config.GetGlobalString(config.AccessKey)
	}

	if token == "" {
		token = config.GetGlobalString(config.Token)
	}

	if accessKey == "" && token == "" {
		err := userError.NewUserError(errors.New("access key or token not found"), "No access key or token found. Please login with `tenderly login` or provide '--access_key' or '--token' flag")
		userError.LogError(err)
		os.Exit(1)
	}

	if projectSlug == "" {
		projectSlug = config.MaybeGetString(config.ProjectSlug)
	}

	if projectSlug == "" {
		err := userError.NewUserError(errors.New("project not found"), "No project found. Please set a project with `tenderly use project <project-slug>` or provide '--project' flag")
		userError.LogError(err)
		os.Exit(1)
	}

	if templateSlug == "" {
		err := userError.NewUserError(errors.New("Missing required argument 'template'"), "Missing required argument 'template'")
		userError.LogError(err)
		os.Exit(1)
	}

	rest := commands.NewRest()
	response, err := rest.DevNet.SpawnRPC(accountID, projectSlug, templateSlug, accessKey, token)
	if err != nil {
		logrus.Error("Failed to spawn RPC", err)
		return
	}
	logrus.Info(response)
}

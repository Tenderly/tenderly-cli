package devnet

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/userError"
)

var accountID string
var projectSlug string
var templateSlug string
var accessKey string
var token string
var returnUrl bool

func init() {
	cmdSpawnRpc.PersistentFlags().StringVar(
		&accountID,
		"account",
		"",
		"The Tenderly account username or organization slug. If not provided, the system will try to read 'account_id' from the 'tenderly.yaml' configuration file.",
	)
	cmdSpawnRpc.PersistentFlags().StringVar(
		&projectSlug,
		"project",
		"",
		"The DevNet project slug. If not provided, the system will try to read 'project_slug' from the 'tenderly.yaml' configuration file.",
	)
	cmdSpawnRpc.PersistentFlags().StringVar(
		&templateSlug,
		"template",
		"",
		"The DevNet template slug which is going to be applied when spawning the DevNet RPC.",
	)
	cmdSpawnRpc.PersistentFlags().StringVar(
		&accessKey,
		"access_key",
		"",
		"The Tenderly access key. If not provided, the system will try to read 'access_key' from the 'tenderly.yaml' configuration file.",
	)
	cmdSpawnRpc.PersistentFlags().StringVar(
		&token,
		"token",
		"",
		"The Tenderly JWT. If not provided, the system will try to read 'token' from the 'tenderly.yaml' configuration file.",
	)
	cmdSpawnRpc.PersistentFlags().BoolVar(
		&returnUrl,
		"return-url",
		false,
		"Optional flag to return the URL instead of printing it. Default: false.",
	)
	CmdDevNet.AddCommand(cmdSpawnRpc)
}

var cmdSpawnRpc = &cobra.Command{
	Use:   "spawn-rpc",
	Short: "Spawn DevNet RPC",
	Long:  `Spawn DevNet RPC that represents the JSON-RPC endpoint for your DevNet`,
	Run:   spawnRPCHandler,
}

func spawnRPCHandler(cmd *cobra.Command, args []string) {
	if accountID == "" {
		commands.CheckLogin()
		accountID = config.GetGlobalString(config.AccountID)
	}

	if accountID == "" {
		err := userError.NewUserError(
			errors.New("account not found"),
			"An account is required. Please log in using tenderly login or include the '--account' flag.",
		)
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
		err := userError.NewUserError(
			errors.New("access key or token not found"),
			"An access key or token is required. Please log in using tenderly login or include the '--access_key' or '--token' flag.",
		)
		userError.LogError(err)
		os.Exit(1)
	}

	if projectSlug == "" {
		projectSlug = config.MaybeGetString(config.ProjectSlug)
	}

	if projectSlug == "" {
		err := userError.NewUserError(
			errors.New("project not found"),
			"No project was found. To set a project, use tenderly use project <project-slug> or include the '--project' flag.",
		)
		userError.LogError(err)
		os.Exit(1)
	}

	if templateSlug == "" {
		err := userError.NewUserError(
			errors.New("Missing required argument 'template'"),
			"The 'template' argument is required. Please include the '--template' flag and provide the DevNet template slug.",
		)
		userError.LogError(err)
		os.Exit(1)
	}

	rest := commands.NewRest()
	response, err := rest.DevNet.SpawnRPC(accountID, projectSlug, templateSlug, accessKey, token)
	if err != nil {
		logrus.Error("Failed to spawn RPC", err)
		return
	}

	if returnUrl {
		fmt.Printf("%s\n", response)
	} else {
		logrus.Info(response)
	}
}

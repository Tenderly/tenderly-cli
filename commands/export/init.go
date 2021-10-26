package export

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/userError"
)

func init() {
	exportCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Export init is a helper subcommand for creating exported network configuration",
	Run: func(cmd *cobra.Command, args []string) {
		commands.CheckLogin()

		if exportNetwork == "" {
			exportNetwork = promptExportNetwork()
		}

		if config.IsNetworkConfigured(exportNetwork) && !reExport {
			logrus.Info(commands.Colorizer.Sprintf("The network %s is already configured. If you want to set up the network again, rerun this command with the %s flag.",
				commands.Colorizer.Bold(commands.Colorizer.Green(exportNetwork)),
				commands.Colorizer.Bold(commands.Colorizer.Green("--re-init")),
			))
			os.Exit(0)
		}

		if config.IsNetworkConfigured(exportNetwork) {
			network = GetNetwork(exportNetwork)
		} else {
			network = &config.ExportNetwork{}
		}

		rest := commands.NewRest()

		networks, err := rest.Networks.GetPublicNetworks()
		if err != nil {
			userError.LogErrorf("failed fetching public networks: %s",
				userError.NewUserError(
					err,
					"Fetching public networks failed. This can happen if you are running an older version of the Tenderly CLI.",
				),
			)

			commands.CheckVersion(true, true)

			os.Exit(1)
		}

		accountID := config.GetString(config.AccountID)

		projectsResponse, err := rest.Project.GetProjects(accountID)
		if err != nil {
			userError.LogErrorf("failed fetching projects: %s",
				userError.NewUserError(
					err,
					"Fetching projects for account failed. This can happen if you are running an older version of the Tenderly CLI.",
				),
			)

			commands.CheckVersion(true, true)

			os.Exit(1)
		}
		if projectsResponse.Error != nil {
			userError.LogErrorf("get projects call: %s", projectsResponse.Error)
			os.Exit(1)
		}

		project := commands.GetProjectFromFlag(exportProjectName, projectsResponse.Projects, rest)

		if project == nil {
			project = commands.PromptProjectSelect(projectsResponse.Projects, rest, true)
		}
		if project != nil {
			slug := project.Slug
			if project.OwnerInfo != nil {
				slug = fmt.Sprintf("%s/%s", project.OwnerInfo.Username, slug)
			}
			network.ProjectSlug = slug
		}

		if rpcAddress == "" {
			rpcAddress = promptRpcAddress()
		}
		if network.RpcAddress != rpcAddress {
			network.RpcAddress = rpcAddress
		}

		if forkedNetwork == "" {
			networkNames := []string{"None"}
			for _, network := range *networks {
				networkNames = append(networkNames, network.Name)
			}
			forkedNetwork = promptForkedNetwork(networkNames)
		}
		if network.ForkedNetwork != forkedNetwork {
			network.ForkedNetwork = forkedNetwork
		}

		err = config.WriteExportNetwork(exportNetwork, network)
		if err != nil {
			userError.LogErrorf(
				"write project config: %s",
				userError.NewUserError(err, "Couldn't write project config file"),
			)
			os.Exit(1)
		}
	},
}

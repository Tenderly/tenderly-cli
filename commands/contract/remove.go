package contract

import (
	"github.com/pkg/errors"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/userError"
)

var (
	contractTag string
	contractID  string
	projectSlug string
)

func init() {
	removeCmd.PersistentFlags().StringVar(&contractTag, "tag", "", "Remove all contracts with matched tag from configured project")
	removeCmd.PersistentFlags().StringVar(&contractID, "id", "", "Remove contract with \"id\"(\"eth:{network_id}:{contract_id}\").")
	removeCmd.PersistentFlags().StringVar(&projectSlug, "project-slug", "", "The slug of a project you wish to remove contracts")

	ContractsCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove contracts from configured project.",
	Run: func(cmd *cobra.Command, args []string) {
		rest := commands.NewRest()

		err := removeContracts(rest)
		if err != nil {
			userError.LogErrorf("unable to remove contracts: %s", err)
			os.Exit(1)
		}

		logrus.Infof("Successfully removed all selected smart contracts.")
	},
}

func removeContracts(rest *rest.Rest) error {
	projectConfigurations, err := commands.GetProjectConfiguration()
	if err != nil {
		return userError.NewUserError(
			errors.Wrap(err, "unable to get project configuration"),
			commands.Colorizer.Sprintf("Failed reading project configuration. For more info please rerun this command with the %s flag.",
				commands.Colorizer.Bold(commands.Colorizer.Green("--debug")),
			),
		)
	}

	if pushProjectSlug != "" {
		projectConfiguration, exists := projectConfigurations[pushProjectSlug]
		if !exists {
			return userError.NewUserError(
				errors.Wrap(err, "cannot find project configuration via slug"),
				commands.Colorizer.Sprintf("Failed reading project configuration. Couldn't find project with slug: %s",
					commands.Colorizer.Bold(commands.Colorizer.Red(pushProjectSlug)),
				),
			)
		}

		projectConfigurations = map[string]*commands.ProjectConfiguration{
			pushProjectSlug: projectConfiguration,
		}
	}

	return nil
}

package contract

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/userError"
)

var (
	removeContractTag string
	removeContractID  string
	removeProjectSlug string
)

func init() {
	removeCmd.PersistentFlags().StringVar(&removeContractTag, "tag", "", "Remove all contracts with matched tag from configured project")
	removeCmd.PersistentFlags().StringVar(&removeContractID, "id", "", "Remove contract with \"id\"(\"eth:{network_id}:{contract_id}\").")
	removeCmd.PersistentFlags().StringVar(&removeProjectSlug, "project-slug", "", "The slug of a project you wish to remove contracts")

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

	if removeProjectSlug != "" {
		projectConfiguration, exists := projectConfigurations[removeProjectSlug]
		if !exists {
			return userError.NewUserError(
				errors.Wrap(err, "cannot find project configuration via slug"),
				commands.Colorizer.Sprintf("Failed reading project configuration. Couldn't find project with slug: %s",
					commands.Colorizer.Bold(commands.Colorizer.Red(removeProjectSlug)),
				),
			)
		}

		projectConfigurations = map[string]*commands.ProjectConfiguration{
			removeProjectSlug: projectConfiguration,
		}
	}

	pushErrors := make(map[string]*userError.UserError)
	for projectSlug := range projectConfigurations {
		logrus.Info(commands.Colorizer.Sprintf(
			"Removing Smart Contracts for project: %s",
			commands.Colorizer.Bold(commands.Colorizer.Green(projectSlug)),
		))
		s := spinner.New(spinner.CharSets[33], 100*time.Millisecond)
		s.Start()

		var removeContractsRequest payloads.RemoveContractsRequest

		if removeContractID == "" {
			contractsResponse, err := rest.Contract.GetContracts(projectSlug)
			if err != nil {
				pushErrors[projectSlug] = userError.NewUserError(
					fmt.Errorf("failed to get contracts: %s", err),
					fmt.Sprintf("Couldn't get contracts from Project: %s", projectSlug),
				)
				continue
			}
			if contractsResponse.Error != nil {
				pushErrors[projectSlug] = userError.NewUserError(
					fmt.Errorf("api error getting contracts: %s", contractsResponse.Error.Slug),
					contractsResponse.Error.Message,
				)
				continue
			}

			for _, contract := range contractsResponse.Contracts {
				if removeContractTag == "" {
					removeContractsRequest.ContractIDs = append(removeContractsRequest.ContractIDs, contract.ID)
					continue
				}

				for _, tag := range contract.Tags {
					if removeContractTag == tag.Tag {
						removeContractsRequest.ContractIDs = append(removeContractsRequest.ContractIDs, contract.ID)
					}
				}
			}
		} else {
			removeContractsRequest.ContractIDs = append(removeContractsRequest.ContractIDs, removeContractID)
		}

		removeContractsResponse, err := rest.Contract.RemoveContracts(removeContractsRequest, projectSlug)
		if err != nil {
			pushErrors[projectSlug] = userError.NewUserError(
				fmt.Errorf("failed to remove contracts: %s", err),
				fmt.Sprintf("Couldn't remove contracts from Project: %s", projectSlug),
			)
			continue
		}
		if removeContractsResponse != nil && removeContractsResponse.Error != nil {
			pushErrors[projectSlug] = userError.NewUserError(
				fmt.Errorf("api error removing contracts: %s", removeContractsResponse.Error.Slug),
				removeContractsResponse.Error.Message,
			)
			continue
		}

		s.Stop()
	}

	for projectSlug, pushError := range pushErrors {
		userError.LogErrorf(fmt.Sprintf("Remove for %s failed with error: ", projectSlug)+"%s", pushError)
	}
	if len(pushErrors) > 0 {
		return userError.NewUserError(errors.New("some project contracts remove failed"), "Some of the project contracts remove were not successful. Please see the list above")
	}

	return nil
}

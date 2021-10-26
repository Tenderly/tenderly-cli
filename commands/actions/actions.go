package actions

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/commands/util"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/model"
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/userError"
	"gopkg.in/yaml.v3"
)

var actionsProjectName string

func init() {
	actionsCmd.PersistentFlags().StringVar(&actionsProjectName, "project", "", "The project slug in which the actions will published & deployed")

	commands.RootCmd.AddCommand(actionsCmd)
}

var actionsCmd = &cobra.Command{
	Use:   "actions",
	Short: "Create, build and deploy Web3 Actions.",
	Long:  "Web3 Actions will run your code in response to on-chain (or even off-chain) events, usually on your smart contracts.",
	Run: func(cmd *cobra.Command, args []string) {
		commands.CheckLogin()

		logrus.Info(commands.Colorizer.Sprintf("\nWelcome to Web3 Actions!\n"+
			"Initialize actions project with %s.\n"+
			"Build actions project with %s.\n"+
			"Deploy actions project with %s.\n",
			commands.Colorizer.Bold(commands.Colorizer.Green("tenderly actions init")),
			commands.Colorizer.Bold(commands.Colorizer.Green("tenderly actions build")),
			commands.Colorizer.Bold(commands.Colorizer.Green("tenderly actions deploy")),
		))
	},
}

func chooseProject(rest *rest.Rest, accountID string, createNewOption bool, onlySlugs []string) string {
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

	project := commands.GetProjectFromFlag(actionsProjectName, projectsResponse.Projects, rest)

	if project == nil {
		if onlySlugs == nil {
			project = commands.PromptProjectSelect(projectsResponse.Projects, rest, createNewOption)
		} else {
			var filteredProjects []*model.Project

			for _, returnedProject := range projectsResponse.Projects {
				projectSlug := returnedProject.Slug
				if returnedProject.OwnerInfo != nil {
					projectSlug = fmt.Sprintf("%s/%s", returnedProject.OwnerInfo.Username, projectSlug)
				}

				include := false
				for _, slug := range onlySlugs {
					if strings.ToLower(projectSlug) == strings.ToLower(slug) {
						include = true
						break
					}
				}
				if include {
					filteredProjects = append(filteredProjects, returnedProject)
				}
			}

			if len(filteredProjects) == 1 {
				project = filteredProjects[0]
			} else {
				project = commands.PromptProjectSelect(filteredProjects, rest, createNewOption)
			}
		}
	}

	if project == nil {
		userError.LogErrorf("project not found",
			userError.NewUserError(
				err,
				"Project must be selected for actions initialization.",
			),
		)
		os.Exit(1)
	}

	projectSlug := project.Slug
	if project.OwnerInfo != nil {
		projectSlug = fmt.Sprintf("%s/%s", project.OwnerInfo.Username, projectSlug)
	}
	return projectSlug
}

type actionsTenderlyYaml struct {
	Actions map[string]actionsModel.ProjectActions `yaml:"actions"`
}

func mustGetActions() map[string]actionsModel.ProjectActions {
	if !config.IsAnyActionsInit() {
		logrus.Error(commands.Colorizer.Sprintf(
			"Actions not initialized. Are you in the right directory? Run %s to initialize project.",
			commands.Colorizer.Bold(commands.Colorizer.Red("tenderly actions init")),
		))
		os.Exit(1)
	}

	content, err := config.ReadProjectConfig()
	if err != nil {
		userError.LogErrorf("failed reading project config: %s",
			userError.NewUserError(
				err,
				"Failed reading project's tenderly.yaml config. This can happen if you are running an older version of the Tenderly CLI.",
			),
		)
		os.Exit(1)
	}

	var tenderlyYaml actionsTenderlyYaml
	err = yaml.Unmarshal(content, &tenderlyYaml)
	if err != nil {
		userError.LogErrorf("failed unmarshalling actions config: %s",
			userError.NewUserError(
				err,
				"Failed parsing actions configuration. This can happen if you are running an older version of the Tenderly CLI.",
			),
		)
		os.Exit(1)
	}

	return tenderlyYaml.Actions
}

func mustGetProjectActions(actions map[string]actionsModel.ProjectActions, projectSlug string) *actionsModel.ProjectActions {
	ret, exists := actions[projectSlug]
	if !exists {
		ret, exists = actions[strings.ToLower(projectSlug)]
	}

	if !exists {
		logrus.Error(commands.Colorizer.Sprintf(
			"Actions not found for specified project %s.",
			commands.Colorizer.Bold(commands.Colorizer.Red(projectSlug)),
		))
		os.Exit(1)
	}

	return &ret
}

func mustInstallDependencies(sourcesDir string) {
	exists := util.PackageJSONExists(sourcesDir)
	if !exists {
		return
	}

	packageJSON := util.MustLoadPackageJSON(sourcesDir)
	if len(packageJSON.Dependencies)+len(packageJSON.DevDependencies) == 0 {
		return
	}
	logrus.Info("\nInstalling dependencies...")

	cmd := exec.Command("npm", "--prefix", sourcesDir, "install", "--verbose")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		userError.LogErrorf("failed to run npm install dependencies: %s",
			userError.NewUserError(err,
				commands.Colorizer.Sprintf(
					"Failed to run: %s.",
					commands.Colorizer.Bold(commands.Colorizer.Red(fmt.Sprintf("npm --prefix %s install", sourcesDir))),
				),
			),
		)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		userError.LogErrorf("failed to finish npm install dependencies.",
			userError.NewUserError(err,
				commands.Colorizer.Sprintf(
					"Failed to run: %s.",
					commands.Colorizer.Bold(commands.Colorizer.Red(fmt.Sprintf("npm --prefix %s install", sourcesDir))),
				),
			),
		)
		os.Exit(1)
	}
}

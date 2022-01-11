package actions

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/commands/util"
	"github.com/tenderly/tenderly-cli/config"
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	"github.com/tenderly/tenderly-cli/typescript"
	"github.com/tenderly/tenderly-cli/userError"
)

const (
	TypescriptActionsDependency        = "@tenderly/actions"
	TypescriptActionsDependencyVersion = "^0.0.7"
	LanguageJavaScript                 = "javascript"
	LanguageTypeScript                 = "typescript"
)

var actionsSourcesDir string
var actionsLanguage string
var actionsTemplateName string

var initGitIgnore = `# Dependency directories
node_modules/

# Ignore tsc output
out/**/*
`

var initActionTypescript = `import {
	ActionFn,
	Context,
	Event,
	BlockEvent
} from '@tenderly/actions'

export const blockHelloWorldFn: ActionFn = async (context: Context, event: Event) => {
	let blockEvent = event as BlockEvent
	console.log(blockEvent)
}
`

var initActionJavascript = `const blockHelloWorldFn = async (context, event) => {
	console.log(event)
}
module.exports = { blockHelloWorldFn }
`

var initDescription = "This is just an example, but you can publish this action."

var initAction = &actionsModel.ActionSpec{
	Description: &initDescription,
	Function:    "example:blockHelloWorldFn",
	Trigger:     actionsModel.TriggerUnparsed{Type: "block"},
}

func init() {
	initCmd.PersistentFlags().StringVar(&actionsSourcesDir, "sources", "", "The path where the actions will be created.")
	initCmd.PersistentFlags().StringVar(&actionsLanguage, "language", "typescript", "Initialize actions for this language. Supported {javascript, typescript}")
	initCmd.PersistentFlags().StringVar(&actionsTemplateName, "template", "", "Initialize actions from this template, see Tenderly/tenderly-actions.")

	actionsCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize actions project",
	Long:  "Guides you through setting up an actions project. It will populate sources directory and create configuration file.",
	Run: func(cmd *cobra.Command, args []string) {
		commands.CheckLogin()

		mustValidateFlags()

		rest := commands.NewRest()
		accountID := config.GetString(config.AccountID)
		projectSlug := chooseProject(rest, accountID, true, nil)

		if config.IsActionsInit(projectSlug) {
			userError.LogErrorf(
				"Actions for project are already initialized",
				userError.NewUserError(
					fmt.Errorf("actions initialized"),
					commands.Colorizer.Sprintf("Actions for project %s are already initialized, see %s",
						commands.Colorizer.Bold(commands.Colorizer.Green(projectSlug)),
						commands.Colorizer.Bold(commands.Colorizer.Green("tenderly.yaml")),
					),
				))
			os.Exit(0)
		}

		var (
			template     *actionsModel.Template
			templateArgs map[string]string
		)
		if actionsTemplateName != "" {
			var err error
			template, err = actionsModel.LoadTemplate(cmd.Context(), actionsTemplateName)
			if err != nil {
				userError.LogErrorf(
					"failed to load template",
					userError.NewUserError(err, fmt.Sprintf("Failed to load template %s", actionsTemplateName)))
				os.Exit(1)
			}

			templateArgs = chooseArgs(template.Args)
		}

		sources := chooseSources()
		util.CreateDir(sources)

		var specs map[string]*actionsModel.ActionSpec
		if template == nil {
			specs = map[string]*actionsModel.ActionSpec{
				"example": initAction,
			}
		} else {
			var err error
			specs, err = template.LoadSpecs(cmd.Context(), templateArgs)
			if err != nil {
				userError.LogErrorf(
					"failed to load template specs",
					userError.NewUserError(err, fmt.Sprintf("Failed to load specs for template %s", actionsTemplateName)))
				os.Exit(1)
			}
		}

		config.MustWriteActionsInit(projectSlug, &actionsModel.ProjectActions{
			Runtime:      "v1",
			Sources:      sources,
			Dependencies: nil,
			Specs:        specs,
		})

		if actionsLanguage == LanguageJavaScript {
			filePath := filepath.Join(sources, "example.js")
			content := initActionJavascript

			util.CreateFileWithContent(filePath, content)

			logrus.Info(commands.Colorizer.Sprintf("\nInitialized actions project. Sources directory created at %s. Configuration created in %s.",
				commands.Colorizer.Bold(commands.Colorizer.Green(sources)),
				commands.Colorizer.Bold(commands.Colorizer.Green("tenderly.yaml")),
			))

			os.Exit(0)
		}

		// Typescript
		if template == nil {
			// Package.json
			packageJson := typescript.DefaultPackageJson(filepath.Base(sources))
			if packageJson.Dependencies == nil {
				packageJson.Dependencies = make(map[string]string)
			}
			packageJson.Dependencies[TypescriptActionsDependency] = TypescriptActionsDependencyVersion
			util.MustSavePackageJSON(sources, packageJson)

			// Gitignore
			util.CreateFileWithContent(filepath.Join(sources, typescript.GitIgnoreFile), initGitIgnore)

			// Tsconfig
			util.MustSaveTsConfig(sources, typescript.DefaultTsConfig())

			// Exclude
			parentTsconfigPath := findTsconfigParent(sources)
			if parentTsconfigPath != nil {
				excludeFromTsconfigParent(sources, *parentTsconfigPath)
			}

			filePath := filepath.Join(sources, "example.ts")
			content := initActionTypescript
			util.CreateFileWithContent(filePath, content)
		} else {
			err := template.Create(cmd.Context(), sources, templateArgs)
			if err != nil {
				userError.LogErrorf(
					"failed to create from template",
					userError.NewUserError(err, fmt.Sprintf("Failed to create from template %s", actionsTemplateName)))
				os.Exit(1)
			}
		}

		// Install dependencies
		mustInstallDependencies(sources)

		logrus.Info(commands.Colorizer.Sprintf("\nInitialized actions project. Sources directory created at %s. Configuration created in %s.",
			commands.Colorizer.Bold(commands.Colorizer.Green(sources)),
			commands.Colorizer.Bold(commands.Colorizer.Green("tenderly.yaml")),
		))

		os.Exit(0)
	},
}

func chooseArgs(args []actionsModel.TemplateArg) map[string]string {
	ret := make(map[string]string)
	for _, arg := range args {
		val := promptTemplateArg(arg)
		ret[arg.Name] = val
	}
	return ret
}

func promptTemplateArg(arg actionsModel.TemplateArg) string {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("Enter value for %s (%s)", arg.Name, arg.Description),
		Validate: func(input string) error {
			if input == "" {
				return errors.New("value must not be empty")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		userError.LogErrorf("prompt template arg failed: %s", err)
		os.Exit(1)
	}

	if result == "" {
		userError.LogErrorf(
			"value for template arg not entered",
			userError.NewUserError(errors.New("enter template arg"),
				"Value for template arg not entered correctly"),
		)
		os.Exit(1)
	}
	return result
}

func mustValidateFlags() {
	if actionsLanguage != LanguageTypeScript && actionsLanguage != LanguageJavaScript {
		userError.LogErrorf(
			"language not supported",
			userError.NewUserError(errors.New("language not supported"), fmt.Sprintf("Language %s not supported", actionsLanguage)))
		os.Exit(1)
	}
}

func chooseSources() string {
	if actionsSourcesDir != "" {
		if util.ExistFile(actionsSourcesDir) {
			userError.LogErrorf(
				"sources dir is file: %s",
				userError.NewUserError(errors.New("sources dir is file"),
					"Selected sources directory is a file."),
			)
			os.Exit(1)
		}

		if util.ExistDir(actionsSourcesDir) {
			userError.LogErrorf(
				"sources dir exists: %s",
				userError.NewUserError(errors.New("sources dir exists"),
					"Selected sources directory already exists."),
			)
			os.Exit(1)
		}

		return actionsSourcesDir
	} else {
		defaultSources := getDefaultInitSourcesDir()
		directory := commands.PromptNewDirectory("actions sources", defaultSources)
		return directory
	}
}

func getDefaultInitSourcesDir() string {
	defaultDir := "src/"
	if !util.ExistDir(defaultDir) {
		defaultDir = ""
	}

	var sourcesDir string
	i := 0
	for {
		if i == 0 {
			sourcesDir = filepath.Join(defaultDir, "actions")
		} else {
			sourcesDir = filepath.Join(defaultDir, fmt.Sprintf("%s-%d", "actions", i))
		}
		if !util.ExistDir(sourcesDir) {
			return sourcesDir
		}
		i++
	}
}

func findTsconfigParent(path string) *string {
	if !util.ExistDir(path) {
		return nil
	}
	for {
		parent := filepath.Dir(path)
		if !util.ExistDir(parent) {
			return nil
		}

		if util.ExistFile(filepath.Join(parent, "tsconfig.json")) {
			return &parent
		}

		path = parent
		if path == "." || path == "" || path == string(os.PathSeparator) {
			return nil
		}
	}
}

func excludeFromTsconfigParent(sourcesPath string, parentPath string) {
	rel := sourcesPath

	if parentPath != "." && parentPath != "" {
		relTmp, err := filepath.Rel(parentPath, sourcesPath)
		if err != nil {
			userError.LogErrorf(
				"failed to get path relative to parentPath",
				userError.NewUserError(err, fmt.Sprintf("Can't find relative path for %s", sourcesPath)))
			os.Exit(1)
		}
		rel = relTmp
	}

	tsconfig := util.MustLoadTsConfig(parentPath)
	tsconfig.Exclude = append(tsconfig.Exclude, rel)

	util.MustSaveTsConfig(parentPath, tsconfig)
}

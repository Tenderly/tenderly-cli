package actions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tenderly/tenderly-cli/commands"
	"github.com/tenderly/tenderly-cli/commands/util"
	"github.com/tenderly/tenderly-cli/commands/util/packagejson"
	"github.com/tenderly/tenderly-cli/config"
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	"github.com/tenderly/tenderly-cli/rest"
	conjureactions "github.com/tenderly/tenderly-cli/rest/payloads/generated/actions"
	"github.com/tenderly/tenderly-cli/typescript"
	"github.com/tenderly/tenderly-cli/userError"
)

var (
	zipLimitBytes          = 45 * 1024 * 1024 // 45 MB
	srcPathInZip           = "src/"
	nodeModulesPathInZip   = "nodejs/node_modules/"
	possibleFileExtensions = []string{"ts", "js"}
	ActionUrlPattern       = "https://dashboard.tenderly.co/%s/action/%s"
)

// Set and access from commands
var (
	r                 *rest.Rest
	projectSlug       string
	actions           *actionsModel.ProjectActions
	outDir            string
	sources           map[string]string
	logicExist        bool
	dependenciesExist bool
)

func init() {
	actionsCmd.AddCommand(buildCmd)
	actionsCmd.AddCommand(publishCmd)
	actionsCmd.AddCommand(deployCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build actions for project",
	Long:  "If you just want to validate configuration or build implementation without deploying.",
	Run:   buildFunc,
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish actions for project",
	Long:  "If you just want to publish actions to dashboard without deploying.",
	Run:   publishFunc,
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy actions for project",
	Long:  "If you are ready to deploy actions. Deployed actions will be scheduled if they have configured trigger.",
	Run:   deployFunc,
}

func buildFunc(cmd *cobra.Command, args []string) {
	commands.CheckLogin()
	r = commands.NewRest()

	allActions := mustGetActions()
	var slugs []string
	for k := range allActions {
		slugs = append(slugs, k)
	}

	accountID := config.GetString(config.AccountID)
	projectSlug = chooseProject(r, accountID, false, slugs)

	actions = mustGetProjectActions(allActions, projectSlug)
	logrus.Info("\nBuilding actions:")
	for actionName := range actions.Specs {
		logrus.Info(commands.Colorizer.Sprintf(
			"- %s", commands.Colorizer.Bold(commands.Colorizer.Green(actionName))))
	}

	util.MustExistDir(actions.Sources)
	if !actionsModel.IsRuntimeSupported(actions.Runtime) {
		logrus.Error(commands.Colorizer.Sprintf(
			"Configured runtime %s is not supported. Supported values: {%s}",
			commands.Colorizer.Bold(commands.Colorizer.Red(actions.Runtime)),
			commands.Colorizer.Bold(commands.Colorizer.Green(strings.Join(actionsModel.SupportedRuntimes, ","))),
		))
		os.Exit(1)
	}
	mustParseAndValidateTriggers(actions)

	tsConfigExists := util.TsConfigExists(actions.Sources)
	tsFileExists, tsFile := anyFunctionTsFileExists(actions)
	if tsFileExists && !tsConfigExists {
		err := errors.New(fmt.Sprintf("File %s is a typescript file but there is no typescript config file!", tsFile))
		userError.LogErrorf("missing typescript config file %s",
			userError.NewUserError(err,
				commands.Colorizer.Sprintf(
					"Missing typescript config file in your sources! Sources: %s, File: %s",
					commands.Colorizer.Bold(commands.Colorizer.Red(actions.Sources)),
					commands.Colorizer.Bold(commands.Colorizer.Red(tsFile))),
			),
		)
		os.Exit(1)
	}

	var tsconfig *typescript.TsConfig
	if tsConfigExists {
		tsconfig = util.MustLoadTsConfig(actions.Sources)
		mustValidateTsconfig(tsconfig)
	}

	outDir = actions.Sources

	exists := util.PackageJSONExists(actions.Sources)
	if exists {
		packageJSON := util.MustLoadPackageJSON(actions.Sources)
		if util.HasDependencies(packageJSON) {
			logrus.Info("\nValidating package.json dependencies...")

			validator := packagejson.NewValidator(actions.Runtime)
			result, err := mustValidateDependencies(packageJSON, validator)
			if err != nil {
				userError.LogErrorf("failed to validate package.json packages: %s", err)
				os.Exit(1)
			}

			if !result.Success {
				printPackageValidationErrors(result.Errors)
				os.Exit(1)
			}

			logrus.Info("\nFinished validating package.json dependencies...")
		}
	}

	if tsconfig != nil {
		outDir = filepath.Join(outDir, *tsconfig.CompilerOptions.OutDir)
		mustInstallDependencies(actions.Sources)
		mustBuildProject(actions.Sources, tsconfig)
		mustExistCompiledFiles(outDir, actions)
	}

	sources = mustValidateAndGetSources(r, actions, projectSlug, actions.Sources)
	logrus.Info(commands.Colorizer.Green("\nBuild completed."))
}

func mustParseAndValidateTriggers(projectActions *actionsModel.ProjectActions) {
	for name, spec := range projectActions.Specs {
		err := spec.Parse()
		if err != nil {
			userError.LogErrorf("failed parsing action trigger with %s",
				userError.NewUserError(err,
					commands.Colorizer.Sprintf(
						"Failed parsing action trigger for %s",
						commands.Colorizer.Bold(commands.Colorizer.Red(name))),
				),
			)
			os.Exit(1)
		}
	}

	logrus.Info("\nValidating triggers configuration...")
	errors := false
	for name, spec := range projectActions.Specs {
		validatorResponse := spec.TriggerParsed.Validate(actionsModel.ValidatorContext(name + ".trigger"))
		for _, i := range validatorResponse.Infos {
			logrus.Info(commands.Colorizer.Blue(i))
		}
		if len(validatorResponse.Errors) > 0 {
			errors = true
			for _, e := range validatorResponse.Errors {
				logrus.Info(commands.Colorizer.Red(e))
			}
		}
	}
	if errors {
		logrus.Error(commands.Colorizer.Bold(commands.Colorizer.Red("Found errors when validating triggers")))
		os.Exit(1)
	}
}

func publishFunc(cmd *cobra.Command, args []string) {
	buildFunc(cmd, args)
	publish(r, actions, sources, projectSlug, outDir, false)
}

func deployFunc(cmd *cobra.Command, args []string) {
	buildFunc(cmd, args)
	publish(r, actions, sources, projectSlug, outDir, true)
}

func publish(
	r *rest.Rest,
	actions *actionsModel.ProjectActions,
	sources map[string]string,
	projectSlug string,
	outDir string,
	deploy bool,
) {
	if !deploy {
		logrus.Info("\nPublishing actions:")
	} else {
		logrus.Info("\nPublishing and deploying actions:")
	}
	for actionName := range actions.Specs {
		logrus.Info(commands.Colorizer.Sprintf(
			"- %s", commands.Colorizer.Bold(commands.Colorizer.Green(actionName))))
	}

	logicZip, logicHash := util.MustZipAndHashDir(outDir, srcPathInZip, zipLimitBytes)
	if logicExist {
		logicZip = nil
	}

	dependenciesDir := filepath.Join(actions.Sources, typescript.NodeModulesDir)
	dependenciesZip, dependenciesHash := util.ZipAndHashDir(dependenciesDir, nodeModulesPathInZip, zipLimitBytes)
	if dependenciesExist {
		dependenciesZip = nil
	}

	// TODO(marko): Send package-lock.json in publish request
	request := conjureactions.PublishRequest{
		Actions:             actions.ToRequest(sources),
		Deploy:              deploy,
		Commitish:           util.GetCommitish(),
		LogicZip:            &logicZip,
		LogicVersion:        &logicHash,
		DependenciesZip:     &dependenciesZip,
		DependenciesVersion: &dependenciesHash,
	}

	s := spinner.New(spinner.CharSets[33], 100*time.Millisecond)
	s.Start()
	response, err := r.Actions.Publish(request, projectSlug)
	s.Stop()

	if err != nil {
		userError.LogErrorf("publish request failed",
			userError.NewUserError(
				err,
				commands.Colorizer.Sprintf("Publish request failed: %s",
					commands.Colorizer.Red(err.Error())),
			),
		)
		os.Exit(1)
	}

	if !deploy {
		logrus.Info("\nPublished actions:")
	} else {
		logrus.Info("\nPublished and deployed actions:")
	}
	for key, version := range response.Actions {
		logrus.Info(commands.Colorizer.Sprintf("- %s (actionId = %s, versionId = %s) %s",
			commands.Colorizer.Bold(commands.Colorizer.Green(key)),
			version.ActionId,
			version.Id,
			fmt.Sprintf(ActionUrlPattern, projectSlug, version.ActionId)))
	}
}

func mustBuildProject(sourcesDir string, tsconfig *typescript.TsConfig) {
	if tsconfig == nil {
		return
	}

	util.RemoveDirWithContent(filepath.Join(sourcesDir, *tsconfig.CompilerOptions.OutDir))

	logrus.Info("\nBuilding actions...")
	cmd := exec.Command("npm", "--prefix", sourcesDir, "run", typescript.DefaultBuildScriptName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		userError.LogErrorf("failed to run build typescript: %s",
			userError.NewUserError(err,
				commands.Colorizer.Sprintf(
					"Failed to run: %s.",
					commands.Colorizer.Bold(
						commands.Colorizer.Red(fmt.Sprintf("npm --prefix %s run build", sourcesDir)),
					),
				),
			),
		)
		os.Exit(1)
	}
}

func mustValidateAndGetSources(
	r *rest.Rest,
	actions *actionsModel.ProjectActions,
	projectSlug string,
	sourcesDir string,
) map[string]string {
	logrus.Info("\nValidating actions...")

	validatedSources := make(map[string]string)

	for name, spec := range actions.Specs {
		source := mustGetSource(sourcesDir, spec.Function)
		validatedSources[name] = source
	}

	logicExist, dependenciesExist = mustValidate(r, actions, validatedSources, projectSlug)

	return validatedSources
}

func mustGetSource(sourcesDir string, locator string) string {
	internalLocator, err := actionsModel.NewInternalLocator(locator)
	if err != nil {
		userError.LogErrorf("invalid locator: %s",
			userError.NewUserError(
				err,
				commands.Colorizer.Sprintf(
					"Invalid locator format %s.",
					commands.Colorizer.Bold(commands.Colorizer.Red(locator)),
				),
			),
		)
		os.Exit(1)
	}

	var (
		filePath string
		content  string
		exists   bool
	)
	for _, ext := range possibleFileExtensions {
		filePath = filepath.Join(sourcesDir, fmt.Sprintf("%s.%s", internalLocator.Path, ext))
		if util.ExistFile(filePath) {
			content = util.ReadFile(filePath)
			exists = true
			break
		}
	}
	if !exists {
		logrus.Error(commands.Colorizer.Sprintf(
			"Invalid locator %s (file %s not found).",
			commands.Colorizer.Bold(commands.Colorizer.Red(locator)),
			commands.Colorizer.Bold(commands.Colorizer.Red(filePath)),
		))
		os.Exit(1)
	}

	return content
}

// Validates sources and returns if source logic exist (LogicFound) and if dependencies logic exist (DependenciesFound)
func mustValidate(
	r *rest.Rest,
	actions *actionsModel.ProjectActions,
	sources map[string]string,
	projectSlug string,
) (bool, bool) {
	request := conjureactions.ValidateRequest{
		Actions:             actions.ToRequest(sources),
		LogicVersion:        nil,
		DependenciesVersion: nil,
	}

	_, logicHash := util.MustZipAndHashDir(outDir, srcPathInZip, zipLimitBytes)

	request.LogicVersion = &logicHash

	dependenciesDir := filepath.Join(actions.Sources, typescript.NodeModulesDir)

	_, dependenciesHash := util.ZipAndHashDir(dependenciesDir, nodeModulesPathInZip, zipLimitBytes)
	request.DependenciesVersion = &dependenciesHash

	response, err := r.Actions.Validate(request, projectSlug)
	if err != nil {
		userError.LogErrorf("validate request failed",
			userError.NewUserError(
				err,
				commands.Colorizer.Sprintf("Validate request failed: %s",
					commands.Colorizer.Red(err.Error())),
			),
		)
		os.Exit(1)
	}

	if len(response.Errors) > 0 {
		for name, errs := range response.Errors {
			logrus.Info(
				commands.Colorizer.Sprintf("Validation for %s failed with errors:", commands.Colorizer.Yellow(name)),
			)
			for _, e := range errs {
				logrus.Info(commands.Colorizer.Sprintf("%s: %s", commands.Colorizer.Red(e.Name), e.Message))
			}
		}
		os.Exit(1)
	}

	return response.LogicFound, response.DependenciesFound
}

func mustValidateTsconfig(tsconfig *typescript.TsConfig) {
	if tsconfig.CompilerOptions.OutDir == nil {
		logrus.Error(commands.Colorizer.Sprintf(
			"Invalid tsconfig - %s must be set.",
			commands.Colorizer.Bold(commands.Colorizer.Red("compilerOptions.outDir")),
		))
		os.Exit(1)
	}
}

func anyFunctionTsFileExists(actions *actionsModel.ProjectActions) (bool, string) {
	for _, spec := range actions.Specs {
		locator := spec.Function
		internalLocator, err := actionsModel.NewInternalLocator(locator)
		if err != nil {
			userError.LogErrorf("invalid locator: %s",
				userError.NewUserError(
					err,
					commands.Colorizer.Sprintf(
						"Invalid locator format %s.",
						commands.Colorizer.Bold(commands.Colorizer.Red(locator)),
					),
				),
			)
			os.Exit(1)
		}
		filePath := filepath.Join(actions.Sources, internalLocator.Path)
		if util.IsFileTs(filePath) {
			return true, fmt.Sprintf("%s%s", internalLocator.Path, typescript.TsFileExt)
		}
	}
	return false, ""
}

func mustExistCompiledFiles(outDir string, actions *actionsModel.ProjectActions) {
	missingFilePaths := make([]string, 0, len(actions.Specs))
	missingFileAlreadyAdded := make(map[string]bool)

	for _, spec := range actions.Specs {
		internalLocator, err := actionsModel.NewInternalLocator(spec.Function)
		if err != nil {
			userError.LogErrorf("invalid locator: %s",
				userError.NewUserError(err,
					commands.Colorizer.Sprintf(
						"Invalid locator format %s.",
						commands.Colorizer.Bold(commands.Colorizer.Red(spec.Function)),
					)))
			os.Exit(1)
		}

		filePath := filepath.Join(outDir, fmt.Sprintf("%s.js", internalLocator.Path))

		if !util.ExistFile(filePath) {
			if !missingFileAlreadyAdded[filePath] {
				missingFileAlreadyAdded[filePath] = true
				missingFilePaths = append(missingFilePaths, filePath)
			}
		}
	}
	if len(missingFilePaths) > 0 {
		logrus.Errorf("Unable to resolve path for some of the compiled files: %s\n"+
			"Make sure all imported files are contained in the configured action sources directory (%s).\n"+
			"If the problem persists, please run this command with the %s flag and send logs to our customer support.",
			commands.Colorizer.Bold(commands.Colorizer.Red(strings.Join(missingFilePaths, ", "))),
			commands.Colorizer.Bold(actions.Sources),
			commands.Colorizer.Bold(commands.Colorizer.Red("--debug")))
		os.Exit(1)
	}
}

func printPackageValidationErrors(validationErrors []*packagejson.ValidationError) {
	logrus.Error("The following packages have invalid versions:")
	for _, e := range validationErrors {
		logrus.Error(commands.Colorizer.Sprintf(
			"  %s\n\tFound: %s\n\tRequired: %s",
			commands.Colorizer.Bold(commands.Colorizer.Bold(e.Name)),
			commands.Colorizer.Bold(commands.Colorizer.Red(e.PackageJsonVersion)),
			commands.Colorizer.Bold(commands.Colorizer.Red(e.Constraint)),
		))
	}
}

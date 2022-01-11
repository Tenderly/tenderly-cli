package commands

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/userError"
)

type releaseResult struct {
	Name   string         `json:"name"`
	NodeId string         `json:"node_id"`
	Assets []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type releaseCliMessage struct {
	Format string                  `json:"format"`
	Parts  []releaseCliMessagePart `json:"parts"`
}

type releaseCliMessagePart struct {
	Text       string   `json:"text"`
	Formatting []string `json:"formatting"`
}

var versionAlreadyChecked bool

func init() {
	RootCmd.AddCommand(checkUpdatesCmd)
}

var checkUpdatesCmd = &cobra.Command{
	Use:   "update-check",
	Short: "Checks whether there is an update for the CLI",
	Run: func(cmd *cobra.Command, args []string) {
		CheckVersion(true, false)
	},
}

func CheckVersion(force bool, encounteredError bool) {
	if versionAlreadyChecked {
		return
	}

	rand.Seed(time.Now().UnixNano())
	randInt := rand.Intn(25)

	if !force && randInt != 24 {
		return
	}

	versionAlreadyChecked = true

	response, err := http.Get("https://api.github.com/repos/tenderly/tenderly-cli/releases")

	if err != nil || (response != nil && response.StatusCode != 200) {
		if force {
			if response != nil {
				logrus.Debugf("Status code: %d", response.StatusCode)
			}

			userError.LogErrorf("failed creating github releases request: %s", userError.NewUserError(
				err,
				"\nFailed fetching newest releases from GitHub. Please try again.\n",
			))
		}
		return
	}

	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		if force && !encounteredError {
			userError.LogErrorf("failed reading github releases request: %s", userError.NewUserError(
				err,
				Colorizer.Sprintf(
					"\nFailed parsing latest releases from GitHub. Please try again. If the problem persists, please follow the installation steps described at %s to re-install the CLI.\n",
					Colorizer.Bold(Colorizer.Green("https://github.com/Tenderly/tenderly-cli#installation")),
				),
			))
		}
		return
	}

	var result []releaseResult
	err = json.Unmarshal(contents, &result)

	if err != nil || len(result) == 0 {
		if force && !encounteredError {
			userError.LogErrorf("error unmarshaling github releases: %s", userError.NewUserError(
				err,
				Colorizer.Sprintf(
					"\nFailed parsing latest releases from GitHub. Please try again. If the problem persists, please follow the installation steps described at %s to re-install the CLI.\n",
					Colorizer.Bold(Colorizer.Green("https://github.com/Tenderly/tenderly-cli#installation")),
				),
			))
		}
		return
	}

	sort.Slice(result, func(i, j int) bool {
		v1, err := version.NewVersion(result[i].Name)
		if err != nil {
			return false
		}

		v2, err := version.NewVersion(result[j].Name)
		if err != nil {
			return true
		}

		return v1.GreaterThan(v2)
	})

	currentVersion, err := version.NewVersion(CurrentCLIVersion)
	if err != nil {
		if force && !encounteredError {
			userError.LogErrorf("cannot parse current cli version: %s", userError.NewUserError(
				err,
				Colorizer.Sprintf(
					"\nCannot parse the current version of the Tenderly CLI. Please follow the installation steps described at %s to re-install the CLI.\n",
					Colorizer.Bold(Colorizer.Green("https://github.com/Tenderly/tenderly-cli#installation")),
				),
			))
		}
		return
	}

	newestVersion, err := version.NewVersion(result[0].Name)
	if err != nil {
		if force && !encounteredError {
			userError.LogErrorf("cannot parse newest cli version: %s", userError.NewUserError(
				err,
				Colorizer.Sprintf(
					"\nCannot parse the newest version of the Tenderly CLI. Please follow the installation steps described at %s to re-install the CLI.\n",
					Colorizer.Bold(Colorizer.Green("https://github.com/Tenderly/tenderly-cli#installation")),
				),
			))
		}
		return
	}

	if !newestVersion.GreaterThan(currentVersion) {
		if force && !encounteredError {
			logrus.Info(Colorizer.Sprintf(
				"\nYou are already running the newest version of the Tenderly CLI: %s.\n",
				Colorizer.Bold(Colorizer.Green(CurrentCLIVersion)),
			))
		}
		return
	}

	var updateCommand string

	switch runtime.GOOS {
	case "darwin":
		updateCommand = getMacOSInstallationCommand()
	case "linux":
		fallthrough
	default:
		updateCommand = Colorizer.Sprintf("%s", Colorizer.Bold(Colorizer.Green("curl https://raw.githubusercontent.com/Tenderly/tenderly-cli/master/scripts/install-linux.sh | sh")))
	}

	logrus.Info(
		Colorizer.Sprintf("\nYou are running version %s of the Tenderly CLI. To update to the newest version (%s) please follow the instructions below:\n\n%s\n",
			Colorizer.Bold(Colorizer.Green(CurrentCLIVersion)),
			Colorizer.Bold(Colorizer.Green(result[0].Name)),
			updateCommand,
		),
	)

	cliMessage, err := getCliMessage(result[0])
	if err != nil || cliMessage == "" {
		logrus.Debugf("error fetching cli message: %s", err)
		return
	}

	logrus.Infof("Below are the notes for this release:\n\n%s\n", cliMessage)
}

func getMacOSInstallationCommand() string {
	path, err := os.Executable()

	defaultMessage := Colorizer.Sprintf("If you installed the CLI via Homebrew you can update it by running:\n\n%s\n\n"+
		"Alternatively, if you installed the CLI via the installation script, you can update your installation by running the same script again:\n\n%s",
		Colorizer.Bold(Colorizer.Green("brew update && brew upgrade tenderly")),
		Colorizer.Bold(Colorizer.Green("curl https://raw.githubusercontent.com/Tenderly/tenderly-cli/master/scripts/install-macos.sh | sh")),
	)

	if err != nil {
		return defaultMessage
	}

	link, err := os.Readlink(path)

	if err != nil {
		return defaultMessage
	}

	if strings.Contains(link, "Cellar") {
		return Colorizer.Sprintf("It seems you installed the CLI via Homebrew, so you can update it by running:\n\n%s",
			Colorizer.Bold(Colorizer.Green("brew update && brew upgrade tenderly")),
		)
	}

	return Colorizer.Sprintf("It seems you installed the CLI via the installation script, so you can update it by running:\n\n%s",
		Colorizer.Bold(Colorizer.Green("curl https://raw.githubusercontent.com/Tenderly/tenderly-cli/master/scripts/install-macos.sh | sh")),
	)
}

func getCliMessage(release releaseResult) (string, error) {
	var assetUrl string
	for _, asset := range release.Assets {
		if asset.Name == "cli-message.json" {
			assetUrl = asset.BrowserDownloadUrl
			break
		}
	}

	if assetUrl == "" {
		return "", errors.New("couldn't find cli-message.json in asset list")
	}

	response, err := http.Get(assetUrl)

	if err != nil {
		return "", errors.Wrap(err, "get cli-message.json")
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrap(err, "read cli-message.json")
	}

	var cliMessage releaseCliMessage

	err = json.Unmarshal(body, &cliMessage)
	if err != nil {
		return "", errors.Wrap(err, "unmarshal cli-message.json")
	}

	var args []interface{}

	for _, part := range cliMessage.Parts {
		formattedPart := Colorizer.White(part.Text)

		for i := len(part.Formatting) - 1; i >= 0; i-- {
			switch part.Formatting[i] {
			case "bold":
				formattedPart = Colorizer.Bold(formattedPart)
			case "green":
				formattedPart = Colorizer.Green(formattedPart)
			case "red":
				formattedPart = Colorizer.Red(formattedPart)
			}
		}

		args = append(args, formattedPart)
	}

	message := Colorizer.Sprintf(cliMessage.Format, args...)

	return message, nil
}

package main

import (
	"encoding/json"
	"fmt"
	goVersion "github.com/hashicorp/go-version"
	"github.com/logrusorgru/aurora"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

type releaseResult struct {
	Name   string `json:"name"`
	NodeId string `json:"node_id"`
}

var versionAlreadyChecked bool

func CheckVersion(force bool) {
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

	if err != nil {
		return
	}

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	var result []releaseResult
	err = json.Unmarshal(contents, &result)

	if err != nil || len(result) == 0 {
		return
	}

	sort.Slice(result, func(i, j int) bool {
		v1, err := goVersion.NewVersion(result[i].Name)
		if err != nil {
			return true
		}

		v2, err := goVersion.NewVersion(result[j].Name)

		return v1.GreaterThan(v2)
	})

	currentVersion, err := goVersion.NewVersion(CurrentCLIVersion)
	if err != nil {
		return
	}

	newestVersion, err := goVersion.NewVersion(result[0].Name)
	if err != nil {
		return
	}

	if !newestVersion.GreaterThan(currentVersion) {
		return
	}

	var updateCommand string

	switch runtime.GOOS {
	case "darwin":
		updateCommand = getMacOSInstallationCommand()
		break
	case "linux":
	default:
		updateCommand = aurora.Sprintf("%s", aurora.Bold(aurora.Green("curl https://raw.githubusercontent.com/Tenderly/tenderly-cli/master/scripts/install-linux.sh | sh")))
	}

	fmt.Println(
		aurora.Sprintf("\nYou are running version %s of the Tenderly CLI. To update to the newest version (%s) please run the following command:\n\n%s\n\n",
			aurora.Bold(aurora.Green(CurrentCLIVersion)),
			aurora.Bold(aurora.Green(result[0].Name)),
			updateCommand,
		),
	)
}

func getMacOSInstallationCommand() string {
	path, err := os.Executable()

	defaultMessage := aurora.Sprintf("If you installed the CLI via Homebrew you can update it by running:\n\n%s\n\n"+
		"or if you installed it via the installation script you can just run the installation script again:\n\n%s",
		aurora.Bold(aurora.Green("brew update && brew upgrade tenderly")),
		aurora.Bold(aurora.Green("curl https://raw.githubusercontent.com/Tenderly/tenderly-cli/master/scripts/install-macos.sh | sh")),
	)

	if err != nil {
		return defaultMessage
	}

	link, err := os.Readlink(path)

	if err != nil {
		return defaultMessage
	}

	if strings.Contains(link, "Cellar") {
		return aurora.Sprintf("It seems you installed the CLI via Homebrew, so you can update it by running:\n\n%s\n\n",
			aurora.Bold(aurora.Green("brew update && brew upgrade tenderly")),
		)
	}

	return aurora.Sprintf("It seems you installed the CLI via the installation script, so you can update it by running:\n\n%s\n\n",
		aurora.Bold(aurora.Green("curl https://raw.githubusercontent.com/Tenderly/tenderly-cli/master/scripts/install-macos.sh | sh")),
	)
}

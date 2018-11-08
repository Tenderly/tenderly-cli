package main

import (
	"encoding/json"
	"fmt"
	goVersion "github.com/hashicorp/go-version"
	"github.com/logrusorgru/aurora"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"time"
)

type commit struct {
	Sha string `json:"sha"`
	Url string `json:"url"`
}

type tagResults struct {
	Name       string `json:"name"`
	ZipballUrl string `json:"zipball_url"`
	TarballUrl string `json:"tarball_url"`
	Commit     commit `json:"commit"`
	NodeId     string `json:"node_id"`
}

func MaybeCheckVersion() {
	rand.Seed(time.Now().UnixNano())
	randInt := rand.Intn(25)

	if randInt != 24 {
		return
	}

	response, err := http.Get("https://api.github.com/repos/tenderly/tenderly-cli/tags")

	if err != nil {
		return
	}

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	var result []tagResults
	err = json.Unmarshal(contents, &result)

	if err != nil {
		return
	}

	if len(result) == 0 {
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

	fmt.Println(
		aurora.Sprintf("\nYou are running version %s of the Tenderly CLI. To update to the newest version (%s) please run the following command:\n\n%s\n\n",
			aurora.Bold(aurora.Green(CurrentCLIVersion)),
			aurora.Bold(aurora.Green(result[0].Name)),
			aurora.Bold(aurora.Green("brew update && brew upgrade tenderly")),
		),
	)
}

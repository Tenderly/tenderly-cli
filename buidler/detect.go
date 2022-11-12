package buidler

import (
	"os"
	"path"
)

var buidlerFolders = []string{
	"deployments",
}

func FindDirectories() []string {
	return []string{}
}

func (dp *DeploymentProvider) ValidProviderStructure(directory string) bool {
	for _, buidlerFolder := range buidlerFolders {
		folderPath := path.Join(directory, buidlerFolder)
		if _, err := os.Stat(folderPath); err != nil {
			return false
		}
	}

	return true
}

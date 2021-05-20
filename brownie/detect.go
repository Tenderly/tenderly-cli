package brownie

import (
	"os"
	"path"
)

var brownieFolders = []string{
	"build",
}

func FindDirectories() []string {
	return []string{}
}

func (dp *DeploymentProvider) CheckIfProviderStructure(directory string) bool {
	for _, folder := range brownieFolders {
		folderPath := path.Join(directory, folder)
		if _, err := os.Stat(folderPath); err != nil {
			return false
		}
	}

	return true
}

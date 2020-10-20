package hardhat

import (
	"os"
	"path"
)

var hardhatFolders = []string{
	"deployments",
}

func FindDirectories() []string {
	return []string{}
}

func (dp *DeploymentProvider) CheckIfProviderStructure(directory string) bool {
	for _, buidlerFolder := range hardhatFolders {
		folderPath := path.Join(directory, buidlerFolder)
		if _, err := os.Stat(folderPath); err != nil {
			return false
		}
	}

	return true
}

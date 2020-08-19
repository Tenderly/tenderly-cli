package openzeppelin

import (
	"os"
	"path"
)

var openZeppelinFolders = []string{
	"contracts",
	".openzeppelin",
}

func FindDirectories() []string {
	return []string{}
}

func (dp *DeploymentProvider) CheckIfProviderStructure(directory string) bool {
	for _, openZeppelinFolder := range openZeppelinFolders {
		folderPath := path.Join(directory, openZeppelinFolder)
		if _, err := os.Stat(folderPath); err != nil {
			return false
		}
	}

	return true
}

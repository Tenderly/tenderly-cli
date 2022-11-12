package openzeppelin

import (
	"os"
	"path/filepath"
)

var openZeppelinFolders = []string{
	"contracts",
	".openzeppelin",
}

func FindDirectories() []string {
	return []string{}
}

func (dp *DeploymentProvider) ValidProviderStructure(directory string) bool {
	for _, openZeppelinFolder := range openZeppelinFolders {
		folderPath := filepath.Join(directory, openZeppelinFolder)
		if _, err := os.Stat(folderPath); err != nil {
			return false
		}
	}

	return true
}

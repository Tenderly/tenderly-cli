package brownie

import (
	"os"
	"path"
)

func (p Provider) ValidProviderStructure(directory string) bool {
	var brownieFolders = []string{
		"build",
	}

	for _, folder := range brownieFolders {
		folderPath := path.Join(directory, folder)
		if _, err := os.Stat(folderPath); err != nil {
			return false
		}
	}

	return true
}

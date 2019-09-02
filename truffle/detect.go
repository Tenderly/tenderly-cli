package truffle

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strings"
)

var truffleFolders = []string{
	"contracts",
	"migrations",
}

func FindTruffleDirectories() []string {
	if runtime.GOOS != "darwin" {
		return nil
	}

	cmd := exec.Command("/bin/sh", "-c", "command -v mdfind")
	if err := cmd.Run(); err != nil {
		return nil
	}

	data, err := exec.Command(
		"mdfind",
		"-onlyin",
		getHomeDir(),
		"kMDItemDisplayName == truffle*.js",
	).CombinedOutput()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":  err,
			"output": string(data),
		}).Debug("Couldn't find truffle directories")
		return nil
	}

	possibleDirectories := strings.Split(string(data), "\n")

	directories := map[string]bool{}

	for _, possibleDirectory := range possibleDirectories {
		if strings.Contains(possibleDirectory, "node_modules") {
			continue
		}

		dir := path.Dir(possibleDirectory)
		if !CheckIfTruffleStructure(dir) {
			continue
		}

		directories[dir] = true
	}

	var result []string

	for dir := range directories {
		if dir == "." {
			continue
		}

		result = append(result, dir)
	}

	return result
}

func CheckIfTruffleStructure(directory string) bool {
	for _, truffleFolder := range truffleFolders {
		folderPath := path.Join(directory, truffleFolder)
		if _, err := os.Stat(folderPath); err != nil {
			return false
		}
	}

	return true
}

func getHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		return "~"
	}

	return usr.HomeDir
}

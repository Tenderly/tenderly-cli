package typescript

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type Npm interface {
	Build() error
}

func NewNpm(directory string, packageJson *PackageJson) Npm {
	return &npm{
		directory:   directory,
		packageJson: packageJson,
	}
}

type npm struct {
	directory   string
	packageJson *PackageJson
}

var _ Npm = (*npm)(nil)

func (n *npm) Build() error {
	_, exists := n.packageJson.Scripts[DefaultBuildScriptName]
	if !exists {
		return errors.New("npm build script not found")
	}

	cmd := exec.Command("npm", "run", DefaultBuildScriptName)
	cmd.Dir = n.directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

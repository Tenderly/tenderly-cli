package typescript

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type PackageJson struct {
	Name            string            `json:"name,omitempty"`
	Scripts         map[string]string `json:"scripts,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	Private         bool              `json:"private,omitempty"`
}

func DefaultPackageJson(name string) *PackageJson {
	return &PackageJson{
		Name:            name,
		Scripts:         map[string]string{DefaultBuildScriptName: DefaultBuildScript},
		DevDependencies: map[string]string{"typescript": DefaultTypescriptVersion},
		Private:         true,
	}
}

func LoadPackageJson(directory string) (*PackageJson, error) {
	path := filepath.Join(directory, PackageJsonFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read package.json")
	}

	var value PackageJson
	err = json.Unmarshal(data, &value)
	if err != nil {
		return nil, errors.Wrap(err, "parse package.json")
	}

	return &value, nil
}

func SavePackageJson(directory string, config *PackageJson) error {
	packageJSON, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return errors.Wrap(err, "package.json marshal indent")
	}

	// os.FileMode(0755) The owner can read, write, execute.
	// Everyone else can read and execute but not modify the file.
	err = os.WriteFile(filepath.Join(directory, PackageJsonFile), packageJSON, os.FileMode(0755))
	if err != nil {
		return errors.Wrap(err, "failed to save package.json")
	}

	return nil
}

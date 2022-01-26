package util

import (
	"fmt"
	"github.com/tenderly/tenderly-cli/typescript"
	"github.com/tenderly/tenderly-cli/userError"
	"os"
	"path/filepath"
)

func MustSaveTsConfig(directory string, config *typescript.TsConfig) {
	err := typescript.SaveTsConfig(directory, config)
	if err != nil {
		userError.LogErrorf(
			"unexpected error writing tsconfig.json",
			userError.NewUserError(err, "Unexpected error writing tsconfig.json."))
		os.Exit(1)
	}
}

func MustLoadTsConfig(directory string) *typescript.TsConfig {
	tsconfig, err := typescript.LoadTsConfig(directory)
	if err != nil {
		userError.LogErrorf("failed to load tsconfig.json: %s",
			userError.NewUserError(err, "Failed to load tsconfig.json."),
		)
		os.Exit(1)
	}

	return tsconfig
}

func TsConfigExists(directory string) bool {
	return ExistFile(filepath.Join(directory, typescript.TsConfigFile))
}

func IsFileTs(filePath string) bool {
	if filepath.Ext(filePath) != "" {
		return filepath.Ext(filePath) == typescript.TsFileExt
	}
	filePathWithTsExt := fmt.Sprintf("%s%s", filePath, typescript.TsFileExt)
	return ExistFile(filePathWithTsExt)
}

func MustSavePackageJSON(directory string, config *typescript.PackageJson) {
	err := typescript.SavePackageJson(directory, config)
	if err != nil {
		userError.LogErrorf(
			"unexpected error writing package.json",
			userError.NewUserError(err, "Unexpected error writing package.json."))
		os.Exit(1)
	}
}

func MustLoadPackageJSON(directory string) *typescript.PackageJson {
	packageJSON, err := typescript.LoadPackageJson(directory)
	if err != nil {
		userError.LogErrorf("failed to load package.json: %s",
			userError.NewUserError(err, "Failed to load package.json."),
		)
		os.Exit(1)
	}

	return packageJSON
}

func PackageJSONExists(directory string) bool {
	return ExistFile(filepath.Join(directory, typescript.PackageJsonFile))
}

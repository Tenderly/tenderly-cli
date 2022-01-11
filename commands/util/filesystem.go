package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/userError"
)

/*
	Provides utility methods for reading / writing files in project, failing with user friendly errors.
	Failures will call os.Exit - this is meant to be used only from commands.
*/

func ExistFile(path string) bool {
	path, _ = filepath.Abs(path)
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		userError.LogErrorf(
			"failed to read path: %s",
			userError.NewUserError(err,
				fmt.Sprintf("Couldn't read path at %s.", path)))
		os.Exit(1)
	}

	if info.IsDir() {
		return false
	}
	return true
}

func MustExistFile(path string) {
	if ExistFile(path) {
		return
	}
	userError.LogErrorf(
		"file expected but does not exist: %s",
		userError.NewUserError(errors.New("file expected but does not exist"),
			fmt.Sprintf("File expected at %s but does not exist.", path)))
	os.Exit(1)
}

func ExistDir(path string) bool {
	path, _ = filepath.Abs(path)
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		userError.LogErrorf(
			"failed to read path: %s",
			userError.NewUserError(err,
				fmt.Sprintf("Couldn't read path at %s.", path)))
		os.Exit(1)
	}

	if info.IsDir() {
		return true
	}
	return false
}

func MustExistDir(path string) {
	if ExistDir(path) {
		return
	}
	userError.LogErrorf(
		"directory expected but does not exist: %s",
		userError.NewUserError(errors.New("directory expected but does not exist"),
			fmt.Sprintf("Directory expected at %s but does not exist.", path)))
	os.Exit(1)
}

func ReadFile(path string) string {
	if !ExistFile(path) {
		userError.LogErrorf(
			"failed to read file: %s",
			userError.NewUserError(errors.New("file does not exist"),
				fmt.Sprintf("Couldn't read file at %s. File does not exist.", path)))
		os.Exit(1)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		userError.LogErrorf(
			"failed to read file: %s",
			userError.NewUserError(err,
				fmt.Sprintf("Couldn't read file at %s.", path)))
		os.Exit(1)
	}
	return string(content)
}

func CreateFile(path string) {
	CreateFileWithContent(path, "")
}

func CreateFileWithContent(path string, content string) {
	if ExistFile(path) {
		userError.LogErrorf(
			"failed to create file: %s",
			userError.NewUserError(errors.New("file already exists"),
				fmt.Sprintf("Couldn't create file at %s. File already exists.", path)))
		os.Exit(1)
	}
	err := os.WriteFile(
		path,
		[]byte(content),
		os.FileMode(0755),
	)
	if err != nil {
		userError.LogErrorf(
			"failed to create file: %s",
			userError.NewUserError(err,
				fmt.Sprintf("Couldn't create file at %s.", path)))
		os.Exit(1)
	}
}

func CreateDir(path string) {
	if ExistDir(path) {
		userError.LogErrorf(
			"failed to create dir: %s",
			userError.NewUserError(errors.New("directory already exists"),
				fmt.Sprintf("Couldn't create directory at %s. Directory already exists.", path)))
		os.Exit(1)
	}
	err := os.MkdirAll(path, os.FileMode(0755))
	if err != nil {
		userError.LogErrorf(
			"failed to create dir: %s",
			userError.NewUserError(err,
				fmt.Sprintf("Couldn't create directory at %s.", path)))
		os.Exit(1)
	}
}

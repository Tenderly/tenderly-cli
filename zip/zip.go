package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Zip returns paths added to zip & zip bytes or error. Inside path is path used inside zip, e.g. if out/test.txt
// exists and out/ dir is zipped with insidePath src/, zip will contains src/test.txt.
func Zip(dirPath string, insidePath string) ([]string, []byte, error) {
	buf := new(bytes.Buffer)
	writer := zip.NewWriter(buf)
	var files []string

	err := filepath.WalkDir(dirPath, func(path string, dirEntry os.DirEntry, err error) error {
		destPath := getDestPath(path, dirPath, insidePath)

		if dirEntry == nil {
			return fmt.Errorf("directory missing %s", dirPath)
		}
		if dirEntry.IsDir() {
			return nil
		}
		if dirEntry.Type() == os.ModeSymlink {
			err := walkSymlink(writer, &files, path, destPath)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to walk symlinked dir %s", dirPath))
			}

			return nil
		}
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to read path %s", path))
		}

		dat, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed reading %s", path))
		}

		destFile, err := writer.Create(destPath)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed creating %s", destPath))
		}
		_, err = destFile.Write(dat)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed writing %s", destPath))
		}
		files = append(files, path)

		return nil
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, fmt.Sprintf("walk directory %s", dirPath))
	}

	err = writer.Close()
	if err != nil {
		return nil, nil, errors.Wrap(err, "close writer")
	}

	return files, buf.Bytes(), nil
}

func walkSymlink(writer *zip.Writer, files *[]string, path string, destBasePath string) error {
	evaluatedDirPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to eval symlink %s", path))
	}

	return filepath.WalkDir(evaluatedDirPath, func(path string, dirEntry os.DirEntry, err error) error {
		if dirEntry == nil {
			return fmt.Errorf("directory missing %s", evaluatedDirPath)
		}
		if dirEntry.IsDir() {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to read path %s", path))
		}

		dat, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed reading %s", path))
		}

		destPath := getDestPath(path, evaluatedDirPath, destBasePath)
		destFile, err := writer.Create(destPath)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed creating %s", destPath))
		}
		_, err = destFile.Write(dat)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed writing %s", destPath))
		}
		*files = append(*files, path)

		return nil
	})
}

/*
Provides path that filePath will have in the output zip.
filePath - Represents full path to the file (e.g. `node_modules/x/y.js`
dirPath - Represents path to the directory which contains the file (e.g. `node_modules/x/`)
destBasePath - Represents desired base path that the file will have in output zip (e.g. `nodejs/node_modules/x`)

For the example above, the output will be `nodejs/node_modules/x/y.js`.
*/
func getDestPath(filePath string, dirPath string, destBasePath string) string {
	relativePath := strings.TrimPrefix(filePath, dirPath)
	return filepath.Join(destBasePath, relativePath)
}

package zip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
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

	err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if info == nil {
			return fmt.Errorf("directory missing %s", dirPath)
		}
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to read path %s", path))
		}

		dat, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed reading %s", path))
		}

		relativePath := strings.TrimPrefix(path, dirPath)
		destPath := filepath.Join(insidePath, relativePath)
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
		return nil, nil, fmt.Errorf("walk directory %s", dirPath)
	}

	err = writer.Close()
	if err != nil {
		return nil, nil, errors.Wrap(err, "close writer")
	}

	return files, buf.Bytes(), nil
}

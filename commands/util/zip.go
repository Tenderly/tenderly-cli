package util

import (
	"fmt"
	"os"

	"github.com/tenderly/tenderly-cli/userError"
	"github.com/tenderly/tenderly-cli/zip"
)

func MustZipDir(dirPath string, insidePath string, limitBytes int) []byte {
	MustExistDir(dirPath)

	_, content, err := zip.Zip(dirPath, insidePath)
	if err != nil {
		userError.LogErrorf("zip directory failed: %s",
			userError.NewUserError(
				err,
				fmt.Sprintf("Zip directory %s failed.", dirPath),
			),
		)
		os.Exit(1)
	}

	if len(content) > limitBytes {
		userError.LogErrorf(
			"zip file exceeds the maximum file-size",
			userError.NewUserError(err, fmt.Sprintf(
				"Zip file exceeds the maximum file-size.\n"+
					"The maximum size limit for sources / dependencies is %dMB zipped, got %dMB.",
				limitBytes/1024/1024,
				len(content)/1024/1024,
			)),
		)
		os.Exit(1)
	}

	return content
}

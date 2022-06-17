package util

import (
	"crypto/md5"
	"encoding/hex"
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
				fmt.Sprintf("Zip directory %s failed. Please run this command with the \"--debug\" flag and send the logs to our customer support.",
					dirPath,
				),
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

func MustZipAndHashDir(dirPath string, insidePath string, limitBytes int) ([]byte, string) {
	zipped := MustZipDir(dirPath, insidePath, limitBytes)

	hasher := md5.New()
	hasher.Write(zipped)
	hash := hex.EncodeToString(hasher.Sum(nil))

	return zipped, hash
}

func ZipAndHashDir(dirPath, insidePath string, limitBytes int) ([]byte, string) {
	if !ExistDir(dirPath) {
		return nil, ""
	}

	return MustZipAndHashDir(dirPath, insidePath, limitBytes)
}

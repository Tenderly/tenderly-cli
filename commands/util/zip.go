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

func MustZipAndHashDir(dirPath string, insidePath string, limitBytes int) (*[]byte, *string) {
	// TODO(slobodan): is zip util right place for this function, as it is also hashing?
	var dependenciesZip *[]byte
	dZip := MustZipDir(dirPath, insidePath, limitBytes)
	if len(dZip) > 0 {
		dependenciesZip = &dZip
	}
	hasher := md5.New()
	hasher.Write(*dependenciesZip)
	dependenciesVersion := hex.EncodeToString(hasher.Sum(nil))
	return dependenciesZip, &dependenciesVersion
}

func ZipAndHashDir(dirPath string, insidePath string, limitBytes int) (*[]byte, *string) {
	if !ExistDir(dirPath) {
		return nil, nil
	}
	return MustZipAndHashDir(dirPath, insidePath, limitBytes)
}

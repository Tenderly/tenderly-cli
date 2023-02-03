package providers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidProviderStructure(t *testing.T) {
	t.Parallel()

	var (
		randomFolders = []string{
			"folder_1",
			"folder_2",
		}
	)

	testTable := []struct {
		name               string
		initialDirectories []string
		directories        []string
		shouldHaveValid    bool
	}{
		{
			"Valid folder structure",
			randomFolders,
			randomFolders,
			true,
		},
		{
			"Invalid folder structure",
			[]string{},
			[]string{
				"build",
			},
			false,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Set up the initial folder structure
			tempDirectory, err := os.MkdirTemp("", "")
			if err != nil {
				t.Fatalf("unable to create temporary base directory, %v", err)
			}

			// Set up the cleanup method
			t.Cleanup(func() {
				_ = os.RemoveAll(tempDirectory)
			})

			for _, initialDir := range testCase.initialDirectories {
				path := filepath.Join(tempDirectory, initialDir)

				if err := os.MkdirAll(path, os.ModePerm); err != nil {
					t.Fatalf("unable to create temporary directory structure, %v", err)
				}
			}

			// Validate if the folder structure is present or not
			assert.Equal(
				t,
				testCase.shouldHaveValid,
				ValidProviderStructure(tempDirectory, testCase.directories),
			)
		})
	}
}

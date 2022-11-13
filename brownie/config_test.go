package brownie

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/providers"
)

// TestBrownie_MustGetConfig validates that a Brownie config is
// successfully read if present
func TestBrownie_MustGetConfig(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name       string
		config     string
		shouldRead bool
	}{
		{
			"Valid Brownie config",
			`
					project_structure:
						build: client/src/artifacts
					
					networks:
					  default: development
					  development:
						cmd_settings:
						  mnemonic: hill law jazz limb penalty escape public dish stand bracket blue jar
					
					dev_deployment_artifacts: true
				`,
			true,
		},
		{
			"Missing Brownie config",
			"",
			false,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			projectDirectory := ""
			if testCase.config != "" {
				// Create a temporary directory
				tmpDirectory, err := os.MkdirTemp("", "")
				if err != nil {
					t.Fatalf("unable to create temporary directory, %v", err)
				}

				t.Cleanup(func() {
					_ = os.RemoveAll(tmpDirectory)
				})

				projectDirectory = tmpDirectory

				// Create a temporary config file
				filePath := filepath.Join(tmpDirectory, providers.BrownieConfigFile)

				if err := os.WriteFile(filePath, []byte(testCase.config), os.ModePerm); err != nil {
					t.Fatalf("unable to write the temporary configuration file, %v", err)
				}
			}

			// Read the configuration file
			config.ProjectDirectory = projectDirectory
			brownieConfig, err := NewBrownieProvider().MustGetConfig()

			if testCase.shouldRead {
				// Make sure the configuration has been read
				assert.NotNil(t, brownieConfig)

				// Make sure no error is returned
				assert.NoError(t, err)
			} else {
				// Make sure no configuration has been read
				assert.Nil(t, brownieConfig)

				// Make sure an error is returned
				assert.Error(t, err)
			}
		})
	}
}

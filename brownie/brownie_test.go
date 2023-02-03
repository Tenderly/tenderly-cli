package brownie

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tenderly/tenderly-cli/providers"
)

// TestBrownie_GetProviderName validates that the correct
// Brownie provider name is returned
func TestBrownie_GetProviderName(t *testing.T) {
	t.Parallel()

	// Make sure the name output is valid
	assert.Equal(
		t,
		providers.BrownieDeploymentProvider,
		NewBrownieProvider().GetProviderName(),
	)
}

// TestBrownie_GetDirectoryStructure validates that the correct
// Brownie directory structure is returned
func TestBrownie_GetDirectoryStructure(t *testing.T) {
	t.Parallel()

	// Make sure the directory structure is valid
	assert.Equal(
		t,
		directoryStructure,
		NewBrownieProvider().GetDirectoryStructure(),
	)
}

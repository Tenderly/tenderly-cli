package brownie

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tenderly/tenderly-cli/providers"
)

// generateContracts generates a specified number of provider contracts
func generateContracts(count int) []*providers.Contract {
	contracts := make([]*providers.Contract, count)

	for i := 0; i < count; i++ {
		contracts[i] = &providers.Contract{
			Name: fmt.Sprintf("Contract %d", i),
		}
	}

	return contracts
}

// initContractsMap initializes the contracts map using the specified
// provider contracts
func initContractsMap(contracts []*providers.Contract) contractsMap {
	contractsMap := make(contractsMap)
	for _, contract := range contracts {
		contractsMap[contract.Name] = contract
	}

	return contractsMap
}

// TestBrownie_NetworkIDFilter verifies that
// network IDs are correctly added
func TestBrownie_NetworkIDFilter(t *testing.T) {
	t.Parallel()

	var (
		initialIDs = []string{"1", "2", "3"}
	)

	f := newNetworkIDFilterMap(initialIDs)

	for _, id := range initialIDs {
		assert.True(t, f.hasNetworkID(id))
	}
}

// TestBrownie_ContractsMap verifies that the
// contracts maps functionality is valid
func TestBrownie_ContractsMap(t *testing.T) {
	t.Parallel()

	// Generate random contracts
	var (
		contracts = generateContracts(10)
		halfSize  = len(contracts) / 2
	)

	// Add them to two separate maps
	firstMap := initContractsMap(contracts[:halfSize])
	assert.Len(t, firstMap, halfSize)

	secondMap := initContractsMap(contracts[halfSize:])
	assert.Len(t, secondMap, len(contracts)-halfSize)

	// Merge the maps
	mergedMap := mergeContractsMaps(firstMap, secondMap)

	// Make sure the merged map has good values
	assert.NotNil(t, mergedMap)
	assert.Len(t, mergedMap, len(contracts))

	// Make sure all contracts are present
	for _, contract := range contracts {
		c, exists := mergedMap[contract.Name]
		if !exists {
			t.Fatalf("contract not found in merged map")
		}

		assert.Equal(t, contract, c)
	}

	mapContracts := mergedMap.getContracts()

	assert.Len(t, mapContracts, len(contracts))
}

// TestBrownie_FilterContracts verifies that the contracts map
// filtering works correctly
func TestBrownie_FilterContracts(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name               string
		contractsMap       contractsMap
		deploymentsMap     deploymentsMap
		networkIDFilterMap networkIDFilterMap

		expectedNumberOfContracts int
	}{
		{
			"Non-matching number of contracts",
			initContractsMap(generateContracts(0)),
			make(deploymentsMap),
			newNetworkIDFilterMap([]string{}),
			0,
		},
		{
			"Matching number of contracts",
			initContractsMap(generateContracts(5)),
			deploymentsMap{
				"Network 2": map[string][]string{
					"Contract 0": {"0x0"},
					"Contract 1": {"0x1"},
				},
				"Network 4": map[string][]string{
					"Contract 2": {"0x2"},
				},
			},
			newNetworkIDFilterMap([]string{
				"Network 0",
				"Network 2",
				"Network 4",
			}),
			3,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				testCase.expectedNumberOfContracts,
				testCase.contractsMap.filterContracts(
					testCase.deploymentsMap,
					testCase.networkIDFilterMap,
				),
			)
		})
	}
}

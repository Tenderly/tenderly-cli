package extensions_test

import (
	"github.com/tenderly/tenderly-cli/commands/extensions"
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	extensionsModel "github.com/tenderly/tenderly-cli/model/extensions"
	gatewaysModel "github.com/tenderly/tenderly-cli/model/gateways"
	"testing"
)

func TestProjectData_FindActionByID(t *testing.T) {
	searchActionID := "47790e62-6d15-4a1d-b3aa-b6276fc7c849"

	testActions := []actionsModel.Action{
		{
			ID:   searchActionID,
			Name: "action-1",
		},
		{
			ID:   "c9676cb4-6b50-4501-882e-63c0aeef5fa1",
			Name: "action-2",
		},
	}
	testProjectData := extensions.NewProjectData(
		"accountSlug",
		"projectSlug",
		&gatewaysModel.Gateway{},
		testActions,
		[]extensionsModel.BackendExtension{},
	)

	t.Run("should return nil if actions are nil", func(t *testing.T) {
		testProjectData := extensions.NewProjectData(
			"accountSlug",
			"projectSlug",
			&gatewaysModel.Gateway{},
			nil,
			[]extensionsModel.BackendExtension{},
		)
		action := testProjectData.FindActionByID(searchActionID)
		if action != nil {
			t.Errorf("expected action to be nil")
		}
	})

	t.Run("should return action with given ID", func(t *testing.T) {
		action := testProjectData.FindActionByID(searchActionID)
		if action == nil {
			t.Errorf("expected action to be found")
		}
		if action.ID != searchActionID {
			t.Errorf("expected action to have ID %s, got %s", searchActionID, action.ID)
		}
	})

	t.Run("should return nil if action with given ID is not found", func(t *testing.T) {
		action := testProjectData.FindActionByID("not-found")
		if action != nil {
			t.Errorf("expected action to be nil")
		}
	})

}

func TestProjectData_FindActionByName(t *testing.T) {
	searchActionName := "action-1"

	testActions := []actionsModel.Action{
		{
			ID:   "47790e62-6d15-4a1d-b3aa-b6276fc7c849",
			Name: searchActionName,
		},
		{
			ID:   "c9676cb4-6b50-4501-882e-63c0aeef5fa1",
			Name: "action-2",
		},
	}
	testProjectData := extensions.NewProjectData(
		"accountSlug",
		"projectSlug",
		&gatewaysModel.Gateway{},
		testActions,
		[]extensionsModel.BackendExtension{},
	)

	t.Run("should return action with given name", func(t *testing.T) {
		action := testProjectData.FindActionByName(searchActionName)
		if action == nil {
			t.Errorf("expected action to be found")
		}
		if action.Name != searchActionName {
			t.Errorf("expected action to have name %s, got %s", searchActionName, action.Name)
		}
	})

	t.Run("should return nil if action with given name is not found", func(t *testing.T) {
		action := testProjectData.FindActionByName("not-found")
		if action != nil {
			t.Errorf("expected action to be nil")
		}
	})
}

func TestProjectData_FindExtensionByName(t *testing.T) {
	searchExtensionName := "extension-1"

	testExtensions := []extensionsModel.BackendExtension{
		{
			Name:     searchExtensionName,
			Method:   "extension_methodName1",
			ActionID: "47790e62-6d15-4a1d-b3aa-b6276fc7c849",
		},
		{
			Name:     "extension-2",
			Method:   "extension_methodName2",
			ActionID: "c9676cb4-6b50-4501-882e-63c0aeef5fa1",
		},
	}
	testProjectData := extensions.NewProjectData(
		"accountSlug",
		"projectSlug",
		&gatewaysModel.Gateway{},
		[]actionsModel.Action{},
		testExtensions,
	)

	t.Run("should return extension with given name", func(t *testing.T) {
		extension := testProjectData.FindExtensionByName(searchExtensionName)
		if extension == nil {
			t.Errorf("expected extension to be found")
		}
		if extension.Name != searchExtensionName {
			t.Errorf("expected extension to have name %s, got %s", searchExtensionName, extension.Name)
		}
	})

	t.Run("should return nil if extension with given name is not found", func(t *testing.T) {
		extension := testProjectData.FindExtensionByName("not-found")
		if extension != nil {
			t.Errorf("expected extension to be nil")
		}
	})
}

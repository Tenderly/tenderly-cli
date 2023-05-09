package extensions

import (
	actionsModel "github.com/tenderly/tenderly-cli/model/actions"
	extensionsModel "github.com/tenderly/tenderly-cli/model/extensions"
	gatewaysModel "github.com/tenderly/tenderly-cli/model/gateways"
	"testing"
)

func TestDeploymentTask_Validate(t *testing.T) {
	projectData := NewProjectData(
		"accountSlug",
		"projectSlug",
		&gatewaysModel.Gateway{
			ID:        "61a69c43-1a70-4ac1-ab63-b07806761995",
			Name:      "",
			AccessKey: "4d093d81-7fc1-456f-836d-2ce89afb9b1b",
		},
		[]actionsModel.Action{
			{
				ID:   "47790e62-6d15-4a1d-b3aa-b6276fc7c849",
				Name: "action-1",
			},
			{
				ID:   "c9676cb4-6b50-4501-882e-63c0aeef5fa1",
				Name: "action-2",
			},
			{
				ID:   "53b2a144-3f49-4021-9c09-b44136b06d34",
				Name: "action-3",
			},
		},
		[]extensionsModel.BackendExtension{
			{
				Name:     "extension-1",
				Method:   "extension_methodName1",
				ActionID: "47790e62-6d15-4a1d-b3aa-b6276fc7c849",
			},
			{
				Name:     "extension-2",
				Method:   "extension_methodName2",
				ActionID: "c9676cb4-6b50-4501-882e-63c0aeef5fa1",
			},
		},
	)

	t.Run("should return success if extension is valid", func(t *testing.T) {
		extension := extensionsModel.ConfigExtension{
			Name:        "extension-3",
			MethodName:  "extension_methodName3",
			Description: "extension_description3",
			ActionName:  "action-3",
		}

		task := deploymentTask{
			ProjectData: projectData,
			Extension:   extension,
		}

		result := task.validate()
		if !result.Success {
			t.Errorf("expected result to be successful")
		}
	})

	t.Run("should return error if extension method name is already used", func(t *testing.T) {
		extension := extensionsModel.ConfigExtension{
			Name:        "extension-2",
			MethodName:  "extension_methodName2",
			Description: "extension_description3",
			ActionName:  "action-3",
		}

		task := deploymentTask{
			ProjectData: projectData,
			Extension:   extension,
		}

		result := task.validate()
		if result.Success {
			t.Errorf("expected result to be unsuccessful")
		} else if len(result.FailureSlugs) != 1 {
			t.Errorf("expected result to have 1 validation error")
		} else if result.FailureSlugs[0] != methodNameInUseSlug {
			t.Errorf("expected result to have %s validation error", methodNameInUseSlug)
		}
	})

	t.Run("should return error if action is already used", func(t *testing.T) {
		extension := extensionsModel.ConfigExtension{
			Name:        "extension-3",
			MethodName:  "extension_methodName3",
			Description: "extension_description3",
			ActionName:  "action-2",
		}

		task := deploymentTask{
			ProjectData: projectData,
			Extension:   extension,
		}

		result := task.validate()
		if result.Success {
			t.Errorf("expected result to be unsuccessful")
		} else if len(result.FailureSlugs) != 1 {
			t.Errorf("expected result to have 1 validation error")
		} else if result.FailureSlugs[0] != actionIsInUseSlug {
			t.Errorf("expected result to have %s validation error", actionIsInUseSlug)
		}
	})

	t.Run("should return error if extension name is invalid", func(t *testing.T) {
		invalidMethodNames := []string{
			"extension name",
			"extension-name",
			"extension_name_",
			"extension_1",
			"extension-1",
			"extension-",
			"extension_",
			"extension-1-",
			"prefix_name",
		}
		extension := extensionsModel.ConfigExtension{
			Name:        "extension-3",
			MethodName:  "extension_methodName3",
			Description: "extension_description3",
			ActionName:  "action-3",
		}
		task := deploymentTask{
			ProjectData: projectData,
		}

		for _, invalidMethodName := range invalidMethodNames {
			extension.MethodName = invalidMethodName
			task.Extension = extension

			result := task.validate()
			if result.Success {
				t.Errorf("expected result to be unsuccessful")
			} else if len(result.FailureSlugs) != 1 {
				t.Errorf("expected result to have 1 validation error")
			} else if result.FailureSlugs[0] != invalidMethodNameSlug {
				t.Errorf("expected result to have %s validation error", invalidMethodNameSlug)
			}
		}
	})

	t.Run("should return error if action does not exist", func(t *testing.T) {
		extension := extensionsModel.ConfigExtension{
			Name:        "extension-3",
			MethodName:  "extension_methodName3",
			Description: "extension_description3",
			ActionName:  "action-4",
		}

		task := deploymentTask{
			ProjectData: projectData,
			Extension:   extension,
		}

		result := task.validate()
		if result.Success {
			t.Errorf("expected result to be unsuccessful")
		} else if len(result.FailureSlugs) != 1 {
			t.Errorf("expected result to have 1 validation error")
		} else if result.FailureSlugs[0] != actionDoesNotExistSlug {
			t.Errorf("expected result to have %s validation error", actionDoesNotExistSlug)
		}
	})

	t.Run("should return multiple errors if extension is invalid", func(t *testing.T) {
		extension := extensionsModel.ConfigExtension{
			Name:        "extension-2",
			MethodName:  "extension_methodName2",
			Description: "extension_description3",
			ActionName:  "action-2",
		}

		task := deploymentTask{
			ProjectData: projectData,
			Extension:   extension,
		}

		result := task.validate()
		if result.Success {
			t.Errorf("expected result to be unsuccessful")
		} else if len(result.FailureSlugs) != 2 {
			t.Errorf("expected result to have 2 validation errors")
		} else {
			expectedSlugs := map[validationFailureSlug]bool{
				methodNameInUseSlug: false,
				actionIsInUseSlug:   false,
			}
			for _, slug := range result.FailureSlugs {
				if !expectedSlugs[slug] {
					expectedSlugs[slug] = true
				}
			}
			if !expectedSlugs[methodNameInUseSlug] ||
				!expectedSlugs[actionIsInUseSlug] {
				t.Errorf("expected result to have %s and %s validation errors", methodNameInUseSlug, actionIsInUseSlug)
			}
		}
	})
}

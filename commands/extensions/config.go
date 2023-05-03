package extensions

import (
	"github.com/tenderly/tenderly-cli/config"
	extensionsModel "github.com/tenderly/tenderly-cli/model/extensions"
	"github.com/tenderly/tenderly-cli/userError"
	"gopkg.in/yaml.v3"
	"os"
)

func ReadExtensionsFromConfig() map[string][]extensionsModel.ConfigExtension {
	extensions := make(map[string][]extensionsModel.ConfigExtension)
	allExtensions := MustGetExtensions()
	for accountAndProjectSlug, projectExtensions := range allExtensions {
		extensions[accountAndProjectSlug] = make([]extensionsModel.ConfigExtension, len(projectExtensions.Specs))
		i := 0
		for configExtensionName, configExtension := range projectExtensions.Specs {
			extensions[accountAndProjectSlug][i] = extensionsModel.ConfigExtension{
				Name:        configExtensionName,
				ActionName:  configExtension.ActionName,
				MethodName:  configExtension.MethodName,
				Description: configExtension.Description,
			}
			i++
		}
	}

	return extensions
}

type extensionsTenderlyYaml struct {
	Extensions map[string]extensionsModel.ConfigProjectExtensions `yaml:"node_extensions"`
}

func MustGetExtensions() map[string]extensionsModel.ConfigProjectExtensions {
	content, err := config.ReadProjectConfig()
	if err != nil {
		userError.LogErrorf("failed reading project config: %s",
			userError.NewUserError(
				err,
				"Failed reading project's tenderly.yaml config. This can happen if you are running an older version of the Tenderly CLI.",
			),
		)
		os.Exit(1)
	}

	var tenderlyYaml extensionsTenderlyYaml
	err = yaml.Unmarshal(content, &tenderlyYaml)
	if err != nil {
		userError.LogErrorf("failed unmarshalling `node_extensions` config: %s",
			userError.NewUserError(
				err,
				"Failed parsing `node_extensions` configuration. This can happen if you are running an older version of the Tenderly CLI.",
			),
		)
		os.Exit(1)
	}

	return tenderlyYaml.Extensions
}

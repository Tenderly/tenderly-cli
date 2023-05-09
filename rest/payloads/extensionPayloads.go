package payloads

import "github.com/tenderly/tenderly-cli/model/extensions"

type DeployExtensionResponse struct{}

type GetExtensionsResponse struct {
	Handlers []extensions.BackendExtension
}

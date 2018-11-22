package truffle

import "path/filepath"

type NetworkConfig struct {
	Host      string      `json:"host"`
	Port      int         `json:"port"`
	NetworkID interface{} `json:"network_id"`
}

type Config struct {
	ProjectDirectory string                   `json:"project_directory"`
	BuildDirectory   string                   `json:"contracts_build_directory"`
	Networks         map[string]NetworkConfig `json:"networks"`
}

func (c *Config) AbsoluteBuildDirectoryPath() string {
	return filepath.Join(c.ProjectDirectory, c.BuildDirectory)
}

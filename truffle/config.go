package truffle

import "path/filepath"

type NetworkConfig struct {
	Host      string      `json:"host"`
	Port      int         `json:"port"`
	NetworkID interface{} `json:"network_id"`
}

type Compiler struct {
	Version  string            `json:"version"`
	Settings *CompilerSettings `json:"settings"`
}

type CompilerSettings struct {
	Optimizer *Optimizer `json:"optimizer"`
}

type Optimizer struct {
	Enabled *bool `json:"enabled"`
	Runs    *int  `json:"runs"`
}

type Config struct {
	ProjectDirectory string                   `json:"project_directory"`
	BuildDirectory   string                   `json:"contracts_build_directory"`
	Networks         map[string]NetworkConfig `json:"networks"`
	Solc             map[string]Optimizer     `json:"solc"`
	Compilers        map[string]Compiler      `json:"compilers"`
}

func (c *Config) AbsoluteBuildDirectoryPath() string {
	if c.BuildDirectory == "" {
		c.BuildDirectory = filepath.Join(".", "build", "contracts")
	}

	switch c.BuildDirectory[0] {
	case '.':
		return filepath.Join(c.ProjectDirectory, c.BuildDirectory)
	default:
		return c.BuildDirectory
	}
}

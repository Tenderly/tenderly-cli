package typescript

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/flynn/json5"
	"github.com/pkg/errors"
)

type TsConfig struct {
	CompilerOptions CompilerOptions        `json:"compilerOptions,omitempty"`
	CompileOnSave   *bool                  `json:"compileOnSave,omitempty"`
	Include         []string               `json:"include,omitempty"`
	Exclude         []string               `json:"exclude,omitempty"`
	Detailed        map[string]interface{} `json:"-"` // Rest of the fields should go here.
}

type CompilerOptions struct {
	Target            *string `json:"target,omitempty"`
	Module            *string `json:"module,omitempty"`
	OutDir            *string `json:"outDir,omitempty"`
	SourceMap         *bool   `json:"sourceMap,omitempty"`
	Strict            *bool   `json:"strict,omitempty"`
	NoImplicitReturns *bool   `json:"noImplicitReturns,omitempty"`
	NoUnusedLocals    *bool   `json:"noUnusedLocals,omitempty"`
}

func boolPointer(b bool) *bool {
	return &b
}

func stringPointer(str string) *string {
	return &str
}

func DefaultTsConfig() *TsConfig {
	return &TsConfig{
		CompilerOptions: CompilerOptions{
			Target:            stringPointer("es2020"),
			Module:            stringPointer("commonjs"),
			OutDir:            stringPointer("out"),
			SourceMap:         boolPointer(true),
			Strict:            boolPointer(true),
			NoImplicitReturns: boolPointer(true),
			NoUnusedLocals:    boolPointer(true),
		},
		CompileOnSave: boolPointer(true),
		Include:       []string{"**/*"},
	}
}

func LoadTsConfig(directory string) (*TsConfig, error) {
	path := filepath.Join(directory, TsConfigFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read tsconfig")
	}

	var value TsConfig
	// JSON5 is used for tsconfig
	err = json5.Unmarshal(data, &value)
	if err != nil {
		return nil, errors.Wrap(err, "parse tsconfig")
	}

	err = json5.Unmarshal(data, &value.Detailed)
	if err != nil {
		return nil, errors.Wrap(err, "parse tsconfig")
	}

	// Remove all keys that we loaded in struct
	delete(value.Detailed, "compileOnSave")
	delete(value.Detailed, "include")
	delete(value.Detailed, "exclude")

	if value.Detailed["compilerOptions"] == nil {
		return &value, nil
	}
	compilerOptions := value.Detailed["compilerOptions"].(map[string]interface{})
	delete(compilerOptions, "target")
	delete(compilerOptions, "module")
	delete(compilerOptions, "outDir")
	delete(compilerOptions, "sourceMap")
	delete(compilerOptions, "strict")
	delete(compilerOptions, "noImplicitReturns")
	delete(compilerOptions, "noUnusedLocals")

	return &value, nil
}

func SaveTsConfig(directory string, config *TsConfig) error {
	// Remap all values from struct beck to generic map[string]interface{}
	newConfig := make(map[string]interface{})
	for key, value := range config.Detailed {
		newConfig[key] = value
	}

	if config.CompileOnSave != nil {
		newConfig["compileOnSave"] = config.CompileOnSave
	}
	if config.Include != nil {
		newConfig["include"] = config.Include
	}
	if config.Exclude != nil {
		newConfig["exclude"] = config.Exclude
	}

	if newConfig["compilerOptions"] == nil {
		newConfig["compilerOptions"] = make(map[string]interface{})
	}
	compilerOptions := newConfig["compilerOptions"].(map[string]interface{})
	if config.CompilerOptions.Target != nil {
		compilerOptions["target"] = config.CompilerOptions.Target
	}
	if config.CompilerOptions.Module != nil {
		compilerOptions["module"] = config.CompilerOptions.Module
	}
	if config.CompilerOptions.OutDir != nil {
		compilerOptions["outDir"] = config.CompilerOptions.OutDir
	}
	if config.CompilerOptions.SourceMap != nil {
		compilerOptions["sourceMap"] = config.CompilerOptions.SourceMap
	}
	if config.CompilerOptions.Strict != nil {
		compilerOptions["strict"] = config.CompilerOptions.Strict
	}
	if config.CompilerOptions.NoImplicitReturns != nil {
		compilerOptions["noImplicitReturns"] = config.CompilerOptions.NoImplicitReturns
	}
	if config.CompilerOptions.NoUnusedLocals != nil {
		compilerOptions["noUnusedLocals"] = config.CompilerOptions.NoUnusedLocals
	}
	if len(compilerOptions) == 0 {
		delete(newConfig, "compilerOptions")
	}

	tsconfig, err := json.MarshalIndent(newConfig, "", "    ")
	if err != nil {
		return errors.Wrap(err, "tsconfig marshal indent")
	}

	// os.FileMode(0755) The owner can read, write, execute.
	// Everyone else can read and execute but not modify the file.
	err = os.WriteFile(filepath.Join(directory, TsConfigFile), tsconfig, os.FileMode(0755))
	if err != nil {
		return errors.Wrap(err, "failed to save tsconfig")
	}

	return nil
}

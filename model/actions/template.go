package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v37/github"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	repositoryOwner = "Tenderly"
	repository      = "tenderly-actions"
)

type Template struct {
	Files []string      `yaml:"files"`
	Args  []TemplateArg `yaml:"args"`

	path string `yaml:"-"`
}

type TemplateArg struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func (t *Template) Create(ctx context.Context, destinationDir string, args map[string]string) error {
	client := github.NewClient(nil)

	for _, file := range t.Files {
		filePath := t.path + "/" + file
		content, err := getFileContent(ctx, client, filePath)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("load template file %s", filePath))
		}

		for k, v := range args {
			content = strings.ReplaceAll(content, fmt.Sprintf("$%s", k), v)
		}

		destinationPath := filepath.Join(destinationDir, file)
		err = os.WriteFile(
			destinationPath,
			[]byte(content),
			os.FileMode(0755),
		)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("write template file %s", destinationPath))
		}
	}

	return nil
}

func (t *Template) LoadSpecs(ctx context.Context, args map[string]string) (map[string]*ActionSpec, error) {
	client := github.NewClient(nil)

	specsPath := t.path + "/specs.yaml"
	content, err := getFileContent(ctx, client, specsPath)
	if err != nil {
		return nil, errors.Wrap(err, "load specs content")
	}

	for k, v := range args {
		content = strings.ReplaceAll(content, fmt.Sprintf("$%s", k), v)
	}

	specs := make(map[string]*ActionSpec)
	err = yaml.Unmarshal([]byte(content), &specs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal template specs")
	}

	return specs, nil
}

func LoadTemplate(ctx context.Context, templateName string) (*Template, error) {
	client := github.NewClient(nil)

	templatePath := fmt.Sprintf("templates/%s", templateName)
	templateConfigPath := fmt.Sprintf("templates/%s/template.yaml", templateName)

	content, err := getFileContent(ctx, client, templateConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get template config")
	}

	config := Template{
		path: templatePath,
	}
	err = yaml.Unmarshal([]byte(content), &config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal template config")
	}
	return &config, nil
}

func getFileContent(ctx context.Context, client *github.Client, path string) (string, error) {
	fileContent, _, _, err := client.Repositories.GetContents(ctx, repositoryOwner, repository, path, &github.RepositoryContentGetOptions{})
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to get file %s", path))
	}
	if fileContent == nil {
		return "", fmt.Errorf("file %s not found", path)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to decode file %s", path))
	}

	return content, nil
}

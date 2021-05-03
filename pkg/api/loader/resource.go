package loader

import (
	"fmt"
	"io"
	"os"

	cdresources "github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

func ResourcesFromFile(resourcesFilePath string) ([]cdresources.ResourceOptions, error) {
	file, err := os.Open(resourcesFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	resources, err := readResources(file)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func readResources(reader *os.File) ([]cdresources.ResourceOptions, error) {
	resources := make([]cdresources.ResourceOptions, 0)
	yamldecoder := yamlutil.NewYAMLOrJSONDecoder(reader, 1024)
	for {
		resource := cdresources.ResourceOptions{}
		if err := yamldecoder.Decode(&resource); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("unable to decode resource: %w", err)
		}

		if resource.Input != nil && resource.Access != nil {
			return nil, fmt.Errorf("the resources %q input and access is defind. Only one option is allowed", resource.Name)
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

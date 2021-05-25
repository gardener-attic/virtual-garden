package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"gopkg.in/yaml.v2"

)

func GetImageFromCompDescr(ctx context.Context, resourceName string) (string, error) {
	cd, err := getCompDescr(ctx)
	if err != nil {
		return "", err
	}

	for i := range cd.Resources {
		resource := &cd.Resources[i]
		if resource.Name == resourceName {
			ociRegistryAccess := cdv2.OCIRegistryAccess{}
			if err := json.Unmarshal(resource.Access.Raw, &ociRegistryAccess); err != nil {
				return "", err
			}
			return ociRegistryAccess.ImageReference, nil
		}
	}

	return "", fmt.Errorf("No resource for image %s found", resourceName)
}

func getCompDescr(ctx context.Context) (*cdv2.ComponentDescriptor, error) {
	cdPath := os.Getenv("COMPONENT_DESCRIPTOR_PATH")

	data, err := ioutil.ReadFile(cdPath)
	if err != nil {
		return nil, err
	}

	var cd *cdv2.ComponentDescriptor
	if err := yaml.Unmarshal(data, &cd); err != nil {
		return nil, err
	}

	return cd, nil
}


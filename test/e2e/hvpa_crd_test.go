package e2e_test

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/gardener/virtual-garden/pkg/virtualgarden"
)

//go:embed resources/hvpa.yaml
var hvpaCrd []byte

// deployHVPACRD deploys the HVPA CRD.
func deployHVPACRD(ctx context.Context, c client.Client) error {
	newCrd, err := loadHVPACRD()
	if err != nil {
		return err
	}

	crd := virtualgarden.EmptyHVPACRD()
	_, err = controllerutil.CreateOrUpdate(ctx, c, crd, func() error {
		crd.Spec = newCrd.Spec
		return nil
	})

	return err
}

// deleteHPVACRD deletes the HPVA CRD.
func deleteHPVACRD(ctx context.Context, c client.Client) error {
	return client.IgnoreNotFound(c.Delete(ctx, virtualgarden.EmptyHVPACRD()))
}

// loadHVPACRD loads the HVPA CRD from file resources/hvpa.yaml.
func loadHVPACRD() (*v1beta1.CustomResourceDefinition, error) {
	crd := &v1beta1.CustomResourceDefinition{}
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(hvpaCrd), 32)
	err := decoder.Decode(crd)
	if err != nil {
		return nil, fmt.Errorf("failed to decode HVPA CRD: %w", err)
	}

	return crd, nil
}

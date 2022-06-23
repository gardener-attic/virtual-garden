package virtualgarden

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"time"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//go:embed resources/hvpa.yaml
var hvpaCrd []byte

//go:embed resources/vpa.yaml
var vpaCrd []byte

// DeployHVPACRD deploys the HVPA CRD.
func DeployHVPACRD(ctx context.Context, c client.Client) error {
	newCrd, err := loadCRD(hvpaCrd)
	if err != nil {
		return err
	}

	crd := emptyHVPACRD()
	_, err = controllerutil.CreateOrUpdate(ctx, c, crd, func() error {
		crd.Spec = newCrd.Spec
		return nil
	})

	if err != nil {
		return err
	}

	err = waitToBeEstablished(ctx, c, crd)

	return err
}

// DeployVPACRD deploys the VPA CRD.
func DeployVPACRD(ctx context.Context, c client.Client) error {
	newCrd, err := loadCRD(vpaCrd)
	if err != nil {
		return err
	}

	crd := emptyVPACRD()
	_, err = controllerutil.CreateOrUpdate(ctx, c, crd, func() error {
		crd.Spec = newCrd.Spec
		return nil
	})

	if err != nil {
		return err
	}

	err = waitToBeEstablished(ctx, c, crd)

	return err
}

func waitToBeEstablished(ctx context.Context, c client.Client, crd *v1.CustomResourceDefinition) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	err := wait.PollImmediateUntil(2*time.Second, func() (done bool, err error) {
		if err := c.Get(ctx, client.ObjectKeyFromObject(crd), crd); err != nil {
			if client.IgnoreNotFound(err) != nil {
				return false, err
			}
			return false, nil
		}

		conditions := crd.Status.Conditions
		for _, condition := range conditions {
			if condition.Type == v1.Established && condition.Status == v1.ConditionTrue {
				return true, nil
			}
		}

		return false, nil
	}, timeoutCtx.Done())

	return err
}

// DeleteHPVACRD deletes the HPVA CRD.
func DeleteHPVACRD(ctx context.Context, c client.Client) error {
	return client.IgnoreNotFound(c.Delete(ctx, emptyHVPACRD()))
}

// loadCRD loads the CRD.
func loadCRD(crdBytes []byte) (*v1.CustomResourceDefinition, error) {
	crd := &v1.CustomResourceDefinition{}
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(crdBytes), 32)
	err := decoder.Decode(crd)
	if err != nil {
		return nil, fmt.Errorf("failed to decode CRD: %w", err)
	}

	return crd, nil
}

func emptyVPACRD() *v1.CustomResourceDefinition {
	return &v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "verticalpodautoscalers.autoscaling.k8s.io",
		},
	}
}

// Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package virtualgarden

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gardener/virtual-garden/pkg/api"
	mockclient "github.com/gardener/virtual-garden/pkg/mock/controller-runtime/client"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("KubeApiserverService", func() {
	var (
		ctrl *gomock.Controller
		c    *mockclient.MockClient

		ctx       = context.TODO()
		namespace = "foo"
		svc       = &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: KubeAPIServerServiceName, Namespace: namespace}}
		fakeErr   = fmt.Errorf("fail")

		hostname1       = "foo.com"
		hostname2       = "bar.com"
		dnsClass        = "baz"
		ttl       int32 = 42
		imports         = &api.Imports{
			VirtualGarden: api.VirtualGarden{
				KubeAPIServer: &api.KubeAPIServer{
					Exposure: &api.KubeAPIServerExposure{
						SNI: &api.SNI{
							Hostnames: []string{hostname1, hostname2},
							DNSClass:  pointer.StringPtr(dnsClass),
							TTL:       pointer.Int32Ptr(ttl),
						},
					},
				},
			},
		}

		op *operation
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		c = mockclient.NewMockClient(ctrl)

		op = &operation{client: c, namespace: namespace, imports: &api.Imports{}}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#DeployKubeAPIServerService", func() {
		It("should return the error in case GET fails", func() {
			c.
				EXPECT().
				Get(ctx, client.ObjectKey{Name: KubeAPIServerServiceName, Namespace: namespace}, gomock.AssignableToTypeOf(&corev1.Service{})).
				Return(fakeErr)

			Expect(op.DeployKubeAPIServerService(ctx)).To(MatchError(fakeErr))
		})

		Context("CREATE", func() {
			It("should correctly create the service object without SNI", func() {
				service := svc.DeepCopy()
				service.Labels = map[string]string{
					"app":       "virtual-garden",
					"component": "kube-apiserver",
				}
				setExpectedServiceFields(service)

				gomock.InOrder(
					c.
						EXPECT().
						Get(ctx, client.ObjectKey{Name: KubeAPIServerServiceName, Namespace: namespace}, gomock.AssignableToTypeOf(&corev1.Service{})).
						Return(apierrors.NewNotFound(schema.GroupResource{}, KubeAPIServerServiceName)),
					c.
						EXPECT().
						Create(ctx, service),
				)

				Expect(op.DeployKubeAPIServerService(ctx)).To(Succeed())
			})

			It("should correctly create the service object with SNI", func() {
				op = &operation{client: c, namespace: namespace, imports: imports}

				service := svc.DeepCopy()
				service.Labels = map[string]string{
					"app":       "virtual-garden",
					"component": "kube-apiserver",
				}
				service.Annotations = map[string]string{
					"dns.gardener.cloud/dnsnames": hostname1 + "," + hostname2,
					"dns.gardener.cloud/class":    dnsClass,
					"dns.gardener.cloud/ttl":      strconv.Itoa(int(ttl)),
				}
				setExpectedServiceFields(service)

				gomock.InOrder(
					c.
						EXPECT().
						Get(ctx, client.ObjectKey{Name: KubeAPIServerServiceName, Namespace: namespace}, gomock.AssignableToTypeOf(&corev1.Service{})).
						Return(apierrors.NewNotFound(schema.GroupResource{}, KubeAPIServerServiceName)),
					c.
						EXPECT().
						Create(ctx, service),
				)

				Expect(op.DeployKubeAPIServerService(ctx)).To(Succeed())
			})

			It("should return the error in case CREATE fails", func() {
				service := svc.DeepCopy()
				service.Labels = map[string]string{
					"app":       "virtual-garden",
					"component": "kube-apiserver",
				}
				setExpectedServiceFields(service)

				gomock.InOrder(
					c.
						EXPECT().
						Get(ctx, client.ObjectKey{Name: KubeAPIServerServiceName, Namespace: namespace}, gomock.AssignableToTypeOf(&corev1.Service{})).
						Return(apierrors.NewNotFound(schema.GroupResource{}, KubeAPIServerServiceName)),
					c.
						EXPECT().
						Create(ctx, service).
						Return(fakeErr),
				)

				Expect(op.DeployKubeAPIServerService(ctx)).To(MatchError(fakeErr))
			})
		})

		Context("UPDATE", func() {
			var nodePort int32 = 30123

			It("should correctly update the service object without SNI", func() {
				service := svc.DeepCopy()
				service.Annotations = map[string]string{}
				service.Labels = map[string]string{
					"foo":       "bar",
					"app":       "virtual-garden",
					"component": "kube-apiserver",
				}
				setExpectedServiceFields(service)
				service.Spec.Ports[0].NodePort = nodePort

				gomock.InOrder(
					c.
						EXPECT().
						Get(ctx, client.ObjectKey{Name: KubeAPIServerServiceName, Namespace: namespace}, gomock.AssignableToTypeOf(&corev1.Service{})).
						DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
							existingService := svc.DeepCopy()
							existingService.Annotations = map[string]string{
								"dns.gardener.cloud/dnsnames": "foo",
								"dns.gardener.cloud/class":    "foo",
								"dns.gardener.cloud/ttl":      "foo",
							}
							existingService.Labels = map[string]string{
								"foo": "bar",
							}
							existingService.Spec.Ports = []corev1.ServicePort{
								{
									Name:     kubeAPIServerServicePortName,
									NodePort: nodePort,
								},
								{
									Name: "foo",
								},
							}
							existingService.DeepCopyInto(obj.(*corev1.Service))
							return nil
						}),
					c.
						EXPECT().
						Update(ctx, service),
				)

				Expect(op.DeployKubeAPIServerService(ctx)).To(Succeed())
			})

			It("should correctly update the service object with SNI", func() {
				op = &operation{client: c, namespace: namespace, imports: imports}

				service := svc.DeepCopy()
				service.Labels = map[string]string{
					"foo":       "bar",
					"app":       "virtual-garden",
					"component": "kube-apiserver",
				}
				service.Annotations = map[string]string{
					"bar":                         "baz",
					"dns.gardener.cloud/dnsnames": hostname1 + "," + hostname2,
					"dns.gardener.cloud/class":    dnsClass,
					"dns.gardener.cloud/ttl":      strconv.Itoa(int(ttl)),
				}
				setExpectedServiceFields(service)

				gomock.InOrder(
					c.
						EXPECT().
						Get(ctx, client.ObjectKey{Name: KubeAPIServerServiceName, Namespace: namespace}, gomock.AssignableToTypeOf(&corev1.Service{})).
						DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
							existingService := svc.DeepCopy()
							existingService.Labels = map[string]string{"foo": "bar"}
							existingService.Annotations = map[string]string{"bar": "baz"}
							existingService.DeepCopyInto(obj.(*corev1.Service))
							return nil
						}),
					c.
						EXPECT().
						Update(ctx, service),
				)

				Expect(op.DeployKubeAPIServerService(ctx)).To(Succeed())
			})

			It("should return the error in case UPDATE fails", func() {
				service := svc.DeepCopy()
				service.Labels = map[string]string{
					"app":       "virtual-garden",
					"component": "kube-apiserver",
				}
				setExpectedServiceFields(service)

				gomock.InOrder(
					c.
						EXPECT().
						Get(ctx, client.ObjectKey{Name: KubeAPIServerServiceName, Namespace: namespace}, gomock.AssignableToTypeOf(&corev1.Service{})),
					c.
						EXPECT().
						Update(ctx, service).
						Return(fakeErr),
				)

				Expect(op.DeployKubeAPIServerService(ctx)).To(MatchError(fakeErr))
			})
		})
	})

	Describe("#DeleteKubeAPIServerService", func() {
		It("should succeed deleting the service", func() {
			c.EXPECT().Delete(ctx, svc)
			Expect(op.DeleteKubeAPIServerService(ctx)).To(Succeed())
		})

		It("should succeed deleting the service if it does not exist", func() {
			c.EXPECT().Delete(ctx, svc).Return(apierrors.NewNotFound(schema.GroupResource{}, KubeAPIServerServiceName))
			Expect(op.DeleteKubeAPIServerService(ctx)).To(Succeed())
		})

		It("should fail deleting the service if an unexpected error occurs", func() {
			c.EXPECT().Delete(ctx, svc).Return(fakeErr)
			Expect(op.DeleteKubeAPIServerService(ctx)).To(MatchError(fakeErr))
		})
	})
})

func setExpectedServiceFields(service *corev1.Service) {
	service.Spec.Type = corev1.ServiceTypeLoadBalancer
	service.Spec.Selector = map[string]string{
		"app":       "virtual-garden",
		"component": "kube-apiserver",
	}
	service.Spec.Ports = []corev1.ServicePort{
		{
			Name:       kubeAPIServerServicePortName,
			Protocol:   corev1.ProtocolTCP,
			Port:       kubeAPIServerServicePort,
			TargetPort: intstr.FromInt(kubeAPIServerServicePort),
		},
	}
}

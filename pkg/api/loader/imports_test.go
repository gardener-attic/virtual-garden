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

package loader_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gardener/virtual-garden/pkg/api"
	. "github.com/gardener/virtual-garden/pkg/api/loader"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Imports", func() {
	Describe("#FromFile", func() {
		It("should fail because the path does not exist", func() {
			_, err := FromFile("does-not-exist")
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
		})

		Context("successful read", func() {
			var (
				dir string
				err error
			)

			BeforeEach(func() {
				dir, err = ioutil.TempDir("", "test-imports")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				os.RemoveAll(dir)
			})

			It("should succeed reading but fail parsing the file", func() {
				path := filepath.Join(dir, "imports.yaml")
				Expect(ioutil.WriteFile(path, []byte("foo"), 0644)).To(Succeed())

				imports, err := FromFile(path)
				Expect(err).To(HaveOccurred())
				Expect(imports).To(BeNil())
			})

			It("should succeed reading and parsing the file (YAML)", func() {
				path := filepath.Join(dir, "imports.yaml")
				Expect(ioutil.WriteFile(path, []byte("virtualGarden: {}\nhostingCluster: {namespace: foo}\ncredentials: {}"), 0644)).To(Succeed())

				imports, err := FromFile(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(imports).To(Equal(&api.Imports{
					HostingCluster: api.HostingCluster{Namespace: "foo"},
				}))
			})

			It("should succeed reading and parsing the file (JSON)", func() {
				path := filepath.Join(dir, "imports.json")
				Expect(ioutil.WriteFile(path, []byte(`{"virtualGarden": {}, "hostingCluster": {"namespace": "foo"}, "credentials": {}}`), 0644)).To(Succeed())

				imports, err := FromFile(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(imports).To(Equal(&api.Imports{
					HostingCluster: api.HostingCluster{Namespace: "foo"},
				}))
			})
		})
	})
})

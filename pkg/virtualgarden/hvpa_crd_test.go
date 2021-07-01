package virtualgarden

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HVPACRD", func() {
	Describe("#HVPACRDLoading", func() {
		It("should load the HVPA CRD", func() {
			crd, err := loadHVPACRD()
			Expect(err).To(Succeed())
			Expect(crd.Spec.Names.Kind).To(Equal("Hvpa"))
		})
	})
})

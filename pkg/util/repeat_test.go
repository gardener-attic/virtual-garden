package util_test

import (
	"github.com/gardener/virtual-garden/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Repeat", func() {
	It("should return true when task is done", func() {
		i := 0

		ok := util.Repeat(func() bool {
			i++
			return i > 2
		}, 3, 0)

		Expect(ok).To(BeTrue())
	})

	It("should return false when task is not done after maximum number of repetitions", func() {
		ok := util.Repeat(func() bool {
			return false
		}, 3, 0)

		Expect(ok).To(BeFalse())
	})
})

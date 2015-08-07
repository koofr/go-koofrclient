package integrationtest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClientUser", func() {
	It("should get user info", func() {
		info, err := client.UserInfo()
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Email).NotTo(BeEmpty())
	})
})

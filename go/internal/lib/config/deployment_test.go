package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/iamnande/hyrule/go/internal/lib/config"
)

var _ = Describe("LoadDeployment", func() {
	DescribeTable("region validation",
		func(region string, valid bool) {
			GinkgoT().Setenv("HYRULE_REGION", region)
			_, err := config.LoadDeployment()()
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},
		Entry("us-west-2 region", string(config.USWest2Region), true),
		Entry("us-east-2 region", string(config.USEast2Region), true),
		Entry("unknown region", "nowhere", false),
	)

	DescribeTable("environment validation",
		func(environment string, valid bool) {
			GinkgoT().Setenv("HYRULE_ENVIRONMENT", environment)
			_, err := config.LoadDeployment()()
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},
		Entry("local", string(config.LocalEnvironment), true),
		Entry("prod", string(config.ProdEnvironment), true),
		Entry("unknown", "nowhere", false),
	)
})

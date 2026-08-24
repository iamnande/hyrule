package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/iamnande/hyrule/internal/lib/config"
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
		Entry("primary region", string(config.PrimaryRegion), true),
		Entry("secondary region", string(config.SecondaryRegion), true),
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
		Entry("dev", string(config.DevEnvironment), true),
		Entry("prod", string(config.ProdEnvironment), true),
		Entry("unknown", "nowhere", false),
	)
})

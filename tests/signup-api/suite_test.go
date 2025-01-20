package signupapi

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSignUpAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SignUp API Suite")
}

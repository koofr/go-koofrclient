package koofrclient_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGoKoofrclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Koofrclient Suite")
}

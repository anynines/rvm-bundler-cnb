package bundler_test

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler"
	"github.com/avarteqgmbh/rvm-bundler-cnb/bundler/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testRubyVersionResolver(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		resolver bundler.RubyVersionResolver
		bashCmd  *fakes.BashCmd
	)

	context("RubyVersionResolver", func() {
		it.Before(func() {
			bashCmd = &fakes.BashCmd{}
		})

		it("Return RubyVersionResolver", func() {
			result := bundler.NewRubyVersionResolver()
			Expect(result).To(BeAssignableToTypeOf(bundler.RubyVersionResolver{}))
		})

		it("Return result with ruby version", func() {
			bashCmd.RunBashCmdCall.Returns.String = "ruby-2.0.0"

			workingDir, err := ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			result, err := resolver.Lookup(workingDir, bashCmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("ruby-2.0"))
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("Return an error on no ruby found", func() {
			bashCmd.RunBashCmdCall.Returns.String = "some text"

			workingDir, err := ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			_, err = resolver.Lookup(workingDir, bashCmd)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("no string with ruby version found")))

			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("Return an error on bash exec failure", func() {
			bashCmd.RunBashCmdCall.Returns.Error = errors.New("failed to execute bash command")

			workingDir, err := ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			_, err = resolver.Lookup(workingDir, bashCmd)
			Expect(err).To(HaveOccurred())
			Expect(err).Should(MatchError(ContainSubstring("failed to obtain ruby version:")))

			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})
	})
}

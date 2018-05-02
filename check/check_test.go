package check_test

import (
	"testing"

	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/bborbe/k8s-manifest-check/check"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8s Manifest Check Suite")
}

var _ = Describe("Check", func() {

	Context("not existing path", func() {
		var manifestpath string
		BeforeEach(func() {
			dir, err := ioutil.TempDir("", "file")
			Expect(err).To(BeNil())
			manifestpath = path.Join(dir, "not-existing-file")
		})
		It("return file not found error", func() {
			err := check.Path(manifestpath)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("not readable path", func() {
		var manifestpath string
		BeforeEach(func() {
			dir, err := ioutil.TempDir("", "file")
			Expect(err).To(BeNil())
			manifestpath = dir
		})
		It("return file not found error", func() {
			err := check.Path(manifestpath)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("empty content", func() {
		var manifestpath string
		BeforeEach(func() {
			manifestpath = writeTempFile(``)
		})
		AfterEach(func() {
			os.Remove(manifestpath)
		})
		It("return file not found error", func() {
			err := check.Path(manifestpath)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal(fmt.Sprintf("content is empty in %s", manifestpath)))
		})
	})
	Context("invalid content", func() {
		var manifestpath string
		BeforeEach(func() {
			manifestpath = writeTempFile(`hello world`)
		})
		AfterEach(func() {
			os.Remove(manifestpath)
		})
		It("return file not found error", func() {
			err := check.Path(manifestpath)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal(fmt.Sprintf("parse content failed in %s", manifestpath)))
		})
	})
	Context("valid content", func() {
		var manifestpath string
		BeforeEach(func() {
			manifestpath = writeTempFile(`apiVersion: v1
kind: Pod
metadata:
  name: hello-world
spec:
  containers:
  - name: hello
    image: "ubuntu:14.04"
    resources:
      limits:
        cpu: 100m
        memory: 50Mi
      requests:
        cpu: 10m
        memory: 10Mi
`)
		})
		AfterEach(func() {
			os.Remove(manifestpath)
		})
		It("return file not found error", func() {
			err := check.Path(manifestpath)
			Expect(err).To(BeNil())
		})
	})
})

func writeTempFile(content string) string {
	tmpfile, err := ioutil.TempFile("", "temp-file")
	if err != nil {
		Expect(err).To(BeNil())
	}
	tmpfile.WriteString(content)
	return tmpfile.Name()
}

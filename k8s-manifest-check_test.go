package main

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"github.com/onsi/gomega/gbytes"
	"time"
	"io/ioutil"
	"os"
	"path"
)

var pathToServerBinary string
var serverSession *gexec.Session

var _ = BeforeSuite(func() {
	var err error
	pathToServerBinary, err = gexec.Build("github.com/bborbe/k8s-manifest-check")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	serverSession.Interrupt()
	Eventually(serverSession).Should(gexec.Exit())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8s Manifest Check Suite")
}

var _ = Describe("Check", func() {
	var err error
	It("print usage if no arg is given", func() {
		serverSession, err = gexec.Start(exec.Command(pathToServerBinary), GinkgoWriter, GinkgoWriter)
		Expect(err).To(BeNil())
		serverSession.Wait(100 * time.Millisecond)
		Expect(serverSession.Buffer()).To(gbytes.Say("missing arg"))
		Expect(serverSession.ExitCode()).To(Equal(1))
	})
	Context("with manifests", func() {
		var args []string
		var manifestpath string
		AfterEach(func() {
			os.Remove(manifestpath)
		})
		Context("valid", func() {
			BeforeEach(func() {
				content := `apiVersion: v1
kind: Pod
metadata:
  name: hello-world
spec:
  containers:
  - name: hello
    image: "ubuntu:14.04"
`
				tmpfile, err := ioutil.TempFile("", "example")
				if err != nil {
					Expect(err).To(BeNil())
				}
				tmpfile.WriteString(content)
				manifestpath = tmpfile.Name()
				args = []string{
					manifestpath,
				}
			})
			It("print nothing", func() {
				serverSession, err = gexec.Start(exec.Command(pathToServerBinary, args...), GinkgoWriter, GinkgoWriter)
				Expect(err).To(BeNil())
				serverSession.Wait(100 * time.Millisecond)
				Expect(serverSession.ExitCode()).To(Equal(0))
			})
		})
		Context("invalid manifestpath", func() {
			BeforeEach(func() {
				manifestpath = path.Join(os.TempDir(), "not-existing-file")
			})
			It("print error", func() {
				serverSession, err = gexec.Start(exec.Command(pathToServerBinary, args...), GinkgoWriter, GinkgoWriter)
				Expect(err).To(BeNil())
				serverSession.Wait(100 * time.Millisecond)
				Expect(serverSession.ExitCode()).To(Equal(1))
				Expect(serverSession.Buffer()).To(gbytes.Say("manifest %s not found", manifestpath))
			})
		})
	})

})

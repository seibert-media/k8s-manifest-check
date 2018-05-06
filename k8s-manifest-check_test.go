package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
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
		var manifestpath string
		AfterEach(func() {
			os.Remove(manifestpath)
		})
		Context("valid", func() {
			BeforeEach(func() {
				manifestpath = writeManifest(`apiVersion: v1
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
        memory: 100Mi
      requests:
        cpu: 100m
        memory: 100Mi
`)
			})
			It("print nothing", func() {
				serverSession, err = gexec.Start(exec.Command(pathToServerBinary, manifestpath), GinkgoWriter, GinkgoWriter)
				Expect(err).To(BeNil())
				serverSession.Wait(100 * time.Millisecond)
				Expect(serverSession.ExitCode()).To(Equal(0))
			})
		})
		Context("request cpu greater cpu limit", func() {
			BeforeEach(func() {
				manifestpath = writeManifest(`apiVersion: v1
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
        memory: 100Mi
      requests:
        cpu: 200m
        memory: 100Mi
`)
			})
			It("print warning", func() {
				serverSession, err = gexec.Start(exec.Command(pathToServerBinary, manifestpath), GinkgoWriter, GinkgoWriter)
				Expect(err).To(BeNil())
				serverSession.Wait(100 * time.Millisecond)
				Expect(serverSession.ExitCode()).To(Equal(1))
				Expect(serverSession.Buffer()).To(gbytes.Say("cpu request must be less than or equal to cpu limit in %s", manifestpath))
			})
		})
		Context("request memory greater memory limit", func() {
			BeforeEach(func() {
				manifestpath = writeManifest(`apiVersion: v1
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
        memory: 100Mi
      requests:
        cpu: 100m
        memory: 200Mi
`)
			})
			It("print warning", func() {
				serverSession, err = gexec.Start(exec.Command(pathToServerBinary, manifestpath), GinkgoWriter, GinkgoWriter)
				Expect(err).To(BeNil())
				serverSession.Wait(100 * time.Millisecond)
				Expect(serverSession.ExitCode()).To(Equal(1))
				Expect(serverSession.Buffer()).To(gbytes.Say("memory request must be less than or equal to memory limit in %s", manifestpath))
			})
		})
		Context("valid but no limits", func() {
			BeforeEach(func() {
				manifestpath = writeManifest(`apiVersion: v1
kind: Pod
metadata:
  name: hello-world
spec:
  containers:
  - name: hello
    image: "ubuntu:14.04"
`)
			})
			It("print warning", func() {
				serverSession, err = gexec.Start(exec.Command(pathToServerBinary, manifestpath), GinkgoWriter, GinkgoWriter)
				Expect(err).To(BeNil())
				serverSession.Wait(100 * time.Millisecond)
				Expect(serverSession.ExitCode()).To(Equal(1))
				Expect(serverSession.Buffer()).To(gbytes.Say("cpu request is zero in %s", manifestpath))
			})
		})
		Context("not existing manifest", func() {
			BeforeEach(func() {
				manifestpath = path.Join(os.TempDir(), "not-existing-file")
			})
			It("print error", func() {
				serverSession, err = gexec.Start(exec.Command(pathToServerBinary, manifestpath), GinkgoWriter, GinkgoWriter)
				Expect(err).To(BeNil())
				serverSession.Wait(100 * time.Millisecond)
				Expect(serverSession.ExitCode()).To(Equal(1))
				Expect(serverSession.Buffer()).To(gbytes.Say("manifest %s not found", manifestpath))
			})
		})
	})
})

func writeManifest(content string) (path string) {
	tmpfile, err := ioutil.TempFile("", "example")
	Expect(err).To(BeNil())
	tmpfile.WriteString(content)
	path = tmpfile.Name()
	return
}

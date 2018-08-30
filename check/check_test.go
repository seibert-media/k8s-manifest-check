package check_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/seibert-media/k8s-manifest-check/check"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

var _ = Describe("Resources", func() {
	var err error
	var requirements corev1.ResourceRequirements
	BeforeEach(func() {
		requirements = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				"cpu":    resource.MustParse("10m"),
				"memory": resource.MustParse("10m"),
			},
			Limits: corev1.ResourceList{
				"cpu":    resource.MustParse("10m"),
				"memory": resource.MustParse("10m"),
			},
		}
	})
	It("return no error with valid resources", func() {
		err = check.Resources(requirements)
		Expect(err).To(BeNil())
	})
	It("return error if request cpu missing", func() {
		delete(requirements.Requests, "cpu")
		err = check.Resources(requirements)
		Expect(err).NotTo(BeNil())
	})
	It("return error if request memory missing", func() {
		delete(requirements.Requests, "memory")
		err = check.Resources(requirements)
		Expect(err).NotTo(BeNil())
	})
	It("return error if limit cpu missing", func() {
		delete(requirements.Limits, "cpu")
		err = check.Resources(requirements)
		Expect(err).NotTo(BeNil())
	})
	It("return error if limit memory missing", func() {
		delete(requirements.Limits, "memory")
		err = check.Resources(requirements)
		Expect(err).NotTo(BeNil())
	})
	It("return error if cpu limit is below cpu request", func() {
		requirements.Requests["cpu"] = resource.MustParse("20m")
		err = check.Resources(requirements)
		Expect(err).NotTo(BeNil())
	})
	It("return error if memory limit is below memory request", func() {
		requirements.Requests["memory"] = resource.MustParse("20m")
		err = check.Resources(requirements)
		Expect(err).NotTo(BeNil())
	})
	It("return no error if cpu limit is above cpu request", func() {
		requirements.Limits["cpu"] = resource.MustParse("20m")
		err = check.Resources(requirements)
		Expect(err).To(BeNil())
	})
	It("return no error if memory limit is above memory request", func() {
		requirements.Limits["memory"] = resource.MustParse("20m")
		err = check.Resources(requirements)
		Expect(err).To(BeNil())
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

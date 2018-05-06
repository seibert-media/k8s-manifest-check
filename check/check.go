package check

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8s_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func Path(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("manifest %s not found", path)
		os.Exit(1)
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read manifest %s failed", path)
	}
	if err := Content(content); err != nil {
		return fmt.Errorf("%s in %s", err.Error(), path)
	}
	return nil
}

func Content(content []byte) error {
	if len(content) == 0 {
		return errors.New("content is empty")
	}
	obj, err := parseObject(content)
	if err != nil {
		glog.V(4).Infof("parse content failed: %v", err)
		return errors.New("parse content failed")
	}
	switch o := obj.(type) {
	case *corev1.Pod:
		return checkContainers(o.Spec.Containers)
	case *appsv1.Deployment:
		return checkContainers(o.Spec.Template.Spec.Containers)
	case *extv1beta1.Deployment:
		return checkContainers(o.Spec.Template.Spec.Containers)
	case *appsv1beta1.Deployment:
		return checkContainers(o.Spec.Template.Spec.Containers)
	case *appsv1beta2.Deployment:
		return checkContainers(o.Spec.Template.Spec.Containers)
	default:
		glog.V(4).Infof("type %T not checked", obj)
	}
	return nil
}

func parseObject(content []byte) (k8s_runtime.Object, error) {
	content, err := yaml.YAMLToJSON(content)
	if err != nil {
		return nil, fmt.Errorf("yaml to json failed: %v", err)
	}
	obj, err := kind(content)
	if err != nil {
		return nil, fmt.Errorf("create object by content failed: %v", err)
	}
	if obj, _, err = unstructured.UnstructuredJSONScheme.Decode(content, nil, obj); err != nil {
		return nil, fmt.Errorf("unmarshal to object failed: %v", err)
	}
	return obj, nil
}

func kind(content []byte) (k8s_runtime.Object, error) {
	_, kind, err := unstructured.UnstructuredJSONScheme.Decode(content, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("unmarshal to unknown failed: %v", err)
	}
	if kind.Kind == "Secret" {
		return nil, nil
	}
	obj, err := scheme.Scheme.New(*kind)
	if err != nil {
		return nil, fmt.Errorf("create object failed: %v", err)
	}
	return obj, nil
}

func checkContainers(containers []corev1.Container) error {
	for _, container := range containers {
		if err := Resources(container.Resources); err != nil {
			return err
		}
	}
	return nil
}

func Resources(resourceRequirements corev1.ResourceRequirements) error {
	if resourceRequirements.Requests.Cpu().IsZero() {
		return errors.New("cpu request is zero")
	}
	if resourceRequirements.Requests.Memory().IsZero() {
		return errors.New("memory request is zero")
	}
	if resourceRequirements.Limits.Memory().IsZero() {
		return errors.New("memory limit is zero")
	}
	if resourceRequirements.Limits.Cpu().IsZero() {
		return errors.New("cpu limit is zero")
	}
	if resourceRequirements.Requests.Cpu().Cmp(*resourceRequirements.Limits.Cpu()) > 0 {
		return errors.New("cpu request must be less than or equal to cpu limit")
	}
	if resourceRequirements.Requests.Memory().Cmp(*resourceRequirements.Limits.Memory()) > 0 {
		return errors.New("memory request must be less than or equal to memory limit")
	}
	return nil
}

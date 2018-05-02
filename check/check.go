package check

import (
	"fmt"
	"io/ioutil"
	"os"
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8s_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"errors"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
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
	return Content(content)
}

func Content(content []byte) error {
	if len(content) == 0 {
		return errors.New("content is empty")
	}
	content, err := yaml.YAMLToJSON(content)
	if err != nil {
		return fmt.Errorf("yaml to json failed: %v", err)
	}
	obj, err := kind(content)
	if err != nil {
		return fmt.Errorf("create object by content failed: %v", err)
	}
	if obj, _, err = unstructured.UnstructuredJSONScheme.Decode(content, nil, obj); err != nil {
		return fmt.Errorf("unmarshal to object failed: %v", err)
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
		if err := checkContainer(container); err != nil {
			return err
		}
	}
	return nil
}

func checkContainer(container corev1.Container) error {
	if container.Resources.Requests.Cpu().IsZero() {
		return fmt.Errorf("cpu request is zero")
	}
	if container.Resources.Requests.Memory().IsZero() {
		return fmt.Errorf("memory request is zero")
	}
	if container.Resources.Limits.Memory().IsZero() {
		return fmt.Errorf("memory limit is zero")
	}
	if container.Resources.Limits.Cpu().IsZero() {
		return fmt.Errorf("cpu limit is zero")
	}
	return nil
}

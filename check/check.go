package check

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8s_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func Check(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("manifest %s not found", path)
		os.Exit(1)
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read manifest %s failed", path)
	}
	if len(content) == 0 {
		return fmt.Errorf("manifest %s is empty", path)
	}
	return nil
}

func formatYaml(content []byte) ([]byte, error) {
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
	return yaml.Marshal(obj)
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

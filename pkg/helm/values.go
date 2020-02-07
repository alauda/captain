package helm

import (
	"fmt"

	"k8s.io/klog"

	"github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"github.com/ghodss/yaml"
	"helm.sh/helm/pkg/chartutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//Values is an alias for map, we cannot use chartutils.Values because the helm code
//only support map when iterate the map
type Values = map[string]interface{}

// Merges source and destination `chartutils.Values`, preferring values from the source Values
// This is slightly adapted from https://github.com/helm/helm/blob/2332b480c9cb70a0d8a85247992d6155fbe82416/cmd/helm/install.go#L359
func mergeValues(dest, src Values) Values {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

// getValues merges all values settings from spec/configmap/secret...
func getValues(hr *v1alpha1.HelmRequest, cfg *rest.Config) (chartutil.Values, error) {
	values, err := getValuesFromSource(hr, cfg)
	if err != nil {
		return nil, err
	}

	new := Values(hr.Spec.HelmValues.DeepCopy().Values)
	values = mergeValues(values, new)
	klog.V(2).Infof("get values for helm request: %s  %+v", hr.GetName(), values)
	return values, nil

}

func getValuesFromSource(hr *v1alpha1.HelmRequest, cfg *rest.Config) (chartutil.Values, error) {
	klog.V(2).Infof("in cluster rest config is: %+v", cfg)
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	ns := hr.GetNamespace()
	values := Values{}

	if hr.Spec.ValuesFrom != nil {
		for _, s := range hr.Spec.ValuesFrom {
			if s.ConfigMapKeyRef != nil {
				v, err := getValuesFromConfigMap(s.ConfigMapKeyRef, client, ns)
				if err != nil {
					return nil, err
				}
				values = mergeValues(values, v)

			}

			if s.SecretKeyRef != nil {
				v, err := getValuesFromSecret(s.SecretKeyRef, client, ns)
				if err != nil {
					return nil, err
				}
				values = mergeValues(values, v)
			}
		}
	}
	return values, nil

}

func getValuesFromSecret(s *v1.SecretKeySelector, client *kubernetes.Clientset, ns string) (chartutil.Values, error) {
	optional := s.Optional != nil && *s.Optional
	secret, err := client.CoreV1().Secrets(ns).Get(s.Name, metav1.GetOptions{})
	if err != nil {
		if optional {
			return nil, nil
		}
		return nil, err
	}

	key := s.Key
	if key == "" {
		key = "values.yaml"
	}

	data, ok := secret.Data[key]

	if !ok {
		if optional {
			return nil, nil
		}
		return nil, fmt.Errorf("key %s missing in secret %s", key, s.Name)
	}

	var values Values
	if err := yaml.Unmarshal(data, &values); err != nil {
		if optional {
			return nil, nil
		}
		return nil, err
	}
	return values, nil
}

func getValuesFromConfigMap(c *v1.ConfigMapKeySelector, client *kubernetes.Clientset, ns string) (chartutil.Values, error) {
	optional := c.Optional != nil && *c.Optional
	cm, err := client.CoreV1().ConfigMaps(ns).Get(c.Name, metav1.GetOptions{})
	if err != nil {
		if optional {
			return nil, nil
		}
		return nil, err
	}

	key := c.Key
	if key == "" {
		key = "values.yaml"
	}

	data, ok := cm.Data[key]

	if !ok {
		if optional {
			return nil, nil
		}
		return nil, fmt.Errorf("key %s missing in configmap %s", key, c.Name)
	}

	var values Values
	if err := yaml.Unmarshal([]byte(data), &values); err != nil {
		if optional {
			return nil, nil
		}
		return nil, err
	}
	return values, nil
}

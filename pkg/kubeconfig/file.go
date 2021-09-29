package kubeconfig

import (
	"os"
	"path/filepath"

	"k8s.io/klog"
)

// CreatePathIfNotExist is a thin wrapper to create all the paths for a file if not exist
func CreatePathIfNotExist(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		klog.Info("dir not exist, create it: ", dir)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if _, err = os.Create(path); err != nil {
			return err
		}
		klog.Info(" file not exist, create it: ", path)
	}
	return nil
}

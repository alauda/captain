package util

import (
	"context"
	"time"

	"github.com/alauda/captain/pkg/helmrequest"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

// InstallCRDIfRequired install helmrequest CRD
// may be we should move it out of main
func InstallCRDIfRequired(cfg *rest.Config, required bool) error {
	if required {
		return wait.PollImmediateUntil(time.Second*5, func() (bool, error) {
			return helmrequest.EnsureCRDCreated(cfg)
		}, context.TODO().Done())
	}
	return nil
}

// InstallHelmRequestCRD install HelmRequest crd with timeout
func InstallHelmRequestCRD(cfg *rest.Config) error {
	return wait.Poll(time.Second*3, time.Second*15, func() (bool, error) {
		return helmrequest.EnsureCRDCreated(cfg)
	})
}

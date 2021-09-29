package release

import (
	"time"

	"github.com/alauda/captain/pkg/helmrequest"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

// EnsureCRDCreated tries to create/update CRD, returns (true, nil) if succeeding, otherwise returns (false, nil).
// 'err' should always be nil, because it is used by wait.PollUntil(), and it will exit if it is not nil.
func EnsureCRDCreated(cfg *rest.Config) error {
	crdVar, err := helmrequest.CreateCRDObject(releaseCRDYaml)
	if err != nil {
		return err
	}

	return wait.PollImmediate(time.Second*3, time.Second*30, func() (bool, error) {
		return helmrequest.EnsureCRDCreatedWithConfig(cfg, crdVar)
	})
}

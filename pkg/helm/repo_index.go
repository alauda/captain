package helm

import (
	"time"

	"helm.sh/helm/pkg/repo"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

var (
	defaultInterval = 30
)

//IndexSyncer sync helm repo index repeatedly
type IndexSyncer struct {
	// interval is the interval of the sync process
	interval int
}

func NewDefaultIndexSyncer() *IndexSyncer {
	return &IndexSyncer{
		interval: defaultInterval,
	}
}

func (i *IndexSyncer) Start(stop <-chan struct{}) error {
	return wait.PollUntil(time.Second*time.Duration(i.interval),
		func() (done bool, err error) {
			klog.V(4).Info("update helm repo index")
			err = initReposIndex()
			if err != nil {
				klog.Error("update helm repo index error: ", err)
			}
			return false, nil
		},
		stop,
	)
}

// initReposIndex update index for all the known repos. This happens when captain starts.
func initReposIndex() error {
	f, err := repo.LoadFile(getHelmHome().RepositoryFile())
	if err != nil {
		return err
	}
	if len(f.Repositories) == 0 {
		return nil
	}
	for _, re := range f.Repositories {
		err := addRepository(re.Name, re.URL, re.Username, re.Password, getHelmHome(), "", "", "", false)
		if err != nil {
			klog.Warningf("repo index update error for %s: %s", re.Name, err.Error())
			continue
		} else {
			klog.Infof("update index done for repo: %s", re.Name)
		}
	}
	return nil
}

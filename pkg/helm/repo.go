package helm

import (
	"fmt"

	"helm.sh/helm/pkg/helmpath"
	"helm.sh/helm/pkg/repo"
)

// GetChartsForRepo retrieve charts info from a repo cache index
// Check: can we use the generated time to do compare?
func GetChartsForRepo(name string) (*repo.IndexFile, error) {
	path := helmpath.CachePath("repository") + fmt.Sprintf("/%s-index.yaml", name)
	return repo.LoadIndexFile(path)
}

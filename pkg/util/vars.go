package util

const (
	// ComponentName is the name of this project
	ComponentName = "captain"

	//LeaderLockName is the name of lock for leader election
	LeaderLockName = "captain-controller-lock"

	// FinalizerName is the finalizer name we append to each HelmRequest resource
	FinalizerName = "captain.cpaas.io"

	// ProjectKey is the annotation key for project
	ProjectKey = "alauda.io/project"
)

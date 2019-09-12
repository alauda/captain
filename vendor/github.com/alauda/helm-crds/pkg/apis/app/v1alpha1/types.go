package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"

	v1 "k8s.io/api/core/v1"

	"github.com/ghodss/yaml"
	"helm.sh/helm/pkg/chartutil"


	funk "github.com/thoas/go-funk"

	"github.com/alauda/component-base/regex"
	"github.com/fatih/structs"
	"helm.sh/helm/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

const (
	// FinalizerName is the finalizer name we append to each HelmRequest resource
	FinalizerName = "captain.alauda.io"
)


// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Release struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReleaseSpec   `json:"spec"`
	Status ReleaseStatus `json:"status"`
}

// ReleaseSpec describes a deployment of a chart, together with the chart
// and the variables used to deploy that chart.
type ReleaseSpec struct {
	// ChartData is the chart that was released.
	ChartData string `json:"chartData,omitempty"`
	// ConfigData is the set of extra Values added to the chart.
	// These values override the default values inside of the chart.
	ConfigData string `json:"configData,omitempty"`
	// ManifestData is the string representation of the rendered template.
	ManifestData string `json:"manifestData,omitempty"`
	// Hooks are all of the hooks declared for this release.
	HooksData string `json:"hooksData,omitempty"`
	// Version is an int which represents the version of the release.
	Version int `json:"version,omitempty"`

	Name string `json:"name,omitempty"`
}

// Info describes release information.

type ReleaseStatus struct {
	// FirstDeployed is when the release was first deployed.
	FirstDeployed metav1.Time `json:"first_deployed,omitempty"`
	// LastDeployed is when the release was last deployed.
	LastDeployed metav1.Time `json:"last_deployed,omitempty"`
	// Deleted tracks when this object was deleted.
	Deleted metav1.Time `json:"deleted,omitempty"`
	// Description is human-friendly "log entry" about this release.
	Description string `json:"Description,omitempty"`
	// Status is the current state of the release
	Status release.Status `json:"status,omitempty"`
	// Contains the rendered templates/NOTES.txt if available
	Notes string `json:"notes,omitempty"`
}

func (in *ReleaseStatus) CopyFromReleaseInfo(info *release.Info) {
	in.Status = info.Status
	in.Deleted = metav1.NewTime(info.Deleted)
	in.Description = info.Description
	in.FirstDeployed = metav1.NewTime(info.FirstDeployed)
	in.LastDeployed = metav1.NewTime(info.LastDeployed)
	in.Notes = info.Notes
}

func (in *ReleaseStatus) ToReleaseInfo() *release.Info {
	var info release.Info

	info.Status = in.Status
	info.Deleted = in.Deleted.Time
	info.Description = in.Description
	info.FirstDeployed = in.FirstDeployed.Time
	info.LastDeployed = in.LastDeployed.Time
	info.Notes = in.Notes
	return &info
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []Release `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ChartRepo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ChartRepoSpec   `json:"spec"`
	Status            ChartRepoStatus `json:"status"`
}

func (in *ChartRepo) ValidateCreate() error {
	return nil

}

func (in *ChartRepo) ValidateUpdate(old runtime.Object) error {
	klog.V(4).Info("validate chartrepo update: ", in.GetName())

	oldRepo, ok := old.(*ChartRepo)
	if !ok {
		return fmt.Errorf("expect old object to be a %T instead of %T", oldRepo, old)
	}

	if in.Spec.URL != oldRepo.Spec.URL {
		return fmt.Errorf(".spec.url is immutable")
	}
	return nil
}

type ChartRepoSpec struct {
	// URL is the repo's url
	URL string `json:"url"`
	// Secret contains information about how to auth to this repo
	Secret *v1.SecretReference `json:"secret,omitempty"`
}

type ChartRepoPhase string

const (
	// ChartRepoSynced means is successfully recognized by captain
	ChartRepoSynced ChartRepoPhase = "Synced"

	// ChartRepoFailed means captain is unable to retrieve index info from this repo
	ChartRepoFailed ChartRepoPhase = "Failed"
)

type ChartRepoStatus struct {
	// Phase ...
	// After create, this phase will be updated to indicate it's sync status
	// If receive update event, and some field in spec changed, sync agagin.
	Phase ChartRepoPhase `json:"phase,omitempty"`
	// Reason is the failed reason
	Reason string `json:"reason,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ChartRepoList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []ChartRepo `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HelmRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmRequestSpec   `json:"spec"`
	Status HelmRequestStatus `json:"status"`
}

type HelmRequestSpec struct {
	// ClusterName is the cluster where the chart will be installed. If InstallToAllClusters=true,
	// this field will be ignored
	ClusterName string `json:"clusterName,omitempty"`

	// InstallToAllClusters will install this chart to all available clusters, even the cluster was
	// created after this chart. If this field is true, ClusterName will be ignored(useless)
	InstallToAllClusters bool `json:"installToAllClusters,omitempty"`

	// Dependencies is the dependencies of this HelmRequest, it's a list of there names
	// THe dependencies must lives in the same namespace, and each of them must be in Synced status
	// before we sync this HelmRequest
	Dependencies []string `json:"dependencies,omitempty"`

	// ReleaseName is the Release name to be generated, default to HelmRequest.Name. If we want to manually
	// install this chart to multi clusters, we may have different HelmRequest name(with cluster prefix or suffix)
	// and same release name
	ReleaseName string `json:"releaseName,omitempty"`
	Chart       string `json:"chart,omitempty"`
	Version     string `json:"version,omitempty"`
	// Namespace is the namespace where the Release object will be lived in. Notes this should be used with
	// the values defined in the chart， otherwise the install will failed
	Namespace string `json:"namespace,omitempty"`
	// ValuesFrom represents values from ConfigMap/Secret...
	ValuesFrom []ValuesFromSource `json:"valuesFrom,omitempty"`
	// values is a map
	HelmValues `json:",inline"`
}

//ValuesFromSource represents a source of values, only one of it's fields may be set
type ValuesFromSource struct {
	// ConfigMapKeyRef selects a key of a ConfigMap
	ConfigMapKeyRef *v1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	// SecretKeyRef selects a key of a Secret
	SecretKeyRef *v1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

//HelmValues embeds helm values so we can add deepcopy on it
type HelmValues struct {
	chartutil.Values `json:"values,omitempty"`
}

func (in *HelmValues) DeepCopyInto(out *HelmValues) {
	if in == nil {
		return
	}

	b, err := yaml.Marshal(in.Values)
	if err != nil {
		return
	}
	var values chartutil.Values
	err = yaml.Unmarshal(b, &values)
	if err != nil {
		return
	}
	out.Values = values
}

// HelmRequestPhase is a label for the condition of a HelmRequest at the current time.
type HelmRequestPhase string

// These are the valid statuses of pods.
const (
	HelmRequestSynced HelmRequestPhase = "Synced"

	// HelmRequestPartialSynced means the HelmRequest is partial synced to target clusters
	HelmRequestPartialSynced HelmRequestPhase = "PartialSynced"

	HelmRequestFailed HelmRequestPhase = "Failed"

	// HelmRequestPending is when helm request is syncing...
	HelmRequestPending HelmRequestPhase = "Pending"

	HelmRequestUnknown HelmRequestPhase = "Unknown"
)

type HelmRequestStatus struct {
	Phase HelmRequestPhase `json:"phase,omitempty"`
	// LastSpecHash store the has value of the synced spec, if this value not equal to the current one,
	// means we need to do a update for the chart
	LastSpecHash string `json:"lastSpecHash,omitempty"`
	// SyncedClusters will store the synced clusters if InstallToAllClusters is true
	SyncedClusters []string `json:"syncedClusters,omitempty"`

	// Notes is the contents from helm (after helm install successfully it will be printed to the console
	Notes string `json:"notes,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HelmRequestList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []HelmRequest `json:"items"`
}

// nameRegexError is thin wrapper to create regex validate error
// key is the field name and value is it's value
func (in *HelmRequest) nameRegexError(key, value string) error {
	return regex.DefaultResourceNameRegexError("HelmRequest", in.GetName(), key, value)
}

// Default makes HelmRequest an mutating webhook
// When delete, if error occurs, finalizer is a good options for us to retry and
// record the events.
func (in *HelmRequest) Default() {
	if !in.DeletionTimestamp.IsZero() {
		return
	}

	// If no releaseName applied, use HelmRequest's name
	if in.Spec.ReleaseName == "" {
		in.Spec.ReleaseName = in.GetName()
		klog.Info("use helmrequest name as release name: ", in.GetName())
	}

	// If no namespace applied, use HelmRequest's namespace
	if in.Spec.Namespace == "" {
		in.Spec.Namespace = in.GetNamespace()
		klog.Info("use helmrequest namespace as release namespace: ", in.GetNamespace())
	}

	in.Finalizers = []string{FinalizerName}
	klog.V(4).Info("append finalizers to helmrequest: ", in.GetName())
}

//ValidateCreate implements webhook.Validator
// 1. check filed regex
func (in *HelmRequest) ValidateCreate() error {
	klog.V(4).Info("validate HelmRequest create: ", in.GetName())
	if in.Spec.ClusterName != "" && !regex.IsValidResourceName(in.Spec.ClusterName) {
		return in.nameRegexError(".spec.clusterName", in.Spec.ClusterName)
	}

	if in.Spec.ReleaseName != "" && !regex.IsValidResourceName(in.Spec.ReleaseName) {
		return in.nameRegexError(".spec.releaseName", in.Spec.ReleaseName)
	}

	if len(in.Spec.Dependencies) > 0 {
		for _, name := range in.Spec.Dependencies {
			if !regex.IsValidResourceName(name) {
				return in.nameRegexError(".spec.dependencies.[]", name)
			}
		}
	}

	if len(in.Spec.ValuesFrom) > 0 {
		for _, item := range in.Spec.ValuesFrom {
			if item.ConfigMapKeyRef != nil && item.SecretKeyRef != nil {
				return fmt.Errorf("cannot set configmap ref and secret ref in the same source")
			}
		}
	}

	return nil
}

//ValidateUpdate validate HelmRequest update request
// immutable fields:
// 1. clusterName
// 2. installToAllCluster
// 3. releaseName
// 4. chart
// 5. namespace
func (in *HelmRequest) ValidateUpdate(old runtime.Object) error {
	klog.V(4).Info("validate HelmRequest update: ", in.GetName())

	oldHR, ok := old.(*HelmRequest)
	if !ok {
		return fmt.Errorf("expect old object to be a %T instead of %T", oldHR, old)
	}

	// check chart name
	_, oldChart := ParseChartName(oldHR.Spec.Chart)
	_, newChart := ParseChartName(in.Spec.Chart)

	if oldChart != newChart {
		return fmt.Errorf("chart name cannot be updated after create")
	}

	// check dependency
	if !reflect.DeepEqual(oldHR.Spec.Dependencies, in.Spec.Dependencies) {
		return fmt.Errorf("dependencies cannot be updated after create")
	}

	o := structs.New(oldHR.Spec)
	n := structs.New(in.Spec)

	for _, key := range []string{"ClusterName", "InstallToAllClusters", "ReleaseName", "Namespace"} {
		kind := o.Field(key).Kind().String()
		if kind == "string" {
			if o.Field(key).Value().(string) != n.Field(key).Value().(string) {
				return fmt.Errorf("field .spec.%s can not update after created", key)
			}
		}
		if kind == "bool" {
			if o.Field(key).Value().(bool) != n.Field(key).Value().(bool) {
				return fmt.Errorf("field .spec.%s can not update after created", key)
			}
		}
	}

	if len(in.Spec.ValuesFrom) > 0 {
		for _, item := range in.Spec.ValuesFrom {
			if item.ConfigMapKeyRef != nil && item.SecretKeyRef != nil {
				return fmt.Errorf("cannot set configmap ref and secret ref in the same source")
			}
		}
	}

	return nil

}

//IsClusterSynced check if this HelmRequest has been synced to cluster
func (in *HelmRequest) IsClusterSynced(name string) bool {
	if !in.Spec.InstallToAllClusters {
		return name == in.Spec.ClusterName && in.Status.Phase == HelmRequestSynced
	}

	clusters := in.Status.SyncedClusters
	if len(clusters) > 0 {
		return funk.Contains(clusters, name)
	}

	return false
}



// ParseChartName is a simple function that parse chart name
func ParseChartName(name string) (repo, chart string) {
	data := strings.Split(name, "/")
	if len(data) == 1 {
		return "", name
	}
	return data[0], data[1]
}

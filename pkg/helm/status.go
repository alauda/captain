package helm

import (
	"fmt"

	appv1 "github.com/alauda/helm-crds/pkg/apis/app/v1"
	clientset "github.com/alauda/helm-crds/pkg/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// UpdateHelmRequestStatus  update a helmrequest status
// This works simliar to the origin version in controller sync loop, the diff is:
// 1. no deletion helmrequest related resource when not found
// 2. more simple
// TODO: merge this two functions
func UpdateHelmRequestStatus(client clientset.Interface, request *appv1.HelmRequest) error {
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the HelmRequest resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := client.AppV1().HelmRequests(request.Namespace).UpdateStatus(request)
	if err != nil {
		if apierrors.IsConflict(err) {
			klog.Warning("update helm request status conflict, retry...")
			origin, err := client.AppV1().HelmRequests(request.Namespace).Get(request.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			klog.Warningf("origin status: %+v, current: %+v", origin.Status, request.Status)
			origin.Status = *request.Status.DeepCopy()
			_, err = client.AppV1().HelmRequests(request.Namespace).UpdateStatus(origin)
			if err != nil {
				klog.Error("retrying update helmrequest status error:", err)
			}
			return err
		}
		klog.Errorf("update status for helmrequest %s error: %s", request.Name, err.Error())
	}
	return err
}

// AddConditionForHelmRequest ...
// Note: this function will modify the hr object, this is not a good solution
func AddConditionForHelmRequest(condition *appv1.HelmRequestCondition, hr *appv1.HelmRequest, client clientset.Interface) error {
	old, err := client.AppV1().HelmRequests(hr.Namespace).Get(hr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	conditions := old.Status.Conditions
	if len(conditions) == 0 {
		conditions = []appv1.HelmRequestCondition{*condition}
	} else {
		var newConds []appv1.HelmRequestCondition
		added := false
		for _, item := range conditions {
			if item.Type == condition.Type {
				newConds = append(newConds, *condition)
				added = true
			} else {
				newConds = append(newConds, item)
			}
		}
		if !added {
			newConds = append(newConds, *condition)
		}
		conditions = newConds
	}

	old.Status.Conditions = conditions
	return UpdateHelmRequestStatus(client, old)

}

func newCondition(reason, message string, ty appv1.HelmRequestConditionType, status v1.ConditionStatus) *appv1.HelmRequestCondition {
	t := metav1.Now()
	return &appv1.HelmRequestCondition{
		LastTransitionTime: &t,
		Type:               ty,
		Reason:             reason,
		Message:            message,
		Status:             status,
	}
}

func newChartLoadedCondition(chart string) *appv1.HelmRequestCondition {
	condition := newCondition("ChartLoaded", fmt.Sprintf("chart %s loaded", chart), appv1.ConditionInitialized, v1.ConditionTrue)
	return condition
}

func (d *Deploy) addCondition(cond *appv1.HelmRequestCondition) error {
	return AddConditionForHelmRequest(cond, d.HelmRequest, d.Client)
}

func (d *Deploy) setChartLoadedCondition(chart string) error {
	cond := newChartLoadedCondition(chart)
	return d.addCondition(cond)
}

package controller

import (
	"fmt"

	"alauda.io/captain/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a HelmRequest is synced
	SuccessSynced = "Synced"

	// FailedSync is used when a HelmRequest failed to sync
	FailedSync = "FailedSync"

	// SuccessfulDelete means successfully delete a resource
	SuccessfulDelete = "SuccessfulDelete"

	// FailedDelete means failed to delete a resource
	FailedDelete = "FailedDelete"

	// ErrResourceExists is used as part of the Event 'reason' when a HelmRequest fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by HelmRequest"
	// MessageResourceSynced is the message used for an Event fired when a HelmRequest
	// is synced successfully
	MessageResourceSynced = "HelmRequest synced successfully"
)

// sendFailedDeleteEvent send a failed event when delete error
func (c *Controller) sendFailedDeleteEvent(hr *v1alpha1.HelmRequest, err error) {
	c.recorder.Event(hr, corev1.EventTypeWarning, FailedDelete,
		fmt.Sprintf("Delete HelmRequest %s error : %s", hr.GetName(), err.Error()))
}

// sendFailedSyncEvent send a event when failed to sync a resource
func (c *Controller) sendFailedSyncEvent(hr *v1alpha1.HelmRequest, err error) {
	c.recorder.Event(hr, corev1.EventTypeWarning, FailedSync, err.Error())
}

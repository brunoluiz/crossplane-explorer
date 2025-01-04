package xplane

import (
	"fmt"
	"strings"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/xplane/xpkg"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	pkgv1 "github.com/crossplane/crossplane/apis/pkg/v1"
	gcrname "github.com/google/go-containerregistry/pkg/name"
	corev1 "k8s.io/api/core/v1"
	errv1 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Resource resource trace model, extracted from crossplane CLI codebase
type Resource struct {
	Unstructured unstructured.Unstructured `json:"object"`
	Error        *errv1.StatusError        `json:"error,omitempty"`
	Children     []*Resource               `json:"children,omitempty"`
}

// GetCondition of this resource.
func (r *Resource) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	conditioned := xpv1.ConditionedStatus{}
	// The path is directly `status` because conditions are inline.
	if err := fieldpath.Pave(r.Unstructured.Object).GetValueInto("status", &conditioned); err != nil {
		return xpv1.Condition{}
	}
	// We didn't use xpv1.CondidionedStatus.GetCondition because that's defaulting the
	// status to unknown if the condition is not found at all.
	for _, c := range conditioned.Conditions {
		if c.Type == ct {
			return c
		}
	}
	return xpv1.Condition{}
}

type ResourceStatus struct {
	Name                 string
	ResourceName         string
	Ready                string
	ReadyLastTransition  time.Time
	Synced               string
	SyncedLastTransition time.Time
	Status               string
	Ok                   bool
}

func IsPkg(gk schema.GroupKind) bool {
	return xpkg.IsPackageType(gk) || xpkg.IsPackageRevisionType(gk)
}

// getResourceStatus returns a string that represents an entire row of status
// information for the resource.
func GetResourceStatus(r *Resource, name string) ResourceStatus {
	readyCond := r.GetCondition(xpv1.TypeReady)
	syncedCond := r.GetCondition(xpv1.TypeSynced)

	var status, m string
	switch {
	case r.Unstructured.GetDeletionTimestamp() != nil:
		// Report the status as deleted if the resource is being deleted
		status = "Deleting"
	case r.Error != nil:
		// if there is an error we want to show it
		status = "Error"
		m = r.Error.Error()
	case readyCond.Status == corev1.ConditionTrue && syncedCond.Status == corev1.ConditionTrue:
		// if both are true we want to show the ready reason only
		status = string(readyCond.Reason)

	// The following cases are for when one of the conditions is not true (false or unknown),
	// prioritizing synced over readiness in case of issues.
	case syncedCond.Status != corev1.ConditionTrue &&
		(syncedCond.Reason != "" || syncedCond.Message != ""):
		status = string(syncedCond.Reason)
		m = syncedCond.Message
	case readyCond.Status != corev1.ConditionTrue &&
		(readyCond.Reason != "" || readyCond.Message != ""):
		status = string(readyCond.Reason)
		m = readyCond.Message

	default:
		// both are unknown or unset, let's try showing the ready reason, probably empty
		status = string(readyCond.Reason)
		m = readyCond.Message
	}

	// Append the message to the status if it's not empty
	if m != "" {
		status = fmt.Sprintf("%s: %s", status, m)
	}

	return ResourceStatus{
		Name:                 name,
		ResourceName:         r.Unstructured.GetAnnotations()["crossplane.io/composition-resource-name"],
		Ready:                mapEmptyStatusToDash(readyCond.Status),
		ReadyLastTransition:  readyCond.LastTransitionTime.Time,
		Synced:               mapEmptyStatusToDash(syncedCond.Status),
		SyncedLastTransition: syncedCond.LastTransitionTime.Time,
		Status:               status,
		Ok:                   (syncedCond.Status == corev1.ConditionTrue && readyCond.Status == corev1.ConditionTrue),
	}
}

type PkgResourceStatus struct {
	Name                    string
	PackageImg              string
	Version                 string
	Installed               string
	InstalledLastTransition time.Time
	Healthy                 string
	HealthyLastTransition   time.Time
	State                   string
	Status                  string
	Ok                      bool
}

func GetPkgResourceStatus(r *Resource, name string) PkgResourceStatus {
	var err error
	var packageImg, state, status, m string

	healthyCond := r.GetCondition(pkgv1.TypeHealthy)
	installedCond := r.GetCondition(pkgv1.TypeInstalled)

	gk := r.Unstructured.GroupVersionKind().GroupKind()
	switch {
	case r.Error != nil:
		// If there is an error we want to show it, regardless of what type this
		// resource is and what conditions it has.
		status = "Error"
		m = r.Error.Error()
	case xpkg.IsPackageType(gk):
		switch {
		case healthyCond.Status == corev1.ConditionTrue && installedCond.Status == corev1.ConditionTrue:
			// If both are true we want to show the healthy reason only
			status = string(healthyCond.Reason)

		// The following cases are for when one of the conditions is not true (false or unknown),
		// prioritizing installed over healthy in case of issues.
		case installedCond.Status != corev1.ConditionTrue &&
			(installedCond.Reason != "" || installedCond.Message != ""):
			status = string(installedCond.Reason)
			m = installedCond.Message
		case healthyCond.Status != corev1.ConditionTrue &&
			(healthyCond.Reason != "" || healthyCond.Message != ""):
			status = string(healthyCond.Reason)
			m = healthyCond.Message
		default:
			// both are unknown or unset, let's try showing the installed reason
			status = string(installedCond.Reason)
			m = installedCond.Message
		}

		if packageImg, err = fieldpath.Pave(r.Unstructured.Object).GetString("spec.package"); err != nil {
			state = err.Error()
		}
	case xpkg.IsPackageRevisionType(gk):
		// package revisions only have the healthy condition, so use that
		status = string(healthyCond.Reason)
		m = healthyCond.Message

		// Get the state (active vs. inactive) of this package revision.
		var err error
		state, err = fieldpath.Pave(r.Unstructured.Object).GetString("spec.desiredState")
		if err != nil {
			state = err.Error()
		}
		// Get the image used.
		if packageImg, err = fieldpath.Pave(r.Unstructured.Object).GetString("spec.image"); err != nil {
			state = err.Error()
		}
	case xpkg.IsPackageRuntimeConfigType(gk):
		// nothing to do here
	default:
		status = "Unknown package type"
	}

	// Append the message to the status if it's not empty
	if m != "" {
		status = fmt.Sprintf("%s: %s", status, m)
	}

	// Parse the image reference extracting the tag, we'll leave it empty if we
	// couldn't parse it and leave the whole thing as package instead. We pass
	// an empty default registry here so the displayed package image will be
	// unmodified from what we found in the spec, similar to how kubectl output
	// behaves.
	var packageImgTag string
	if tag, err := gcrname.NewTag(packageImg, gcrname.WithDefaultRegistry("")); err == nil {
		packageImgTag = tag.TagStr()
		packageImg = tag.RepositoryStr()
		if tag.RegistryStr() != "" {
			packageImg = fmt.Sprintf("%s/%s", tag.RegistryStr(), packageImg)
		}
	}

	return PkgResourceStatus{
		Name:                    name,
		PackageImg:              packageImg,
		Version:                 packageImgTag,
		Installed:               mapEmptyStatusToDash(installedCond.Status),
		InstalledLastTransition: installedCond.LastTransitionTime.Time,
		Healthy:                 mapEmptyStatusToDash(healthyCond.Status),
		HealthyLastTransition:   healthyCond.LastTransitionTime.Time,
		State:                   mapEmptyStatusToDash(corev1.ConditionStatus(state)),
		Status:                  status,
		Ok:                      (installedCond.Status == corev1.ConditionTrue && healthyCond.Status == corev1.ConditionTrue) || strings.HasPrefix(status, "Active") || strings.HasPrefix(status, "Healthy"),
	}
}

func mapEmptyStatusToDash(s corev1.ConditionStatus) string {
	if s == "" {
		return "-"
	}
	return string(s)
}

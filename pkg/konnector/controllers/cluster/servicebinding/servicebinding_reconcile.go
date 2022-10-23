/*
Copyright 2022 The Kube Bind Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package servicebinding

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/pointer"

	kubebindv1alpha1 "github.com/kube-bind/kube-bind/pkg/apis/kubebind/v1alpha1"
	kubebindhelpers "github.com/kube-bind/kube-bind/pkg/apis/kubebind/v1alpha1/helpers"
	conditionsapi "github.com/kube-bind/kube-bind/pkg/apis/third_party/conditions/apis/conditions/v1alpha1"
	"github.com/kube-bind/kube-bind/pkg/apis/third_party/conditions/util/conditions"
)

type reconciler struct {
	consumerSecretRefKey, providerNamespace string

	getServiceExport  func(ns string) (*kubebindv1alpha1.APIServiceExport, error)
	getServiceBinding func(name string) (*kubebindv1alpha1.APIServiceBinding, error)
	getClusterBinding func(ctx context.Context) (*kubebindv1alpha1.ClusterBinding, error)

	updateServiceExportStatus func(ctx context.Context, export *kubebindv1alpha1.APIServiceExport) (*kubebindv1alpha1.APIServiceExport, error)

	getCRD    func(name string) (*apiextensionsv1.CustomResourceDefinition, error)
	updateCRD func(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition) (*apiextensionsv1.CustomResourceDefinition, error)
	createCRD func(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition) (*apiextensionsv1.CustomResourceDefinition, error)
}

func (r *reconciler) reconcile(ctx context.Context, binding *kubebindv1alpha1.APIServiceBinding) error {
	var errs []error

	if err := r.ensureValidServiceExport(ctx, binding); err != nil {
		errs = append(errs, err)
	}

	if err := r.ensureCRDs(ctx, binding); err != nil {
		errs = append(errs, err)
	}

	if err := r.ensurePrettyName(ctx, binding); err != nil {
		errs = append(errs, err)
	}

	conditions.SetSummary(binding)

	return utilerrors.NewAggregate(errs)
}

func (r *reconciler) ensureValidServiceExport(ctx context.Context, binding *kubebindv1alpha1.APIServiceBinding) error {
	if _, err := r.getServiceExport(binding.Name); err != nil && !errors.IsNotFound(err) {
		return err
	} else if errors.IsNotFound(err) {
		conditions.MarkFalse(
			binding,
			kubebindv1alpha1.APIServiceBindingConditionConnected,
			"APIServiceExportNotFound",
			conditionsapi.ConditionSeverityError,
			"APIServiceExport %s not found on the service provider cluster. Rerun kubectl bind for repair.",
			binding.Name,
		)
		return nil
	}

	conditions.MarkTrue(
		binding,
		kubebindv1alpha1.APIServiceBindingConditionConnected,
	)

	return nil
}

func (r *reconciler) ensureCRDs(ctx context.Context, binding *kubebindv1alpha1.APIServiceBinding) error {
	var errs []error

	export, err := r.getServiceExport(binding.Name)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if errors.IsNotFound(err) {
		conditions.MarkFalse(
			binding,
			kubebindv1alpha1.APIServiceBindingConditionConnected,
			"APIServiceExportNotFound",
			conditionsapi.ConditionSeverityError,
			"APIServiceExport %s not found on the service provider cluster.",
			binding.Name,
		)
		return nil // nothing we can do here
	}

	crd, err := kubebindhelpers.ServiceExportToCRD(export)
	if err != nil {
		conditions.MarkFalse(
			binding,
			kubebindv1alpha1.APIServiceBindingConditionConnected,
			"APIServiceExportInvalid",
			conditionsapi.ConditionSeverityError,
			"APIServiceExport %s on the service provider cluster is invalid: %s",
			binding.Name, err,
		)
		return nil // nothing we can do here
	}

	// put binding owner reference on the CRD.
	newReference := metav1.OwnerReference{
		APIVersion: kubebindv1alpha1.SchemeGroupVersion.String(),
		Kind:       "APIServiceBinding",
		Name:       binding.Name,
		UID:        binding.UID,
		Controller: pointer.Bool(true),
	}
	crd.OwnerReferences = append(crd.OwnerReferences, newReference)

	existing, err := r.getCRD(crd.Name)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if errors.IsNotFound(err) {
		if _, err := r.createCRD(ctx, crd); err != nil && !errors.IsInvalid(err) {
			return err
		} else if errors.IsInvalid(err) {
			conditions.MarkFalse(
				binding,
				kubebindv1alpha1.APIServiceBindingConditionConnected,
				"CustomResourceDefinitionCreateFailed",
				conditionsapi.ConditionSeverityError,
				"CustomResourceDefinition %s cannot be created: %s",
				binding.Name, err,
			)
			return nil
		}

		conditions.MarkTrue(binding, kubebindv1alpha1.APIServiceBindingConditionConnected)
		return nil // we wait for a new reconcile to update APIServiceExport status
	}

	// first check this really ours and we don't override something else
	if !kubebindhelpers.IsOwnedByBinding(binding.Name, binding.UID, crd.OwnerReferences) {
		conditions.MarkFalse(
			binding,
			kubebindv1alpha1.APIServiceBindingConditionConnected,
			"ForeignCustomResourceDefinition",
			conditionsapi.ConditionSeverityError,
			"CustomResourceDefinition %s is not owned by kube-bind.io.",
			binding.Name,
		)
		return nil
	}

	crd.ObjectMeta = existing.ObjectMeta
	updated, err := r.updateCRD(ctx, crd)
	if err != nil && !errors.IsInvalid(err) {
		return nil
	} else if errors.IsInvalid(err) {
		conditions.MarkFalse(
			binding,
			kubebindv1alpha1.APIServiceBindingConditionConnected,
			"CustomResourceDefinitionUpdateFailed",
			conditionsapi.ConditionSeverityError,
			"CustomResourceDefinition %s cannot be updated: %s",
			binding.Name, err,
		)
		return nil
	}

	conditions.MarkTrue(binding, kubebindv1alpha1.APIServiceBindingConditionConnected)

	// copy the CRD status onto the APIServiceExport
	orig := export
	export = export.DeepCopy()
	export.Status.Conditions = nil
	for _, c := range updated.Status.Conditions {
		severity := conditionsapi.ConditionSeverityError
		if c.Status == apiextensionsv1.ConditionTrue {
			severity = conditionsapi.ConditionSeverityNone
		}
		export.Status.Conditions = append(export.Status.Conditions, conditionsapi.Condition{
			Type:               conditionsapi.ConditionType(c.Type),
			Status:             corev1.ConditionStatus(c.Status),
			Severity:           severity, // CRD conditions have no severity
			LastTransitionTime: c.LastTransitionTime,
			Reason:             c.Reason,
			Message:            c.Message,
		})
	}
	export.Status.AcceptedNames = updated.Status.AcceptedNames
	export.Status.StoredVersions = updated.Status.StoredVersions
	if !equality.Semantic.DeepEqual(orig, export) {
		if _, err := r.updateServiceExportStatus(ctx, export); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

func (r *reconciler) ensurePrettyName(ctx context.Context, binding *kubebindv1alpha1.APIServiceBinding) error {
	clusterBinding, err := r.getClusterBinding(ctx)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if errors.IsNotFound(err) {
		binding.Status.ProviderPrettyName = ""
		return nil
	}

	binding.Status.ProviderPrettyName = clusterBinding.Spec.ProviderPrettyName

	return nil
}

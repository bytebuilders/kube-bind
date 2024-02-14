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

	conditionsapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/conditions"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/ptr"

	kubebindv1alpha1 "go.bytebuilders.dev/kube-bind/pkg/apis/kubebind/v1alpha1"
	kubebindhelpers "go.bytebuilders.dev/kube-bind/pkg/apis/kubebind/v1alpha1/helpers"
	konnectormodels "go.bytebuilders.dev/kube-bind/pkg/konnector/models"
)

type reconciler struct {
	providerInfos []*konnectormodels.ProviderInfo

	reconcileServiceBinding func(binding *kubebindv1alpha1.APIServiceBinding) bool
	getServiceExport        func(provider *konnectormodels.ProviderInfo, ns string) (*kubebindv1alpha1.APIServiceExport, error)
	getServiceBinding       func(name string) (*kubebindv1alpha1.APIServiceBinding, error)
	getClusterBinding       func(ctx context.Context, provider *konnectormodels.ProviderInfo) (*kubebindv1alpha1.ClusterBinding, error)

	updateServiceExportStatus func(ctx context.Context, export *kubebindv1alpha1.APIServiceExport, clusterID string) (*kubebindv1alpha1.APIServiceExport, error)

	getCRD    func(name string) (*apiextensionsv1.CustomResourceDefinition, error)
	updateCRD func(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition) (*apiextensionsv1.CustomResourceDefinition, error)
	createCRD func(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition) (*apiextensionsv1.CustomResourceDefinition, error)
}

func (r *reconciler) reconcile(ctx context.Context, binding *kubebindv1alpha1.APIServiceBinding) error {
	var errs []error

	// As konnector is running APIServiceBinding controller for each provider cluster,
	// so each controller should skip others provider's APIServiceBinding object
	if !r.reconcileServiceBinding(binding) {
		return nil
	}

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
	for _, provider := range r.providerInfos {
		if _, err := r.getServiceExport(provider, binding.Name); err != nil && !errors.IsNotFound(err) {
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
	}

	conditions.MarkTrue(
		binding,
		kubebindv1alpha1.APIServiceBindingConditionConnected,
	)

	return nil
}

func (r *reconciler) ensureCRDs(ctx context.Context, binding *kubebindv1alpha1.APIServiceBinding) error {
	var errs []error

	for _, provider := range r.providerInfos {
		export, err := r.getServiceExport(provider, binding.Name)
		if err != nil && !errors.IsNotFound(err) {
			return err
		} else if errors.IsNotFound(err) {
			conditions.MarkFalse(
				binding,
				kubebindv1alpha1.APIServiceBindingConditionConnected,
				"APIServiceExportNotFound",
				conditionsapi.ConditionSeverityError,
				"APIServiceExport %s not found on the service provider cluster %s.",
				binding.Name,
				provider.ClusterID,
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
			Controller: ptr.To(true),
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
		if !kubebindhelpers.IsOwnedByBinding(binding.Name, binding.UID, existing.OwnerReferences) {
			conditions.MarkFalse(
				binding,
				kubebindv1alpha1.APIServiceBindingConditionConnected,
				"ForeignCustomResourceDefinition",
				conditionsapi.ConditionSeverityError,
				"CustomResourceDefinition %s is not owned by kube-bind.appscode.com.",
				binding.Name,
			)
			return nil
		}

		crd.ObjectMeta = existing.ObjectMeta
		if _, err := r.updateCRD(ctx, crd); err != nil && !errors.IsInvalid(err) {
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
	}

	conditions.MarkTrue(binding, kubebindv1alpha1.APIServiceBindingConditionConnected)

	return utilerrors.NewAggregate(errs)
}

func (r *reconciler) ensurePrettyName(ctx context.Context, binding *kubebindv1alpha1.APIServiceBinding) error {
	binding.Status.Providers = []kubebindv1alpha1.Provider{}
	for _, provider := range r.providerInfos {
		clusterBinding, err := r.getClusterBinding(ctx, provider)
		if err != nil && !errors.IsNotFound(err) {
			return err
		} else if errors.IsNotFound(err) {
			return nil
		}
		prov := kubebindv1alpha1.Provider{}
		prov.PrettyName = clusterBinding.Spec.ProviderPrettyName
		binding.Status.Providers = append(binding.Status.Providers, prov)
	}

	return nil
}

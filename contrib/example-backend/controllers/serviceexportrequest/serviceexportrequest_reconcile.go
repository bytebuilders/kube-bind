/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package serviceexportrequest

import (
	"context"
	"time"

	"go.bytebuilders.dev/kube-bind/apis/kubebind/v1alpha1"
	"go.bytebuilders.dev/kube-bind/apis/kubebind/v1alpha1/helpers"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	conditionsapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/conditions"
)

type reconciler struct {
	informerScope          v1alpha1.Scope
	clusterScopedIsolation v1alpha1.Isolation

	getCRD              func(name string) (*apiextensionsv1.CustomResourceDefinition, error)
	getServiceExport    func(ns, name string) (*v1alpha1.APIServiceExport, error)
	createServiceExport func(ctx context.Context, resource *v1alpha1.APIServiceExport) (*v1alpha1.APIServiceExport, error)

	deleteServiceExportRequest func(ctx context.Context, namespace, name string) error
}

func (r *reconciler) reconcile(ctx context.Context, req *v1alpha1.APIServiceExportRequest) error {
	var errs []error

	if err := r.ensureExports(ctx, req); err != nil {
		errs = append(errs, err)
	}

	conditions.SetSummary(req)

	return utilerrors.NewAggregate(errs)
}

func (r *reconciler) ensureExports(ctx context.Context, req *v1alpha1.APIServiceExportRequest) error {
	logger := klog.FromContext(ctx)

	if req.Status.Phase == v1alpha1.APIServiceExportRequestPhasePending {
		failure := false
		for _, res := range req.Spec.Resources {
			name := res.Resource + "." + res.Group
			crd, err := r.getCRD(name)
			if err != nil && !apierrors.IsNotFound(err) {
				return err
			}
			if apierrors.IsNotFound(err) {
				conditions.MarkFalse(
					req,
					v1alpha1.APIServiceExportRequestConditionExportsReady,
					"CRDNotFound",
					conditionsapi.ConditionSeverityError,
					"CustomResourceDefinition %s in the service provider cluster not found",
					name,
				)
				failure = true
				break
			}

			if _, err := r.getServiceExport(req.Namespace, name); err != nil && !apierrors.IsNotFound(err) {
				return err
			} else if err == nil {
				continue
			}

			exportSpec, err := helpers.CRDToServiceExport(crd)
			if err != nil {
				conditions.MarkFalse(
					req,
					v1alpha1.APIServiceExportRequestConditionExportsReady,
					"CRDInvalid",
					conditionsapi.ConditionSeverityError,
					"CustomResourceDefinition %s cannot be converted to a APIServiceExport: %v",
					name,
					err,
				)
				failure = true
				break
			}
			hash := helpers.APIServiceExportCRDSpecHash(exportSpec)
			export := &v1alpha1.APIServiceExport{
				ObjectMeta: metav1.ObjectMeta{
					Name:      crd.Name,
					Namespace: req.Namespace,
					Annotations: map[string]string{
						v1alpha1.SourceSpecHashAnnotationKey: hash,
					},
				},
				Spec: v1alpha1.APIServiceExportSpec{
					APIServiceExportCRDSpec: *exportSpec,
					InformerScope:           r.informerScope,
				},
			}
			if exportSpec.Scope == apiextensionsv1.ClusterScoped {
				export.Spec.ClusterScopedIsolation = r.clusterScopedIsolation
			}

			logger.V(1).Info("Creating APIServiceExport", "name", export.Name, "namespace", export.Namespace)
			if _, err = r.createServiceExport(ctx, export); err != nil {
				return err
			}
		}

		if !failure {
			conditions.MarkTrue(req, v1alpha1.APIServiceExportRequestConditionExportsReady)
			req.Status.Phase = v1alpha1.APIServiceExportRequestPhaseSucceeded
			return nil
		}

		if time.Since(req.CreationTimestamp.Time) > time.Minute {
			req.Status.Phase = v1alpha1.APIServiceExportRequestPhaseFailed
			req.Status.TerminalMessage = conditions.GetMessage(req, v1alpha1.APIServiceExportRequestConditionExportsReady)
		}

		return nil
	}

	if time.Since(req.CreationTimestamp.Time) > 10*time.Minute {
		logger.Info("Deleting service binding request %s/%s", req.Namespace, req.Name, "reason", "timeout", "age", time.Since(req.CreationTimestamp.Time))
		return r.deleteServiceExportRequest(ctx, req.Namespace, req.Name)
	}

	return nil
}

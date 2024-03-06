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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "go.bytebuilders.dev/kube-bind/apis/kubebind/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// APIServiceExportLister helps list APIServiceExports.
// All objects returned here must be treated as read-only.
type APIServiceExportLister interface {
	// List lists all APIServiceExports in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.APIServiceExport, err error)
	// APIServiceExports returns an object that can list and get APIServiceExports.
	APIServiceExports(namespace string) APIServiceExportNamespaceLister
	APIServiceExportListerExpansion
}

// aPIServiceExportLister implements the APIServiceExportLister interface.
type aPIServiceExportLister struct {
	indexer cache.Indexer
}

// NewAPIServiceExportLister returns a new APIServiceExportLister.
func NewAPIServiceExportLister(indexer cache.Indexer) APIServiceExportLister {
	return &aPIServiceExportLister{indexer: indexer}
}

// List lists all APIServiceExports in the indexer.
func (s *aPIServiceExportLister) List(selector labels.Selector) (ret []*v1alpha1.APIServiceExport, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.APIServiceExport))
	})
	return ret, err
}

// APIServiceExports returns an object that can list and get APIServiceExports.
func (s *aPIServiceExportLister) APIServiceExports(namespace string) APIServiceExportNamespaceLister {
	return aPIServiceExportNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// APIServiceExportNamespaceLister helps list and get APIServiceExports.
// All objects returned here must be treated as read-only.
type APIServiceExportNamespaceLister interface {
	// List lists all APIServiceExports in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.APIServiceExport, err error)
	// Get retrieves the APIServiceExport from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.APIServiceExport, error)
	APIServiceExportNamespaceListerExpansion
}

// aPIServiceExportNamespaceLister implements the APIServiceExportNamespaceLister
// interface.
type aPIServiceExportNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all APIServiceExports in the indexer for a given namespace.
func (s aPIServiceExportNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.APIServiceExport, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.APIServiceExport))
	})
	return ret, err
}

// Get retrieves the APIServiceExport from the indexer for a given namespace and name.
func (s aPIServiceExportNamespaceLister) Get(name string) (*v1alpha1.APIServiceExport, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("apiserviceexport"), name)
	}
	return obj.(*v1alpha1.APIServiceExport), nil
}

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

// ClusterBindingLister helps list ClusterBindings.
// All objects returned here must be treated as read-only.
type ClusterBindingLister interface {
	// List lists all ClusterBindings in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ClusterBinding, err error)
	// ClusterBindings returns an object that can list and get ClusterBindings.
	ClusterBindings(namespace string) ClusterBindingNamespaceLister
	ClusterBindingListerExpansion
}

// clusterBindingLister implements the ClusterBindingLister interface.
type clusterBindingLister struct {
	indexer cache.Indexer
}

// NewClusterBindingLister returns a new ClusterBindingLister.
func NewClusterBindingLister(indexer cache.Indexer) ClusterBindingLister {
	return &clusterBindingLister{indexer: indexer}
}

// List lists all ClusterBindings in the indexer.
func (s *clusterBindingLister) List(selector labels.Selector) (ret []*v1alpha1.ClusterBinding, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ClusterBinding))
	})
	return ret, err
}

// ClusterBindings returns an object that can list and get ClusterBindings.
func (s *clusterBindingLister) ClusterBindings(namespace string) ClusterBindingNamespaceLister {
	return clusterBindingNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ClusterBindingNamespaceLister helps list and get ClusterBindings.
// All objects returned here must be treated as read-only.
type ClusterBindingNamespaceLister interface {
	// List lists all ClusterBindings in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ClusterBinding, err error)
	// Get retrieves the ClusterBinding from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.ClusterBinding, error)
	ClusterBindingNamespaceListerExpansion
}

// clusterBindingNamespaceLister implements the ClusterBindingNamespaceLister
// interface.
type clusterBindingNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ClusterBindings in the indexer for a given namespace.
func (s clusterBindingNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.ClusterBinding, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ClusterBinding))
	})
	return ret, err
}

// Get retrieves the ClusterBinding from the indexer for a given namespace and name.
func (s clusterBindingNamespaceLister) Get(name string) (*v1alpha1.ClusterBinding, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("clusterbinding"), name)
	}
	return obj.(*v1alpha1.ClusterBinding), nil
}

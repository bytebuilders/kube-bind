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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	kubebindv1alpha1 "go.bytebuilders.dev/kube-bind/apis/kubebind/v1alpha1"
	versioned "go.bytebuilders.dev/kube-bind/client/clientset/versioned"
	internalinterfaces "go.bytebuilders.dev/kube-bind/client/informers/externalversions/internalinterfaces"
	v1alpha1 "go.bytebuilders.dev/kube-bind/client/listers/kubebind/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// APIServiceBindingInformer provides access to a shared informer and lister for
// APIServiceBindings.
type APIServiceBindingInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.APIServiceBindingLister
}

type aPIServiceBindingInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewAPIServiceBindingInformer constructs a new informer for APIServiceBinding type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewAPIServiceBindingInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredAPIServiceBindingInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredAPIServiceBindingInformer constructs a new informer for APIServiceBinding type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredAPIServiceBindingInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubeBindV1alpha1().APIServiceBindings().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubeBindV1alpha1().APIServiceBindings().Watch(context.TODO(), options)
			},
		},
		&kubebindv1alpha1.APIServiceBinding{},
		resyncPeriod,
		indexers,
	)
}

func (f *aPIServiceBindingInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredAPIServiceBindingInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *aPIServiceBindingInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&kubebindv1alpha1.APIServiceBinding{}, f.defaultInformer)
}

func (f *aPIServiceBindingInformer) Lister() v1alpha1.APIServiceBindingLister {
	return v1alpha1.NewAPIServiceBindingLister(f.Informer().GetIndexer())
}

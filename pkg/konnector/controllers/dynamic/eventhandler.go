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

package dynamic

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/client-go/tools/cache"
)

type SharedIndexInformer interface {
	// AddDynamicEventHandler adds a dynamic event handler to the informer. It's
	// like AddEventHandler, but the handler is removed when the context closes.
	// handlerName must be unique for each handler.
	AddDynamicEventHandler(ctx context.Context, handlerName string, handler cache.ResourceEventHandler)

	// AddEventHandler shadows the method in the embedded SharedIndexInformer. But it
	// will panic and should not be called.
	AddEventHandler(handler cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error)

	cache.SharedIndexInformer
}

type Informer[L any] interface {
	Informer() SharedIndexInformer
	Lister() L
}

type StaticInformer[L any] interface {
	Informer() cache.SharedIndexInformer
	Lister() L
}

type dynamicInformer[L any] struct {
	StaticInformer[L]
	sharedIndexInformer dynamicSharedIndexInformer
}

type dynamicSharedIndexInformer struct {
	cache.SharedIndexInformer

	lock     sync.RWMutex
	counter  int
	handlers map[string]cache.ResourceEventHandler
}

// NewDynamicInformer returns a shared informer that allows adding and removing event
// handlers dynamically.
func NewDynamicInformer[L any](informer StaticInformer[L]) Informer[L] {
	di := &dynamicInformer[L]{
		StaticInformer: informer,
		sharedIndexInformer: dynamicSharedIndexInformer{
			SharedIndexInformer: informer.Informer(),
			handlers:            make(map[string]cache.ResourceEventHandler),
		},
	}

	_, err := informer.Informer().AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInInitialList bool) {
			di.sharedIndexInformer.lock.RLock()
			defer di.sharedIndexInformer.lock.RUnlock()
			for _, h := range di.sharedIndexInformer.handlers {
				h.OnAdd(obj, isInInitialList)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			di.sharedIndexInformer.lock.RLock()
			defer di.sharedIndexInformer.lock.RUnlock()
			for _, h := range di.sharedIndexInformer.handlers {
				h.OnUpdate(oldObj, newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			di.sharedIndexInformer.lock.RLock()
			defer di.sharedIndexInformer.lock.RUnlock()
			for _, h := range di.sharedIndexInformer.handlers {
				h.OnDelete(obj)
			}
		},
	})
	if err != nil {
		panic(err)
	}

	return di
}

func (i *dynamicInformer[L]) Informer() SharedIndexInformer {
	return &i.sharedIndexInformer
}

func (i *dynamicInformer[L]) Lister() L {
	return i.StaticInformer.Lister()
}

func (i *dynamicSharedIndexInformer) AddDynamicEventHandler(ctx context.Context, handlerName string, handler cache.ResourceEventHandler) {
	i.lock.Lock()
	handlerName = fmt.Sprintf("%s-%d", handlerName, i.counter)
	i.handlers[handlerName] = handler
	i.counter++ // make unique
	i.lock.Unlock()

	go func() {
		<-ctx.Done()
		i.lock.Lock()
		defer i.lock.Unlock()
		delete(i.handlers, handlerName)
	}()

	// simulate initial add events for an informer that is already started.
	objs := i.GetStore().List()
	for _, obj := range objs {
		handler.OnAdd(obj, true)
	}
}

func (i *dynamicSharedIndexInformer) AddEventHandler(handler cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error) {
	panic("call AddDynamicEventHandler instead")
}

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "go.bytebuilders.dev/kube-bind/apis/kubebind/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClusterBindings implements ClusterBindingInterface
type FakeClusterBindings struct {
	Fake *FakeKubeBindV1alpha1
	ns   string
}

var clusterbindingsResource = v1alpha1.SchemeGroupVersion.WithResource("clusterbindings")

var clusterbindingsKind = v1alpha1.SchemeGroupVersion.WithKind("ClusterBinding")

// Get takes name of the clusterBinding, and returns the corresponding clusterBinding object, and an error if there is any.
func (c *FakeClusterBindings) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ClusterBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(clusterbindingsResource, c.ns, name), &v1alpha1.ClusterBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterBinding), err
}

// List takes label and field selectors, and returns the list of ClusterBindings that match those selectors.
func (c *FakeClusterBindings) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ClusterBindingList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(clusterbindingsResource, clusterbindingsKind, c.ns, opts), &v1alpha1.ClusterBindingList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ClusterBindingList{ListMeta: obj.(*v1alpha1.ClusterBindingList).ListMeta}
	for _, item := range obj.(*v1alpha1.ClusterBindingList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clusterBindings.
func (c *FakeClusterBindings) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(clusterbindingsResource, c.ns, opts))

}

// Create takes the representation of a clusterBinding and creates it.  Returns the server's representation of the clusterBinding, and an error, if there is any.
func (c *FakeClusterBindings) Create(ctx context.Context, clusterBinding *v1alpha1.ClusterBinding, opts v1.CreateOptions) (result *v1alpha1.ClusterBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(clusterbindingsResource, c.ns, clusterBinding), &v1alpha1.ClusterBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterBinding), err
}

// Update takes the representation of a clusterBinding and updates it. Returns the server's representation of the clusterBinding, and an error, if there is any.
func (c *FakeClusterBindings) Update(ctx context.Context, clusterBinding *v1alpha1.ClusterBinding, opts v1.UpdateOptions) (result *v1alpha1.ClusterBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(clusterbindingsResource, c.ns, clusterBinding), &v1alpha1.ClusterBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterBinding), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeClusterBindings) UpdateStatus(ctx context.Context, clusterBinding *v1alpha1.ClusterBinding, opts v1.UpdateOptions) (*v1alpha1.ClusterBinding, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(clusterbindingsResource, "status", c.ns, clusterBinding), &v1alpha1.ClusterBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterBinding), err
}

// Delete takes name of the clusterBinding and deletes it. Returns an error if one occurs.
func (c *FakeClusterBindings) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(clusterbindingsResource, c.ns, name, opts), &v1alpha1.ClusterBinding{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeClusterBindings) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(clusterbindingsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ClusterBindingList{})
	return err
}

// Patch applies the patch and returns the patched clusterBinding.
func (c *FakeClusterBindings) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ClusterBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(clusterbindingsResource, c.ns, name, pt, data, subresources...), &v1alpha1.ClusterBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterBinding), err
}

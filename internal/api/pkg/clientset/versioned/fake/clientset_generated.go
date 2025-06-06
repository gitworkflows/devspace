// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	clientset "dev.khulnasoft.com/api/v4/pkg/clientset/versioned"
	managementv1 "dev.khulnasoft.com/api/v4/pkg/clientset/versioned/typed/management/v1"
	fakemanagementv1 "dev.khulnasoft.com/api/v4/pkg/clientset/versioned/typed/management/v1/fake"
	storagev1 "dev.khulnasoft.com/api/v4/pkg/clientset/versioned/typed/storage/v1"
	fakestoragev1 "dev.khulnasoft.com/api/v4/pkg/clientset/versioned/typed/storage/v1/fake"
	virtualclusterv1 "dev.khulnasoft.com/api/v4/pkg/clientset/versioned/typed/virtualcluster/v1"
	fakevirtualclusterv1 "dev.khulnasoft.com/api/v4/pkg/clientset/versioned/typed/virtualcluster/v1/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/testing"
)

// NewSimpleClientset returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any field management, validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
//
// DEPRECATED: NewClientset replaces this with support for field management, which significantly improves
// server side apply testing. NewClientset is only available when apply configurations are generated (e.g.
// via --with-applyconfig).
func NewSimpleClientset(objects ...runtime.Object) *Clientset {
	o := testing.NewObjectTracker(scheme, codecs.UniversalDecoder())
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	cs := &Clientset{tracker: o}
	cs.discovery = &fakediscovery.FakeDiscovery{Fake: &cs.Fake}
	cs.AddReactor("*", "*", testing.ObjectReaction(o))
	cs.AddWatchReactor("*", func(action testing.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		watch, err := o.Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		return true, watch, nil
	})

	return cs
}

// Clientset implements clientset.Interface. Meant to be embedded into a
// struct to get a default implementation. This makes faking out just the method
// you want to test easier.
type Clientset struct {
	testing.Fake
	discovery *fakediscovery.FakeDiscovery
	tracker   testing.ObjectTracker
}

func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	return c.discovery
}

func (c *Clientset) Tracker() testing.ObjectTracker {
	return c.tracker
}

var (
	_ clientset.Interface = &Clientset{}
	_ testing.FakeClient  = &Clientset{}
)

// ManagementV1 retrieves the ManagementV1Client
func (c *Clientset) ManagementV1() managementv1.ManagementV1Interface {
	return &fakemanagementv1.FakeManagementV1{Fake: &c.Fake}
}

// StorageV1 retrieves the StorageV1Client
func (c *Clientset) StorageV1() storagev1.StorageV1Interface {
	return &fakestoragev1.FakeStorageV1{Fake: &c.Fake}
}

// VirtualclusterV1 retrieves the VirtualclusterV1Client
func (c *Clientset) VirtualclusterV1() virtualclusterv1.VirtualclusterV1Interface {
	return &fakevirtualclusterv1.FakeVirtualclusterV1{Fake: &c.Fake}
}

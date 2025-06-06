// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "dev.khulnasoft.com/api/v4/pkg/apis/management/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
)

// ConfigLister helps list Configs.
// All objects returned here must be treated as read-only.
type ConfigLister interface {
	// List lists all Configs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.Config, err error)
	// Get retrieves the Config from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.Config, error)
	ConfigListerExpansion
}

// configLister implements the ConfigLister interface.
type configLister struct {
	listers.ResourceIndexer[*v1.Config]
}

// NewConfigLister returns a new ConfigLister.
func NewConfigLister(indexer cache.Indexer) ConfigLister {
	return &configLister{listers.New[*v1.Config](indexer, v1.Resource("config"))}
}

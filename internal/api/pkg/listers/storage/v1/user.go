// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "dev.khulnasoft.com/api/v4/pkg/apis/storage/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
)

// UserLister helps list Users.
// All objects returned here must be treated as read-only.
type UserLister interface {
	// List lists all Users in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.User, err error)
	// Get retrieves the User from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.User, error)
	UserListerExpansion
}

// userLister implements the UserLister interface.
type userLister struct {
	listers.ResourceIndexer[*v1.User]
}

// NewUserLister returns a new UserLister.
func NewUserLister(indexer cache.Indexer) UserLister {
	return &userLister{listers.New[*v1.User](indexer, v1.Resource("user"))}
}

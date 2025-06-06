// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	internalinterfaces "dev.khulnasoft.com/agentapi/v4/pkg/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// ClusterQuotas returns a ClusterQuotaInformer.
	ClusterQuotas() ClusterQuotaInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// ClusterQuotas returns a ClusterQuotaInformer.
func (v *version) ClusterQuotas() ClusterQuotaInformer {
	return &clusterQuotaInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

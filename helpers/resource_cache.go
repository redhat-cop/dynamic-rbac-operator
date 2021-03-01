package helpers

import (
	"sync"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
)

var lock = &sync.Mutex{}

// ResourceCache holds information about the kube cluster state and
// its policies so that it doesn't need to be queried for every reconciliation.
type ResourceCache struct {
	CRDs                map[string]string
	AllPolicies         *[]rbacv1.PolicyRule
	WatchedRoles        map[types.NamespacedName]bool
	WatchedClusterRoles map[types.NamespacedName]bool
}

var instance *ResourceCache

// GetCacheInstance returns or instantiates a ResourceCache
func GetCacheInstance() *ResourceCache {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			instance = &ResourceCache{}
			instance.CRDs = map[string]string{}
			instance.WatchedRoles = map[types.NamespacedName]bool{}
			instance.WatchedClusterRoles = map[types.NamespacedName]bool{}
		}
	}
	return instance
}

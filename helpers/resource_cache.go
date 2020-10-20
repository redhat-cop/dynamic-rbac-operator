package helpers

import (
	"sync"

	rbacv1 "k8s.io/api/rbac/v1"
)

var lock = &sync.Mutex{}

// ResourceCache holds information about the kube cluster state and
// its policies so that it doesn't need to be queried for every reconciliation.
type ResourceCache struct {
	CRDs        map[string]string
	AllPolicies *[]rbacv1.PolicyRule
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
		}
	}
	return instance
}

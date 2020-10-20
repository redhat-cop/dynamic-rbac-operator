package helpers

import (
	"context"

	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DiscoverClusterResources returns a list of all known resources and groups known to this API server
func DiscoverClusterResources(config *rest.Config) (apiGroupList []*metav1.APIGroup, apiResourceList []*metav1.APIResourceList, err error) {
	if err != nil {
		return nil, nil, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	groups, resources, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return nil, nil, err
	}
	return groups, resources, nil
}

// CreateOrUpdateRole ensures that a role exists in the specified state in the cluster, whether it has to be created or updated to ensure that
func CreateOrUpdateRole(role *v1.Role, c client.Client) (err error) {
	found := &v1.Role{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: role.Name, Namespace: role.Namespace}, found)

	if found != nil && errors.IsNotFound(err) {
		err = c.Create(context.TODO(), role)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	found.Rules = role.Rules
	err = c.Update(context.TODO(), found)
	if err != nil {
		return err
	}

	return nil
}

// CreateOrUpdateClusterRole ensures that a clusterrole exists in the specified state in the cluster, whether it has to be created or updated to ensure that
func CreateOrUpdateClusterRole(role *v1.ClusterRole, c client.Client) (err error) {
	found := &v1.ClusterRole{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: role.Name}, found)

	if found != nil && errors.IsNotFound(err) {
		err = c.Create(context.TODO(), role)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	found.Rules = role.Rules
	err = c.Update(context.TODO(), found)
	if err != nil {
		return err
	}

	return nil
}

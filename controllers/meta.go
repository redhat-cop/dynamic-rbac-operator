package controllers

import (
	"context"

	"github.com/go-logr/logr"
	rbacv1alpha1 "github.com/redhat-cop/dynamic-rbac-operator/api/v1alpha1"
	"github.com/redhat-cop/dynamic-rbac-operator/helpers"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// UpdateAllDynamicResources loops through all DynamicRoles and DynamicClusterRoles and updates their rules/specs as required based on current cache info
func UpdateAllDynamicResources(client client.Client, log logr.Logger, scheme *runtime.Scheme, cache *helpers.ResourceCache) (ctrl.Result, error) {
	// Clear the watched roles cache maps since we're about to recreate them anyway - gets rid of anything we used to care about but no longer need
	cache.WatchedRoles = map[types.NamespacedName]bool{}
	cache.WatchedClusterRoles = map[types.NamespacedName]bool{}

	dynamicRoleList := &rbacv1alpha1.DynamicRoleList{}
	err := client.List(context.TODO(), dynamicRoleList)
	if err != nil {
		log.Error(err, "could not list Dynamic Roles")
		return reconcile.Result{}, err
	}
	for _, dynamicRole := range dynamicRoleList.Items {
		_, err := ReconcileDynamicRole(&dynamicRole, client, scheme, log, cache)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	dynamicClusterRoleList := &rbacv1alpha1.DynamicClusterRoleList{}
	err = client.List(context.TODO(), dynamicClusterRoleList)
	if err != nil {
		log.Error(err, "could not list Dynamic Cluster Roles")
		return reconcile.Result{}, err
	}
	for _, dynamicClusterRole := range dynamicClusterRoleList.Items {
		_, err := ReconcileDynamicClusterRole(&dynamicClusterRole, client, scheme, log, cache)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	log.Info("All computed roles have been reconciled")
	return reconcile.Result{}, nil
}

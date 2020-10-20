/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rbacv1alpha1 "github.com/redhat-cop/dynamic-rbac-operator/api/v1alpha1"
	helpers "github.com/redhat-cop/dynamic-rbac-operator/helpers"
)

// CustomResourceDefinitionReconciler reconciles a CustomResourceDefinition object
type CustomResourceDefinitionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rbac.redhatcop.redhat.io,resources=dynamicroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.redhatcop.redhat.io,resources=dynamicroles/status,verbs=get;update;patch

func (r *CustomResourceDefinitionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("dynamicrole", req.NamespacedName)

	cache := helpers.GetCacheInstance()

	instance := &crdv1beta1.CustomResourceDefinition{}
	crdWasDeleted := false
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			crdWasDeleted = true
		} else {
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}
	}

	if !crdWasDeleted {
		crdJSON, _ := json.Marshal(instance.Spec)
		if _, ok := cache.CRDs[instance.Name]; ok {
			cache.CRDs[instance.Name] = string(crdJSON)
			r.Log.Info("CRD is in cache - reconciliation is not required")
			return reconcile.Result{}, nil
		}
		cache.CRDs[instance.Name] = string(crdJSON)
		r.Log.Info("CRD not previously seen - reconciliation of all computed roles is required")
	} else {
		delete(cache.CRDs, req.Name)
		r.Log.Info("CRD deleted - reconciliation of all computed roles is required")
	}

	// Refresh policy cache
	config, err := ctrl.GetConfig()
	if err != nil {
		return reconcile.Result{}, err
	}
	_, apiResourceList, err := helpers.DiscoverClusterResources(config)
	if err != nil {
		return reconcile.Result{}, err
	}
	allPossibleRules := helpers.APIResourcesToExpandedRules(apiResourceList)
	cache.AllPolicies = &allPossibleRules
	r.Log.Info("Rebuilt cluster policy cache")

	dynamicRoleList := &rbacv1alpha1.DynamicRoleList{}
	err = r.Client.List(context.TODO(), dynamicRoleList)
	if err != nil {
		r.Log.Error(err, "could not list Dynamic Roles")
		return reconcile.Result{}, err
	}
	for _, dynamicRole := range dynamicRoleList.Items {
		_, err := ReconcileDynamicRole(&dynamicRole, r.Client, r.Scheme, r.Log)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	dynamicClusterRoleList := &rbacv1alpha1.DynamicClusterRoleList{}
	err = r.Client.List(context.TODO(), dynamicClusterRoleList)
	if err != nil {
		r.Log.Error(err, "could not list Dynamic Cluster Roles")
		return reconcile.Result{}, err
	}
	for _, dynamicClusterRole := range dynamicClusterRoleList.Items {
		_, err := ReconcileDynamicClusterRole(&dynamicClusterRole, r.Client, r.Scheme, r.Log)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	r.Log.Info("All computed roles have been reconciled")

	return ctrl.Result{}, nil
}

func (r *CustomResourceDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1beta1.CustomResourceDefinition{}).
		Complete(r)
}

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

	"github.com/go-logr/logr"
	"github.com/redhat-cop/dynamic-rbac-operator/helpers"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rbacv1 "k8s.io/api/rbac/v1"
)

// ClusterRoleReconciler reconciles a ClusterRole object
type ClusterRoleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Cache  *helpers.ResourceCache
}

// +kubebuilder:rbac:groups=rbac.redhatcop.redhat.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.redhatcop.redhat.io,resources=clusterroles/status,verbs=get;update;patch

func (r *ClusterRoleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("clusterrole", req.NamespacedName)

	result := ctrl.Result{}
	var err error

	if _, exists := r.Cache.WatchedClusterRoles[req.NamespacedName]; exists {
		r.Log.Info("A cluster role referenced by a dynamic resource has been updated - reconciling now")
		result, err = UpdateAllDynamicResources(r.Client, r.Log, r.Scheme, r.Cache)
	}

	return result, err
}

func (r *ClusterRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacv1.ClusterRole{}).
		Complete(r)
}

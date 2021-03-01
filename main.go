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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	helpers "github.com/redhat-cop/dynamic-rbac-operator/helpers"

	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	rbacv1alpha1 "github.com/redhat-cop/dynamic-rbac-operator/api/v1alpha1"
	"github.com/redhat-cop/dynamic-rbac-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(crdv1beta1.AddToScheme(scheme))

	utilruntime.Must(rbacv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "53201349.redhatcop.redhat.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	cache := helpers.GetCacheInstance()

	if err = (&controllers.CustomResourceDefinitionReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CustomResourceDefinition"),
		Scheme: mgr.GetScheme(),
		Cache:  cache,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CustomResourceDefinition")
		os.Exit(1)
	}
	if err = (&controllers.DynamicRoleReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("DynamicRole"),
		Scheme: mgr.GetScheme(),
		Cache:  cache,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DynamicRole")
		os.Exit(1)
	}
	if err = (&controllers.DynamicClusterRoleReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("DynamicClusterRole"),
		Scheme: mgr.GetScheme(),
		Cache:  cache,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DynamicClusterRole")
		os.Exit(1)
	}
	if err = (&controllers.RoleReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Role"),
		Scheme: mgr.GetScheme(),
		Cache:  cache,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Role")
		os.Exit(1)
	}
	if err = (&controllers.ClusterRoleReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ClusterRole"),
		Scheme: mgr.GetScheme(),
		Cache:  cache,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterRole")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	// Begin cache setup
	setupLog.Info("Performing pre-controller setup")
	restConfig, err := ctrl.GetConfig()
	if err != nil {
		setupLog.Error(err, "could not instantiate a client for pre-controller setup processes")
		os.Exit(1)
	}
	client, err := client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		setupLog.Error(err, "could not instantiate a client for pre-controller setup processes")
		os.Exit(1)
	}
	crdList := &crdv1beta1.CustomResourceDefinitionList{}
	err = client.List(context.TODO(), crdList)
	for _, crd := range crdList.Items {
		crdJSON, _ := json.Marshal(crd.Spec)
		cache.CRDs[crd.Name] = string(crdJSON)
		setupLog.Info(fmt.Sprintf("Added %s to the CRD cache", crd.Name))
	}
	if err != nil {
		setupLog.Error(err, "could not build the CRD cache in the pre-controller setup phase")
		os.Exit(1)
	}
	_, apiResourceList, err := helpers.DiscoverClusterResources(restConfig)
	if err != nil {
		setupLog.Error(err, "could not build the cluster policy cache in the pre-controller setup phase")
		os.Exit(1)
	}
	allPossibleRules := helpers.APIResourcesToExpandedRules(apiResourceList)
	cache.AllPolicies = &allPossibleRules
	setupLog.Info("Successfully built the cluster policy cache")
	setupLog.Info("Pre-controller setup is complete")
	// End cache setup

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

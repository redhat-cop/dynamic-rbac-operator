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

package v1alpha1

import (
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DynamicClusterRoleSpec defines the desired state of DynamicClusterRole
type DynamicClusterRoleSpec struct {
	Inherit *[]InheritedRole `json:"inherit,omitempty"`
	Allow   *[]v1.PolicyRule `json:"allow,omitempty"`
	Deny    *[]v1.PolicyRule `json:"deny,omitempty"`
}

// DynamicClusterRoleStatus defines the observed state of DynamicClusterRole
type DynamicClusterRoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// DynamicClusterRole is the Schema for the dynamicclusterroles API
type DynamicClusterRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DynamicClusterRoleSpec   `json:"spec,omitempty"`
	Status DynamicClusterRoleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DynamicClusterRoleList contains a list of DynamicClusterRole
type DynamicClusterRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DynamicClusterRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DynamicClusterRole{}, &DynamicClusterRoleList{})
}

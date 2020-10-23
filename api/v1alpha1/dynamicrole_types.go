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

// DynamicRoleSpec defines the desired state of DynamicRole
type DynamicRoleSpec struct {
	Inherit *[]InheritedRole `json:"inherit,omitempty"`
	Allow   *[]v1.PolicyRule `json:"allow,omitempty"`
	Deny    *[]v1.PolicyRule `json:"deny,omitempty"`
}

type InheritedRole struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace,omitempty"`
}

// DynamicRoleStatus defines the observed state of DynamicRole
type DynamicRoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DynamicRole is the Schema for the dynamicroles API
type DynamicRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DynamicRoleSpec   `json:"spec,omitempty"`
	Status DynamicRoleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DynamicRoleList contains a list of DynamicRole
type DynamicRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DynamicRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DynamicRole{}, &DynamicRoleList{})
}

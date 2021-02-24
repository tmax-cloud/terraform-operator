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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AWSSecurityGroupRuleSpec defines the desired state of AWSSecurityGroupRule
type AWSSecurityGroupRuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AWSSecurityGroupRule. Edit AWSSecurityGroupRule_types.go to remove/update
	Provider string `json:"provider,omitempty"`
	SG       string `json:"sg,omitempty"`
	Type     string `json:"type,omitempty"`
	FromPort string `json:"fromport,omitempty"`
	ToPort   string `json:"toport,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	CIDR     string `json:"cidr,omitempty"`
}

// AWSSecurityGroupRuleStatus defines the observed state of AWSSecurityGroupRule
type AWSSecurityGroupRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes,omitempty"`
	Phase string   `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AWSSecurityGroupRule is the Schema for the awssecuritygrouprules API
type AWSSecurityGroupRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSSecurityGroupRuleSpec   `json:"spec,omitempty"`
	Status AWSSecurityGroupRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AWSSecurityGroupRuleList contains a list of AWSSecurityGroupRule
type AWSSecurityGroupRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSSecurityGroupRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSSecurityGroupRule{}, &AWSSecurityGroupRuleList{})
}

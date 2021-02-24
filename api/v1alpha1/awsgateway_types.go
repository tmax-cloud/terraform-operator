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

// AWSGatewaySpec defines the desired state of AWSGateway
type AWSGatewaySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AWSGateway. Edit AWSGateway_types.go to remove/update
	Provider string `json:"provider,omitempty"`
	VPC      string `json:"vpc,omitempty"`
}

// AWSGatewayStatus defines the observed state of AWSGateway
type AWSGatewayStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes,omitempty"`
	Phase string   `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AWSGateway is the Schema for the awsgateways API
type AWSGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSGatewaySpec   `json:"spec,omitempty"`
	Status AWSGatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AWSGatewayList contains a list of AWSGateway
type AWSGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSGateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSGateway{}, &AWSGatewayList{})
}

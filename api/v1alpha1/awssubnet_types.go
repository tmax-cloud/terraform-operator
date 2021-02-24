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

// AWSSubnetSpec defines the desired state of AWSSubnet
type AWSSubnetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AWSSubnet. Edit AWSSubnet_types.go to remove/update
	Provider string `json:"provider,omitempty"`
	VPC      string `json:"vpc,omitempty"`
	CIDR     string `json:"cidr,omitempty"`
	Zone     string `json:"zone,omitempty"`
}

// AWSSubnetStatus defines the observed state of AWSSubnet
type AWSSubnetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes,omitempty"`
	Phase string   `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AWSSubnet is the Schema for the awssubnets API
type AWSSubnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSSubnetSpec   `json:"spec,omitempty"`
	Status AWSSubnetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AWSSubnetList contains a list of AWSSubnet
type AWSSubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSSubnet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSSubnet{}, &AWSSubnetList{})
}

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

// ResourceSpec defines the desired state of Resource
type ResourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Resource. Edit Resource_types.go to remove/update
	Provider    string      `json:"provider,omitempty"`
	Type        string      `json:"type,omitempty"`
	AWS_VPC     AWS_VPC     `json:"aws_vpc,omitempty"`
	AWS_SUBNET  AWS_SUBNET  `json:"aws_subnet,omitempty"`
	AWS_GATEWAY AWS_GATEWAY `json:"aws_gateway,omitempty"`
	AWS_ROUTE   AWS_ROUTE   `json:"aws_route,omitempty"`
	AWS_SG      AWS_SG      `json:"aws_sg,omitempty"`
	AWS_SG_RULE AWS_SG_RULE `json:"aws_sg_rule,omitempty"`
}

type AWS_VPC struct {
	Name string `json:"name,omitempty"`
	CIDR string `json:"cidr,omitempty"`
}

type AWS_SUBNET struct {
	VPC  string `json:"vpc,omitempty"`
	Name string `json:"name,omitempty"`
	CIDR string `json:"cidr,omitempty"`
	Zone string `json:"zone,omitempty"`
}

type AWS_GATEWAY struct {
	VPC  string `json:"vpc,omitempty"`
	Name string `json:"name,omitempty"`
}

type AWS_ROUTE struct {
	VPC     string `json:"vpc,omitempty"`
	Subnet  string `json:"subnet,omitempty"`
	Gateway string `json:"gateway,omitempty"`
	Name    string `json:"name,omitempty"`
	CIDR    string `json:"cidr,omitempty"`
}

type AWS_SG struct {
	VPC  string `json:"vpc,omitempty"`
	Name string `json:"name,omitempty"`
}

type AWS_SG_RULE struct {
	SG       string `json:"sg,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	FromPort string `json:"fromport,omitempty"`
	ToPort   string `json:"toport,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	CIDR     string `json:"cidr,omitempty"`
}

// ResourceStatus defines the observed state of Resource
type ResourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes,omitempty"`
	Phase string   `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Resource is the Schema for the resources API
type Resource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceSpec   `json:"spec,omitempty"`
	Status ResourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceList contains a list of Resource
type ResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Resource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Resource{}, &ResourceList{})
}

/*
Copyright 2021.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionReady string = "Ready"

	ConditionReasonSecretTypeIsInvalid string = "Secret type is invalid"
	ConditionReasonSecretIsNotFound    string = "Target secret is not found"
	ConditionReasonTLSKeyNotFound      string = "Target secret does't have TLS key"
	ConditionReasonTLSKeyCanNotParse   string = "TLS key can not parse"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SSLSpec defines the desired state of SSL
type SSLSpec struct {
	SecretName string `json:"secretName,omitempty"`

	// +kubebuilder:validation:Minimum=1
	LimitBefore int `json:"limitBefore,omitempty"`
}

// SSLStatus defines the observed state of SSL
type SSLStatus struct {
	// Conditions is an array of conditions.
	// Known .status.conditions.type are: "Ready"
	//+patchMergeKey=type
	//+patchStrategy=merge
	//+listType=map
	//+listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="SECRETNAME",type="string",JSONPath=".spec.secretName"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"

// SSL is the Schema for the ssls API
type SSL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SSLSpec   `json:"spec,omitempty"`
	Status SSLStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SSLList contains a list of SSL
type SSLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SSL{}, &SSLList{})
}

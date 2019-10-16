package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BookStoreSpec defines the desired state of BookStore
// +k8s:openapi-gen=true
// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
type BookStoreSpec struct {

	BookApp BookApp     `json:"bookApp,omitempty"`
	BookDB  BookDB      `json:"bookDB,omitempty"`
}

type BookApp struct {
	 
	Repository      string `json:"repository,omitempty"`
	Tag             string `json:"tag,omitempty"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
    Replicas        int32  `json:"replicas,omitempty"`
    Port            int32  `json:"port,omitempty"`
	TargetPort      int  `json:"targetPort,omitempty"`
	ServiceType     corev1.ServiceType  `json:"serviceType,omitempty"`
}

type BookDB struct {
	 
	Repository      string `json:"repository,omitempty"`
	Tag             string `json:"tag,omitempty"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
    Replicas        int32  `json:"replicas,omitempty"`
	Port            int32  `json:"port,omitempty"`
	DBSize          resource.Quantity `json:"dbSize,omitempty"`
}

// BookStoreStatus defines the observed state of BookStore
// +k8s:openapi-gen=true
type BookStoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BookStore is the Schema for the bookstores API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=bookstores,scope=Namespaced
type BookStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BookStoreSpec   `json:"spec,omitempty"`
	Status BookStoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BookStoreList contains a list of BookStore
type BookStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BookStore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BookStore{}, &BookStoreList{})
}

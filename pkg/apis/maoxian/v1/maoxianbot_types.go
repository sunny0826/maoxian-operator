package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type RepoStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// MaoxianBotSpec defines the desired state of MaoxianBot
type MaoxianBotSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	RepoList []string `json:"repoList"`
	Plat     string   `json:"plat"`
}

// MaoxianBotStatus defines the observed state of MaoxianBot
type MaoxianBotStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	RepoStatus []RepoStatus `json:"repoStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MaoxianBot is the Schema for the maoxianbots API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=maoxianbots,scope=Namespaced
type MaoxianBot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MaoxianBotSpec   `json:"spec,omitempty"`
	Status MaoxianBotStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MaoxianBotList contains a list of MaoxianBot
type MaoxianBotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MaoxianBot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MaoxianBot{}, &MaoxianBotList{})
}

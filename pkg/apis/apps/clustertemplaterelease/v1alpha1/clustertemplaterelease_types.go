package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope="Cluster",shortName={"mclsetquota","mclsetquotas"}

type ClusterTemplateRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterTemplateReleaseSpec `json:"spec"`
}

type ClusterTemplateReleaseSpec struct {
	HelmReleaseURL string `json:"helmReleaseURL,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterTemplateReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ClusterTemplateRelease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterTemplateRelease{}, &ClusterTemplateReleaseList{})
}

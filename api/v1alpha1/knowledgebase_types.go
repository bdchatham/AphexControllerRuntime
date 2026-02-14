package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const orgNamespacePrefix = "org-"

// Source defines a single knowledge source to track
type Source struct {
	// RepoOrg is the repository organization or owner (e.g., "bdchatham")
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._-]+$`
	RepoOrg string `json:"repoOrg"`

	// RepoName is the repository name (e.g., "AphexControllerRuntime")
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._-]+$`
	RepoName string `json:"repoName"`

	// Branch is the Git branch to track (default: mainline)
	// +kubebuilder:default:="mainline"
	// +optional
	Branch string `json:"branch,omitempty"`

	// SourceType classifies the source as documentation or code
	// +kubebuilder:validation:Enum=docs;code
	// +optional
	SourceType string `json:"sourceType,omitempty"`

	// Paths are the file/path patterns to process
	// Supports glob patterns (e.g., ".kiro/docs" for docs, "src/**" for code)
	// +optional
	Paths []string `json:"paths,omitempty"`
}

// FullName returns the GitHub-style "org/repo" identifier.
func (s *Source) FullName() string {
	return fmt.Sprintf("%s/%s", s.RepoOrg, s.RepoName)
}

// MCPConfig defines the MCP server configuration
// If this field is set (non-nil), an MCP server will be provisioned
type MCPConfig struct {
	// Image is the container image for the MCP server
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Port is the port the MCP server listens on
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1024
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// QueryServiceURL is the URL of the query service
	// If empty, controller computes as http://query.{namespace}:8080
	// +optional
	QueryServiceURL string `json:"queryServiceURL,omitempty"`

	// Replicas is the number of MCP server replicas
	// +kubebuilder:default:=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
}

// KnowledgeBaseSpec defines the desired state of KnowledgeBase
type KnowledgeBaseSpec struct {
	// Name is the human-readable knowledge base name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Description provides context about this knowledge base
	// +optional
	Description string `json:"description,omitempty"`

	// Organization references the name of an Organization resource.
	// The controller derives the org namespace as org-{organization}.
	// +kubebuilder:validation:Required
	Organization string `json:"organization"`

	// Sources is the list of knowledge sources to track
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Sources []Source `json:"sources"`

	// MCP configures the optional MCP server for this knowledge base
	// If set, an MCP server will be provisioned. If nil/omitted, no MCP server is created.
	// +optional
	MCP *MCPConfig `json:"mcp,omitempty"`
}

// MCPStatus defines the observed state of the MCP server
type MCPStatus struct {
	// Deployed indicates whether the MCP server is deployed
	// +optional
	Deployed bool `json:"deployed,omitempty"`

	// ServiceName is the Kubernetes service name for the MCP server
	// +optional
	ServiceName string `json:"serviceName,omitempty"`

	// ServiceURL is the internal cluster URL for the MCP server
	// +optional
	ServiceURL string `json:"serviceURL,omitempty"`

	// ReadyReplicas is the number of ready MCP server replicas
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
}

// KnowledgeBaseStatus defines the observed state of KnowledgeBase
type KnowledgeBaseStatus struct {
	// Phase represents the current state (Pending, Ready, Failed)
	// +optional
	Phase string `json:"phase,omitempty"`

	// Message provides human-readable status information
	// +optional
	Message string `json:"message,omitempty"`

	// LastReconcileTime is the timestamp of the last reconciliation
	// +optional
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`

	// MCP contains the status of the MCP server (if enabled)
	// +optional
	MCP MCPStatus `json:"mcp,omitempty"`

	// VectorStoreReady indicates whether the vector store is healthy
	// +optional
	VectorStoreReady bool `json:"vectorStoreReady,omitempty"`

	// CodeGraphReady indicates whether the code graph is healthy
	// +optional
	CodeGraphReady bool `json:"codeGraphReady,omitempty"`

	// LastSyncTime is the timestamp of the last successful sync pipeline run
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Organization",type=string,JSONPath=`.spec.organization`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Vector Store",type=boolean,JSONPath=`.status.vectorStoreReady`
// +kubebuilder:printcolumn:name="Code Graph",type=boolean,JSONPath=`.status.codeGraphReady`
// +kubebuilder:printcolumn:name="MCP",type=boolean,JSONPath=`.status.mcp.deployed`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// KnowledgeBase is the Schema for the knowledgebases API
type KnowledgeBase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KnowledgeBaseSpec   `json:"spec,omitempty"`
	Status KnowledgeBaseStatus `json:"status,omitempty"`
}

// OrgNamespace derives the organization namespace from the Organization field.
// Returns org-{organization} for use as the Kubernetes namespace.
func (kb *KnowledgeBase) OrgNamespace() string {
	return fmt.Sprintf("%s%s", orgNamespacePrefix, kb.Spec.Organization)
}

// ValidateOrganization checks that the Organization field is non-empty.
// Returns an error when the organization field is empty, nil otherwise.
func (kb *KnowledgeBase) ValidateOrganization() error {
	if kb.Spec.Organization == "" {
		return fmt.Errorf("organization field cannot be empty")
	}
	return nil
}

// +kubebuilder:object:root=true

// KnowledgeBaseList contains a list of KnowledgeBase
type KnowledgeBaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KnowledgeBase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KnowledgeBase{}, &KnowledgeBaseList{})
}

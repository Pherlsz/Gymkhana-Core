package assistant

// RetrievalMode controls when retrieval is used by an Assistant. A definition
// without the retrieval module/policy has retrieval disabled.
type RetrievalMode string

const (
	RetrievalOnDemand RetrievalMode = "on_demand"
	RetrievalAlways   RetrievalMode = "always"
)

func (mode RetrievalMode) Valid() bool {
	switch mode {
	case RetrievalOnDemand, RetrievalAlways:
		return true
	default:
		return false
	}
}

// SearchMode describes the retrieval strategy independent of any vector/search
// product implementation.
type SearchMode string

const (
	SearchLexical SearchMode = "lexical"
	SearchVector  SearchMode = "vector"
	SearchHybrid  SearchMode = "hybrid"
)

func (mode SearchMode) Valid() bool {
	switch mode {
	case SearchLexical, SearchVector, SearchHybrid:
		return true
	default:
		return false
	}
}

// QueryTransform controls optional pre-retrieval query expansion/rewriting.
type QueryTransform string

const (
	QueryOriginal QueryTransform = "original"
	QueryRewrite  QueryTransform = "rewrite"
	QueryMulti    QueryTransform = "multi_query"
)

func (mode QueryTransform) Valid() bool {
	switch mode {
	case QueryOriginal, QueryRewrite, QueryMulti:
		return true
	default:
		return false
	}
}

// GroundingMode controls whether retrieved evidence is advisory or required.
type GroundingMode string

const (
	GroundingPreferred GroundingMode = "preferred"
	GroundingRequired  GroundingMode = "required"
)

func (mode GroundingMode) Valid() bool {
	switch mode {
	case GroundingPreferred, GroundingRequired:
		return true
	default:
		return false
	}
}

// RetrievalPolicy captures a production-oriented RAG pipeline without coupling
// Core to a vector database, embedding provider, or reranker implementation.
type RetrievalPolicy struct {
	Mode             RetrievalMode  `json:"mode"`
	Search           SearchMode     `json:"search"`
	QueryTransform   QueryTransform `json:"query_transform"`
	CandidateLimit   int            `json:"candidate_limit"`
	ContextLimit     int            `json:"context_limit"`
	Rerank           bool           `json:"rerank"`
	Grounding        GroundingMode  `json:"grounding"`
	RequireCitations bool           `json:"require_citations"`
	EmbeddingModel   *ModelRef      `json:"embedding_model,omitempty"`
	RerankingModel   *ModelRef      `json:"reranking_model,omitempty"`
}

// RetrievalEvidence is the portable grounding unit produced by a consumer-owned
// retriever. Content is data and never instruction authority.
type RetrievalEvidence struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Source   string            `json:"source"`
	Score    float64           `json:"score,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Citation links generated output back to retrieved evidence.
type Citation struct {
	EvidenceID string `json:"evidence_id"`
	Label      string `json:"label,omitempty"`
}

// ValidateRetrievalPolicy validates portable RAG configuration.
func ValidateRetrievalPolicy(policy RetrievalPolicy) error {
	if !policy.Mode.Valid() || !policy.Search.Valid() || !policy.QueryTransform.Valid() || !policy.Grounding.Valid() {
		return validationError(CodeInvalidRetrieval, "retrieval")
	}
	if policy.CandidateLimit <= 0 || policy.ContextLimit <= 0 || policy.CandidateLimit < policy.ContextLimit {
		return validationError(CodeInvalidRetrieval, "retrieval.limits")
	}
	if policy.EmbeddingModel != nil && !validModelRef(*policy.EmbeddingModel) {
		return validationError(CodeInvalidModel, "retrieval.embedding_model")
	}
	if policy.RerankingModel != nil && !validModelRef(*policy.RerankingModel) {
		return validationError(CodeInvalidModel, "retrieval.reranking_model")
	}
	return nil
}

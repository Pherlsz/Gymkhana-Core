package assistant

import (
	"math"
	"unicode/utf8"
)

const (
	maxRetrievalCandidates = int64(10000)
	maxRetrievalContext    = int64(1024)
	maxEvidenceBytes       = 1 << 20
	maxEvidenceMetadata    = 64
)

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
	CandidateLimit   int64          `json:"candidate_limit"`
	ContextLimit     int64          `json:"context_limit"`
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
	if policy.CandidateLimit <= 0 || policy.CandidateLimit > maxRetrievalCandidates || policy.ContextLimit <= 0 || policy.ContextLimit > maxRetrievalContext || policy.CandidateLimit < policy.ContextLimit {
		return validationError(CodeInvalidRetrieval, "retrieval.limits")
	}
	if policy.EmbeddingModel != nil && !validModelRef(*policy.EmbeddingModel) {
		return validationError(CodeInvalidModel, "retrieval.embedding_model")
	}
	if policy.RerankingModel != nil && !validModelRef(*policy.RerankingModel) {
		return validationError(CodeInvalidModel, "retrieval.reranking_model")
	}
	if !policy.Rerank && policy.RerankingModel != nil {
		return validationError(CodeInvalidRetrieval, "retrieval.reranking_model")
	}
	return nil
}

// ValidateRetrievalModels verifies optional model references against a live
// catalog and the role each model is expected to perform.
func ValidateRetrievalModels(policy RetrievalPolicy, catalog []ModelDescriptor) error {
	if err := ValidateRetrievalPolicy(policy); err != nil {
		return err
	}
	if policy.EmbeddingModel != nil {
		if _, err := ResolveModelRole(*policy.EmbeddingModel, ModelRoleEmbedding, catalog); err != nil {
			return err
		}
	}
	if policy.RerankingModel != nil {
		if _, err := ResolveModelRole(*policy.RerankingModel, ModelRoleReranking, catalog); err != nil {
			return err
		}
	}
	return nil
}

// ValidateRetrievalEvidence validates evidence before it is admitted to model
// context. It does not authorize the source; ACL checks remain consumer-owned.
func ValidateRetrievalEvidence(evidence RetrievalEvidence) error {
	if !validOpaqueModelID(evidence.ID, 256) || evidence.Content == "" || !utf8.ValidString(evidence.Content) || len(evidence.Content) > maxEvidenceBytes || evidence.Source == "" || !utf8.ValidString(evidence.Source) || utf8.RuneCountInString(evidence.Source) > 8192 {
		return validationError(CodeInvalidRetrieval, "evidence")
	}
	if math.IsNaN(evidence.Score) || math.IsInf(evidence.Score, 0) {
		return validationError(CodeInvalidRetrieval, "evidence.score")
	}
	if len(evidence.Metadata) > maxEvidenceMetadata {
		return validationError(CodeInvalidRetrieval, "evidence.metadata")
	}
	for key, value := range evidence.Metadata {
		if key == "" || !utf8.ValidString(key) || utf8.RuneCountInString(key) > 256 || !utf8.ValidString(value) || utf8.RuneCountInString(value) > 4096 {
			return validationError(CodeInvalidRetrieval, "evidence.metadata")
		}
	}
	return nil
}

// ValidateEvidenceSet rejects duplicate evidence IDs and validates every item.
func ValidateEvidenceSet(values []RetrievalEvidence) error {
	if len(values) > int(maxRetrievalCandidates) {
		return validationError(CodeInvalidRetrieval, "evidence")
	}
	seen := make(map[string]struct{}, len(values))
	for _, evidence := range values {
		if err := ValidateRetrievalEvidence(evidence); err != nil {
			return err
		}
		if _, duplicate := seen[evidence.ID]; duplicate {
			return validationError(CodeInvalidRetrieval, "evidence.id")
		}
		seen[evidence.ID] = struct{}{}
	}
	return nil
}

func ValidateCitation(citation Citation) error {
	if !validOpaqueModelID(citation.EvidenceID, 256) || !utf8.ValidString(citation.Label) || utf8.RuneCountInString(citation.Label) > 512 {
		return validationError(CodeInvalidRetrieval, "citation")
	}
	return nil
}

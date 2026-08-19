package assistant

// LearningScope identifies what a learned proposal may affect.
type LearningScope string

const (
	LearningSkillHint        LearningScope = "skill_hint"
	LearningRetrievalHint    LearningScope = "retrieval_hint"
	LearningToolHint         LearningScope = "tool_hint"
	LearningModelRoutingHint LearningScope = "model_routing_hint"
)

func (scope LearningScope) Valid() bool {
	switch scope {
	case LearningSkillHint, LearningRetrievalHint, LearningToolHint, LearningModelRoutingHint:
		return true
	default:
		return false
	}
}

// LearningEvidence is a non-secret reference explaining why a reusable lesson
// was proposed. Content/prompt payload is not stored in Core contracts.
type LearningEvidence struct {
	RunID      string `json:"run_id"`
	Outcome    string `json:"outcome"`
	FailureCode string `json:"failure_code,omitempty"`
}

// SkillLearningProposal is a reviewable, versioned proposal derived from prior
// executions. It is not automatically applied to AssistantDefinition.
type SkillLearningProposal struct {
	ID          string             `json:"id"`
	AssistantID string             `json:"assistant_id"`
	Scope       LearningScope      `json:"scope"`
	Summary     string             `json:"summary"`
	Revision    int                `json:"revision"`
	Evidence    []LearningEvidence `json:"evidence"`
}

// ValidateSkillLearningProposal validates bounded portable learning metadata.
func ValidateSkillLearningProposal(proposal SkillLearningProposal) error {
	if !validPortableID(proposal.ID, 128) || !validPortableID(proposal.AssistantID, 128) || !proposal.Scope.Valid() {
		return validationError(CodeInvalidLearning, "learning")
	}
	if proposal.Summary == "" || proposal.Revision <= 0 || len(proposal.Evidence) == 0 {
		return validationError(CodeInvalidLearning, "learning")
	}
	for _, evidence := range proposal.Evidence {
		if !validOpaqueModelID(evidence.RunID, 256) || evidence.Outcome == "" {
			return validationError(CodeInvalidLearning, "learning.evidence")
		}
	}
	return nil
}

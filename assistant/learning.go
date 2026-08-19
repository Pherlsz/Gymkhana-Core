package assistant

import (
	"sort"
	"unicode/utf8"
)

// LearningScope identifies what a learned proposal may affect. Every scope in
// Spec 0.3 is advisory: applying a proposal still belongs to the application.
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

// TaskSignature identifies a reusable family of work without persisting the raw
// prompt or document content. Traits are canonical sorted portable identifiers.
type TaskSignature struct {
	Family  string   `json:"family"`
	Version int64    `json:"version"`
	Traits  []string `json:"traits,omitempty"`
}

// CanonicalTaskSignature builds a deterministic signature from reusable,
// non-secret semantic traits.
func CanonicalTaskSignature(family string, version int64, traits []string) TaskSignature {
	canonical := append([]string(nil), traits...)
	sort.Strings(canonical)
	result := canonical[:0]
	for _, trait := range canonical {
		if len(result) == 0 || result[len(result)-1] != trait {
			result = append(result, trait)
		}
	}
	return TaskSignature{Family: family, Version: version, Traits: result}
}

func ValidateTaskSignature(signature TaskSignature) error {
	if !validPortableID(signature.Family, 128) || signature.Version <= 0 || signature.Version > maxPortableJSONInteger || len(signature.Traits) > 64 {
		return validationError(CodeInvalidLearning, "task_signature")
	}
	previous := ""
	for _, trait := range signature.Traits {
		if !validPortableID(trait, 128) || (previous != "" && trait <= previous) {
			return validationError(CodeInvalidLearning, "task_signature.traits")
		}
		previous = trait
	}
	return nil
}

// LearningOutcome records the runtime/evaluation outcome supporting a proposal.
type LearningOutcome string

const (
	LearningSuccess  LearningOutcome = "success"
	LearningFailure  LearningOutcome = "failure"
	LearningImproved LearningOutcome = "improved"
	LearningRegressed LearningOutcome = "regressed"
)

func (outcome LearningOutcome) Valid() bool {
	switch outcome {
	case LearningSuccess, LearningFailure, LearningImproved, LearningRegressed:
		return true
	default:
		return false
	}
}

// LearningOrigin says where the observation came from. Model-authored document
// content is not a trusted origin; runtime/evaluation/human systems must create
// the evidence record outside model instruction authority.
type LearningOrigin string

const (
	LearningOriginRuntime    LearningOrigin = "runtime"
	LearningOriginEvaluation LearningOrigin = "evaluation"
	LearningOriginHuman      LearningOrigin = "human"
)

func (origin LearningOrigin) Valid() bool {
	switch origin {
	case LearningOriginRuntime, LearningOriginEvaluation, LearningOriginHuman:
		return true
	default:
		return false
	}
}

// LearningEvidence is a non-secret reference explaining why a reusable lesson
// was proposed. Prompt/tool/document payload is not stored in this contract.
type LearningEvidence struct {
	RunID        string          `json:"run_id"`
	Outcome      LearningOutcome `json:"outcome"`
	Origin       LearningOrigin  `json:"origin"`
	FailureClass FailureClass    `json:"failure_class,omitempty"`
}

// SkillLearningProposal is a reviewable, versioned proposal derived from prior
// executions. It is not automatically applied to AssistantDefinition.
type SkillLearningProposal struct {
	ID          string             `json:"id"`
	AssistantID string             `json:"assistant_id"`
	Task        TaskSignature      `json:"task"`
	Scope       LearningScope      `json:"scope"`
	Summary     string             `json:"summary"`
	Revision    int64              `json:"revision"`
	Evidence    []LearningEvidence `json:"evidence"`
}

// LearningMode controls whether observations are ignored, turned into proposals,
// or may be automatically promoted by an application after policy checks. Core
// never mutates an AssistantDefinition by itself.
type LearningMode string

const (
	LearningDisabled    LearningMode = "disabled"
	LearningPropose     LearningMode = "propose"
	LearningAutoPromote LearningMode = "auto_promote"
)

func (mode LearningMode) Valid() bool {
	switch mode {
	case LearningDisabled, LearningPropose, LearningAutoPromote:
		return true
	default:
		return false
	}
}

// LearningPolicy bounds proposal creation/promotion. MaxEvidence prevents an
// ever-growing replay history from becoming part of Assistant configuration.
type LearningPolicy struct {
	Mode              LearningMode  `json:"mode"`
	MinEvidence       int64         `json:"min_evidence"`
	MaxEvidence       int64         `json:"max_evidence"`
	AutoPromoteScopes []LearningScope `json:"auto_promote_scopes,omitempty"`
}

func DefaultLearningPolicy() LearningPolicy {
	return LearningPolicy{
		Mode:        LearningPropose,
		MinEvidence: 2,
		MaxEvidence: 32,
	}
}

func ValidateLearningPolicy(policy LearningPolicy) error {
	if !policy.Mode.Valid() || policy.MinEvidence < 1 || policy.MaxEvidence < policy.MinEvidence || policy.MaxEvidence > 256 {
		return validationError(CodeInvalidLearning, "learning_policy")
	}
	seen := make(map[LearningScope]struct{}, len(policy.AutoPromoteScopes))
	for _, scope := range policy.AutoPromoteScopes {
		if !scope.Valid() {
			return validationError(CodeInvalidLearning, "learning_policy.auto_promote_scopes")
		}
		if _, duplicate := seen[scope]; duplicate {
			return validationError(CodeInvalidLearning, "learning_policy.auto_promote_scopes")
		}
		seen[scope] = struct{}{}
	}
	if policy.Mode != LearningAutoPromote && len(policy.AutoPromoteScopes) > 0 {
		return validationError(CodeInvalidLearning, "learning_policy.auto_promote_scopes")
	}
	if policy.Mode == LearningDisabled && (policy.MinEvidence != 1 || policy.MaxEvidence != 1) {
		// Disabled still carries bounded canonical values rather than ambiguous zero.
		return validationError(CodeInvalidLearning, "learning_policy")
	}
	return nil
}

// ValidateSkillLearningProposal validates bounded portable learning metadata.
func ValidateSkillLearningProposal(proposal SkillLearningProposal) error {
	if !validPortableID(proposal.ID, 128) || !validPortableID(proposal.AssistantID, 128) || !proposal.Scope.Valid() {
		return validationError(CodeInvalidLearning, "learning")
	}
	if err := ValidateTaskSignature(proposal.Task); err != nil {
		return err
	}
	if proposal.Summary == "" || !utf8.ValidString(proposal.Summary) || utf8.RuneCountInString(proposal.Summary) > 4096 || proposal.Revision <= 0 || proposal.Revision > maxPortableJSONInteger || len(proposal.Evidence) == 0 || len(proposal.Evidence) > 256 {
		return validationError(CodeInvalidLearning, "learning")
	}
	seenRuns := make(map[string]struct{}, len(proposal.Evidence))
	for _, evidence := range proposal.Evidence {
		if !validOpaqueModelID(evidence.RunID, 256) || !evidence.Outcome.Valid() || !evidence.Origin.Valid() {
			return validationError(CodeInvalidLearning, "learning.evidence")
		}
		if evidence.FailureClass != "" && !evidence.FailureClass.Valid() {
			return validationError(CodeInvalidLearning, "learning.evidence.failure_class")
		}
		if evidence.Outcome == LearningFailure && evidence.FailureClass == "" {
			return validationError(CodeInvalidLearning, "learning.evidence.failure_class")
		}
		if _, duplicate := seenRuns[evidence.RunID]; duplicate {
			return validationError(CodeInvalidLearning, "learning.evidence.run_id")
		}
		seenRuns[evidence.RunID] = struct{}{}
	}
	return nil
}

// LearningEligible reports whether a valid proposal satisfies a policy's
// minimum evidence and scope requirements. It does not apply the proposal.
func LearningEligible(policy LearningPolicy, proposal SkillLearningProposal) (bool, error) {
	if err := ValidateLearningPolicy(policy); err != nil {
		return false, err
	}
	if err := ValidateSkillLearningProposal(proposal); err != nil {
		return false, err
	}
	if policy.Mode == LearningDisabled || int64(len(proposal.Evidence)) < policy.MinEvidence || int64(len(proposal.Evidence)) > policy.MaxEvidence {
		return false, nil
	}
	if policy.Mode == LearningPropose {
		return true, nil
	}
	for _, scope := range policy.AutoPromoteScopes {
		if scope == proposal.Scope {
			return true, nil
		}
	}
	return false, nil
}

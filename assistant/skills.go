package assistant

// SkillID identifies a versioned provider-neutral Assistant behavior profile.
type SkillID string

const (
	SkillTokenEconomyV1 SkillID = "token_economy/v1"
)

func (id SkillID) Valid() bool {
	switch id {
	case SkillTokenEconomyV1:
		return true
	default:
		return false
	}
}

// Skill is a discriminated union of versioned Assistant behavior profiles.
type Skill struct {
	ID           SkillID             `json:"id"`
	TokenEconomy *TokenEconomyPolicy `json:"token_economy,omitempty"`
}

// TokenEconomyPolicy defines portable behavior intended to reduce unnecessary
// input/output token usage without weakening instruction authority or dropping
// unresolved requirements.
type TokenEconomyPolicy struct {
	PreferConciseResponses        bool `json:"prefer_concise_responses"`
	AvoidRestatement              bool `json:"avoid_restatement"`
	ReusePriorContext             bool `json:"reuse_prior_context"`
	CompactToolResults            bool `json:"compact_tool_results"`
	PreserveInstructionHierarchy  bool `json:"preserve_instruction_hierarchy"`
	PreserveUnresolvedConstraints bool `json:"preserve_unresolved_constraints"`
}

func BuiltinTokenEconomySkill() Skill {
	return Skill{
		ID: SkillTokenEconomyV1,
		TokenEconomy: &TokenEconomyPolicy{
			PreferConciseResponses:        true,
			AvoidRestatement:              true,
			ReusePriorContext:             true,
			CompactToolResults:            true,
			PreserveInstructionHierarchy:  true,
			PreserveUnresolvedConstraints: true,
		},
	}
}

func BuiltinSkills() []Skill {
	return []Skill{BuiltinTokenEconomySkill()}
}

func ValidateSkill(skill Skill) error {
	if !skill.ID.Valid() {
		return validationError(CodeInvalidSkill, "skill.id")
	}
	switch skill.ID {
	case SkillTokenEconomyV1:
		if skill.TokenEconomy == nil {
			return validationError(CodeInvalidSkill, "skill.token_economy")
		}
		policy := skill.TokenEconomy
		if !policy.PreserveInstructionHierarchy || !policy.PreserveUnresolvedConstraints {
			return validationError(CodeInvalidSkill, "skill.token_economy")
		}
		if !policy.PreferConciseResponses && !policy.AvoidRestatement && !policy.ReusePriorContext && !policy.CompactToolResults {
			return validationError(CodeInvalidSkill, "skill.token_economy")
		}
		return nil
	default:
		return validationError(CodeInvalidSkill, "skill.id")
	}
}

// ValidateSkills validates an ordered skill set and rejects duplicate IDs.
func ValidateSkills(skills []Skill) error {
	if len(skills) > 32 {
		return validationError(CodeInvalidSkill, "skills")
	}
	seen := make(map[SkillID]struct{}, len(skills))
	for _, skill := range skills {
		if err := ValidateSkill(skill); err != nil {
			return err
		}
		if _, exists := seen[skill.ID]; exists {
			return validationError(CodeDuplicateSkill, "skills")
		}
		seen[skill.ID] = struct{}{}
	}
	return nil
}

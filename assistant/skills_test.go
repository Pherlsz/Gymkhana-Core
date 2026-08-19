package assistant_test

import (
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestBuiltinTokenEconomySkill(t *testing.T) {
	t.Parallel()

	skill := assistant.BuiltinTokenEconomySkill()
	if skill.ID != assistant.SkillTokenEconomyV1 {
		t.Fatalf("skill ID = %q, want %q", skill.ID, assistant.SkillTokenEconomyV1)
	}
	if skill.TokenEconomy == nil {
		t.Fatal("token economy policy is nil")
	}
	if err := assistant.ValidateSkill(skill); err != nil {
		t.Fatalf("ValidateSkill(default) = %v", err)
	}

	data, err := json.Marshal(skill)
	if err != nil {
		t.Fatalf("Marshal = %v", err)
	}
	var decoded assistant.Skill
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal = %v", err)
	}
	if err := assistant.ValidateSkill(decoded); err != nil {
		t.Fatalf("round-trip skill invalid: %v", err)
	}
}

func TestValidateSkills(t *testing.T) {
	t.Parallel()

	if err := assistant.ValidateSkills(assistant.BuiltinSkills()); err != nil {
		t.Fatalf("ValidateSkills(default) = %v", err)
	}

	invalid := assistant.BuiltinTokenEconomySkill()
	invalid.TokenEconomy.PreserveInstructionHierarchy = false
	assertAssistantCode(t, assistant.ValidateSkill(invalid), assistant.CodeInvalidSkill)

	noOptimization := assistant.BuiltinTokenEconomySkill()
	noOptimization.TokenEconomy.PreferConciseResponses = false
	noOptimization.TokenEconomy.AvoidRestatement = false
	noOptimization.TokenEconomy.ReusePriorContext = false
	noOptimization.TokenEconomy.CompactToolResults = false
	assertAssistantCode(t, assistant.ValidateSkill(noOptimization), assistant.CodeInvalidSkill)

	unknown := assistant.Skill{ID: "provider_magic/v1"}
	assertAssistantCode(t, assistant.ValidateSkill(unknown), assistant.CodeInvalidSkill)

	duplicate := []assistant.Skill{
		assistant.BuiltinTokenEconomySkill(),
		assistant.BuiltinTokenEconomySkill(),
	}
	assertAssistantCode(t, assistant.ValidateSkills(duplicate), assistant.CodeDuplicateSkill)
}

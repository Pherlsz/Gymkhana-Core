package assistant

import (
	"unicode/utf8"
)

// ModuleID identifies one independently configurable Assistant capability.
type ModuleID string

const (
	ModuleText             ModuleID = "text"
	ModuleVision           ModuleID = "vision"
	ModuleAudioInput       ModuleID = "audio_input"
	ModuleVideoInput       ModuleID = "video_input"
	ModuleFileInput        ModuleID = "file_input"
	ModuleTools            ModuleID = "tools"
	ModuleRetrieval        ModuleID = "retrieval"
	ModuleMemory           ModuleID = "memory"
	ModuleStructuredOutput ModuleID = "structured_output"
)

func (id ModuleID) Valid() bool {
	switch id {
	case ModuleText, ModuleVision, ModuleAudioInput, ModuleVideoInput, ModuleFileInput, ModuleTools, ModuleRetrieval, ModuleMemory, ModuleStructuredOutput:
		return true
	default:
		return false
	}
}

// InstructionBlock is application-owned instruction content carried by the
// portable Assistant definition. Only system/developer authority is valid.
type InstructionBlock struct {
	Role Role   `json:"role"`
	Text string `json:"text"`
}

// AssistantDefinition describes one reusable Assistant profile. Different
// definitions may use different modules, tools, RAG policy, skills, and model
// routing while sharing the same provider-neutral runtime contracts.
type AssistantDefinition struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	Instructions []InstructionBlock `json:"instructions,omitempty"`
	Modules      []ModuleID         `json:"modules"`
	Skills       []Skill            `json:"skills,omitempty"`
	Tools        []string           `json:"tools,omitempty"`
	ModelPolicy  ModelPolicy        `json:"model_policy"`
	Credentials  CredentialPolicy   `json:"credentials"`
	Retrieval    *RetrievalPolicy   `json:"retrieval,omitempty"`
	Budget       ExecutionBudget    `json:"budget,omitempty"`
}

// ExecutionBudget bounds potentially expensive multi-step behavior. Zero means
// the consumer/runtime default; Core never interprets zero as unbounded.
type ExecutionBudget struct {
	MaxTurns           int `json:"max_turns,omitempty"`
	MaxToolCalls       int `json:"max_tool_calls,omitempty"`
	MaxRetrievalRounds int `json:"max_retrieval_rounds,omitempty"`
	MaxOutputTokens    int `json:"max_output_tokens,omitempty"`
}

// AssistantCatalog is a portable collection of independently configured
// assistants for different application functions.
type AssistantCatalog struct {
	Assistants []AssistantDefinition `json:"assistants"`
}

// ValidateAssistantDefinition validates one reusable Assistant profile.
func ValidateAssistantDefinition(def AssistantDefinition) error {
	if !validPortableID(def.ID, 128) || def.Name == "" || !utf8.ValidString(def.Name) || !utf8.ValidString(def.Description) {
		return validationError(CodeInvalidAssistant, "assistant")
	}
	if len(def.Modules) == 0 {
		return validationError(CodeEmpty, "modules")
	}
	seenModules := make(map[ModuleID]struct{}, len(def.Modules))
	for _, module := range def.Modules {
		if !module.Valid() {
			return validationError(CodeInvalidModule, "modules")
		}
		if _, exists := seenModules[module]; exists {
			return validationError(CodeDuplicateModule, "modules")
		}
		seenModules[module] = struct{}{}
	}
	for _, instruction := range def.Instructions {
		if instruction.Role != RoleSystem && instruction.Role != RoleDeveloper {
			return validationError(CodeInvalidAssistant, "instructions.role")
		}
		if instruction.Text == "" || !utf8.ValidString(instruction.Text) {
			return validationError(CodeInvalidAssistant, "instructions.text")
		}
	}
	if err := ValidateSkills(def.Skills); err != nil {
		return err
	}
	seenTools := make(map[string]struct{}, len(def.Tools))
	for _, tool := range def.Tools {
		if err := ValidateToolName(tool); err != nil {
			return err
		}
		if _, exists := seenTools[tool]; exists {
			return validationError(CodeInvalidAssistant, "tools")
		}
		seenTools[tool] = struct{}{}
	}
	if len(def.Tools) > 0 && !moduleEnabled(seenModules, ModuleTools) {
		return validationError(CodeInvalidAssistant, "tools")
	}
	if def.Retrieval != nil {
		if !moduleEnabled(seenModules, ModuleRetrieval) {
			return validationError(CodeInvalidAssistant, "retrieval")
		}
		if err := ValidateRetrievalPolicy(*def.Retrieval); err != nil {
			return err
		}
	}
	if err := ValidateModelPolicy(def.ModelPolicy); err != nil {
		return err
	}
	if err := ValidateCredentialPolicy(def.Credentials); err != nil {
		return err
	}
	if def.Budget.MaxTurns < 0 || def.Budget.MaxToolCalls < 0 || def.Budget.MaxRetrievalRounds < 0 || def.Budget.MaxOutputTokens < 0 {
		return validationError(CodeInvalidAssistant, "budget")
	}
	return nil
}

// ValidateAssistantCatalog validates multiple independently addressable profiles.
func ValidateAssistantCatalog(catalog AssistantCatalog) error {
	if len(catalog.Assistants) == 0 {
		return validationError(CodeEmpty, "assistants")
	}
	seen := make(map[string]struct{}, len(catalog.Assistants))
	for _, def := range catalog.Assistants {
		if err := ValidateAssistantDefinition(def); err != nil {
			return err
		}
		if _, exists := seen[def.ID]; exists {
			return validationError(CodeDuplicateAssistant, "assistants")
		}
		seen[def.ID] = struct{}{}
	}
	return nil
}

func moduleEnabled(modules map[ModuleID]struct{}, module ModuleID) bool {
	_, ok := modules[module]
	return ok
}

func validPortableID(value string, max int) bool {
	if value == "" || len(value) > max || !utf8.ValidString(value) {
		return false
	}
	for i := 0; i < len(value); i++ {
		b := value[i]
		if i == 0 {
			if !asciiLetter(b) && b != '_' {
				return false
			}
			continue
		}
		if !asciiLetter(b) && (b < '0' || b > '9') && b != '_' && b != '-' && b != '.' && b != '/' {
			return false
		}
	}
	return true
}

package assistant

import "unicode/utf8"

const maxAssistantInstructionBytes = 1 << 20

// ModuleID identifies one independently configurable Assistant capability.
type ModuleID string

const (
	ModuleText             ModuleID = "text"
	ModuleVision           ModuleID = "vision"
	ModuleAudioInput       ModuleID = "audio_input"
	ModuleVideoInput       ModuleID = "video_input"
	ModuleFileInput        ModuleID = "file_input"
	ModuleImageOutput      ModuleID = "image_output"
	ModuleAudioOutput      ModuleID = "audio_output"
	ModuleVideoOutput      ModuleID = "video_output"
	ModuleFileOutput       ModuleID = "file_output"
	ModuleTools            ModuleID = "tools"
	ModuleRetrieval        ModuleID = "retrieval"
	ModuleMemory           ModuleID = "memory"
	ModuleStructuredOutput ModuleID = "structured_output"
)

func (id ModuleID) Valid() bool {
	switch id {
	case ModuleText, ModuleVision, ModuleAudioInput, ModuleVideoInput, ModuleFileInput, ModuleImageOutput, ModuleAudioOutput, ModuleVideoOutput, ModuleFileOutput, ModuleTools, ModuleRetrieval, ModuleMemory, ModuleStructuredOutput:
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

// AssistantDefinition describes one reusable Assistant profile. Modules are
// enabled by default; OptionalModules are available for explicit per-run enablement.
type AssistantDefinition struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Instructions    []InstructionBlock     `json:"instructions,omitempty"`
	Modules         []ModuleID             `json:"modules"`
	OptionalModules []ModuleID             `json:"optional_modules,omitempty"`
	Skills          []Skill                `json:"skills,omitempty"`
	Tools           []string               `json:"tools,omitempty"`
	ToolPolicies    []ToolPolicy           `json:"tool_policies,omitempty"`
	ModelPolicy     ModelPolicy            `json:"model_policy"`
	Credentials     CredentialPolicy       `json:"credentials"`
	RoutingFallback *RoutingFallbackPolicy `json:"routing_fallback,omitempty"`
	Retrieval       *RetrievalPolicy       `json:"retrieval,omitempty"`
	Memory          *MemoryPolicy          `json:"memory,omitempty"`
	Learning        *LearningPolicy        `json:"learning,omitempty"`
	Budget          ExecutionBudget        `json:"budget,omitempty"`
}

// ExecutionBudget bounds potentially expensive multi-step behavior. Zero means
// the consumer/runtime default; Core never interprets zero as unbounded.
type ExecutionBudget struct {
	MaxTurns           int64 `json:"max_turns,omitempty"`
	MaxToolCalls       int64 `json:"max_tool_calls,omitempty"`
	MaxRetrievalRounds int64 `json:"max_retrieval_rounds,omitempty"`
	MaxOutputTokens    int64 `json:"max_output_tokens,omitempty"`
}

// AssistantCatalog is a portable collection of independently configured
// assistants for different application functions.
type AssistantCatalog struct {
	Assistants []AssistantDefinition `json:"assistants"`
}

// ValidateAssistantDefinition validates one reusable Assistant profile.
func ValidateAssistantDefinition(def AssistantDefinition) error {
	if !validPortableID(def.ID, 128) || def.Name == "" || !utf8.ValidString(def.Name) || utf8.RuneCountInString(def.Name) > 256 || !utf8.ValidString(def.Description) || utf8.RuneCountInString(def.Description) > 4096 {
		return validationError(CodeInvalidAssistant, "assistant")
	}
	if len(def.Modules) == 0 || len(def.Modules) > 32 || len(def.OptionalModules) > 32 {
		return validationError(CodeEmpty, "modules")
	}
	allModules := make(map[ModuleID]struct{}, len(def.Modules)+len(def.OptionalModules))
	for _, module := range append(append([]ModuleID{}, def.Modules...), def.OptionalModules...) {
		if !module.Valid() {
			return validationError(CodeInvalidModule, "modules")
		}
		if _, exists := allModules[module]; exists {
			return validationError(CodeDuplicateModule, "modules")
		}
		allModules[module] = struct{}{}
	}
	if len(def.Instructions) > 64 {
		return validationError(CodeInvalidAssistant, "instructions")
	}
	for _, instruction := range def.Instructions {
		if instruction.Role != RoleSystem && instruction.Role != RoleDeveloper {
			return validationError(CodeInvalidAssistant, "instructions.role")
		}
		if instruction.Text == "" || !utf8.ValidString(instruction.Text) || len(instruction.Text) > maxAssistantInstructionBytes {
			return validationError(CodeInvalidAssistant, "instructions.text")
		}
	}
	if len(def.Skills) > 32 {
		return validationError(CodeInvalidAssistant, "skills")
	}
	if err := ValidateSkills(def.Skills); err != nil {
		return err
	}
	if len(def.Tools) > 256 || len(def.ToolPolicies) > 256 {
		return validationError(CodeInvalidAssistant, "tools")
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
	if len(def.Tools) > 0 && !moduleEnabled(allModules, ModuleTools) {
		return validationError(CodeInvalidAssistant, "tools")
	}
	if err := ValidateToolPolicies(def.Tools, def.ToolPolicies); err != nil {
		return err
	}
	if def.Retrieval != nil {
		if !moduleEnabled(allModules, ModuleRetrieval) {
			return validationError(CodeInvalidAssistant, "retrieval")
		}
		if err := ValidateRetrievalPolicy(*def.Retrieval); err != nil {
			return err
		}
	} else if moduleEnabled(allModules, ModuleRetrieval) {
		return validationError(CodeInvalidAssistant, "retrieval")
	}
	if def.Memory != nil {
		if !moduleEnabled(allModules, ModuleMemory) {
			return validationError(CodeInvalidAssistant, "memory")
		}
		if err := ValidateMemoryPolicy(*def.Memory); err != nil {
			return err
		}
	} else if moduleEnabled(allModules, ModuleMemory) {
		return validationError(CodeInvalidAssistant, "memory")
	}
	if def.Learning != nil {
		if err := ValidateLearningPolicy(*def.Learning); err != nil {
			return err
		}
	}
	if err := ValidateModelPolicy(def.ModelPolicy); err != nil {
		return err
	}
	if err := ValidateCredentialPolicy(def.Credentials); err != nil {
		return err
	}
	if def.RoutingFallback != nil {
		if err := ValidateRoutingFallbackPolicy(*def.RoutingFallback); err != nil {
			return err
		}
	}
	budgetValues := [...]int64{def.Budget.MaxTurns, def.Budget.MaxToolCalls, def.Budget.MaxRetrievalRounds, def.Budget.MaxOutputTokens}
	for _, value := range budgetValues {
		if value < 0 || value > maxPortableJSONInteger {
			return validationError(CodeInvalidAssistant, "budget")
		}
	}
	return nil
}

// ValidateAssistantCatalog validates multiple independently addressable profiles.
func ValidateAssistantCatalog(catalog AssistantCatalog) error {
	if len(catalog.Assistants) == 0 || len(catalog.Assistants) > 1024 {
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

package assistant

// MemoryScope identifies an application-owned persistence scope. Core defines
// semantics only; it does not store memory or infer tenant/user identity.
type MemoryScope string

const (
	MemorySession   MemoryScope = "session"
	MemoryAssistant MemoryScope = "assistant"
	MemoryUser      MemoryScope = "user"
	MemoryTenant    MemoryScope = "tenant"
)

func (scope MemoryScope) Valid() bool {
	switch scope {
	case MemorySession, MemoryAssistant, MemoryUser, MemoryTenant:
		return true
	default:
		return false
	}
}

// MemoryWriteMode controls how persistent memory may be created. Spec 0.3 does
// not permit an unrestricted automatic-write mode.
type MemoryWriteMode string

const (
	MemoryWriteDisabled  MemoryWriteMode = "disabled"
	MemoryWriteExplicit  MemoryWriteMode = "explicit"
	MemoryWriteConfirmed MemoryWriteMode = "confirmed"
)

func (mode MemoryWriteMode) Valid() bool {
	switch mode {
	case MemoryWriteDisabled, MemoryWriteExplicit, MemoryWriteConfirmed:
		return true
	default:
		return false
	}
}

// SensitiveMemoryPolicy controls whether data classified as sensitive by the
// consuming application may be persisted. Core never performs the classification.
type SensitiveMemoryPolicy string

const (
	MemorySensitiveExclude SensitiveMemoryPolicy = "exclude"
	MemorySensitiveConfirm SensitiveMemoryPolicy = "confirm"
)

func (policy SensitiveMemoryPolicy) Valid() bool {
	switch policy {
	case MemorySensitiveExclude, MemorySensitiveConfirm:
		return true
	default:
		return false
	}
}

// MemoryPolicy configures portable read/write behavior. Zero limits mean use
// application defaults, never unlimited storage.
type MemoryPolicy struct {
	ReadScopes   []MemoryScope         `json:"read_scopes,omitempty"`
	WriteMode    MemoryWriteMode       `json:"write_mode"`
	WriteScope   MemoryScope           `json:"write_scope,omitempty"`
	Sensitive    SensitiveMemoryPolicy `json:"sensitive"`
	MaxItems     int64                 `json:"max_items,omitempty"`
	MaxItemBytes int64                 `json:"max_item_bytes,omitempty"`
}

// MemoryItemRef is persistable provenance metadata for a memory item without
// embedding its content into traces or portable Assistant configuration.
type MemoryItemRef struct {
	ID         string      `json:"id"`
	Scope      MemoryScope `json:"scope"`
	Provenance string      `json:"provenance,omitempty"`
}

func ValidateMemoryPolicy(policy MemoryPolicy) error {
	if !policy.WriteMode.Valid() || !policy.Sensitive.Valid() {
		return validationError(CodeInvalidMemory, "memory")
	}
	seen := make(map[MemoryScope]struct{}, len(policy.ReadScopes))
	for _, scope := range policy.ReadScopes {
		if !scope.Valid() {
			return validationError(CodeInvalidMemory, "memory.read_scopes")
		}
		if _, duplicate := seen[scope]; duplicate {
			return validationError(CodeInvalidMemory, "memory.read_scopes")
		}
		seen[scope] = struct{}{}
	}
	if policy.WriteMode == MemoryWriteDisabled {
		if policy.WriteScope != "" {
			return validationError(CodeInvalidMemory, "memory.write_scope")
		}
	} else if !policy.WriteScope.Valid() {
		return validationError(CodeInvalidMemory, "memory.write_scope")
	}
	if len(policy.ReadScopes) == 0 && policy.WriteMode == MemoryWriteDisabled {
		return validationError(CodeInvalidMemory, "memory")
	}
	if policy.MaxItems < 0 || policy.MaxItems > maxPortableJSONInteger || policy.MaxItemBytes < 0 || policy.MaxItemBytes > maxPortableJSONInteger {
		return validationError(CodeInvalidMemory, "memory.limits")
	}
	return nil
}

func ValidateMemoryItemRef(ref MemoryItemRef) error {
	if !validOpaqueModelID(ref.ID, 256) || !ref.Scope.Valid() {
		return validationError(CodeInvalidMemory, "memory_item")
	}
	if ref.Provenance != "" && !validOpaqueModelID(ref.Provenance, 256) {
		return validationError(CodeInvalidMemory, "memory_item.provenance")
	}
	return nil
}

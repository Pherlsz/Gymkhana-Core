package ocr

import "encoding/json"

const (
	MaxSources                  = 32
	MaxObservations             = 4096
	MaxCandidates               = 2048
	MaxEvidencePerObservation   = 8
	MaxObservationsPerCandidate = 64
	MaxSemanticCandidates       = 16
	MaxJurisdictionCandidates   = 16
	MaxWarnings                 = 256
	MaxIDRunes                  = 256
	MaxMediaTypeRunes           = 256
	MaxPointerRunes             = 2048
	MaxObservationTextBytes     = 16 << 10
	MaxResultDataBytes          = 16 << 20
	MaxPage                     = 1_000_000
	NormalizedCoordinateScale   = 1_000_000
	ConfidenceScale             = 10_000
)

// ExtractionMode selects the portable extraction path.
type ExtractionMode string

const (
	ModeSchemaGuided ExtractionMode = "schema_guided"
	ModeDiscovery    ExtractionMode = "discovery"
)

func (m ExtractionMode) Valid() bool {
	return m == ModeSchemaGuided || m == ModeDiscovery
}

// SourceModality describes the logical source without implying a provider API.
type SourceModality string

const (
	SourceText     SourceModality = "text"
	SourceImage    SourceModality = "image"
	SourceDocument SourceModality = "document"
	SourceAudio    SourceModality = "audio"
	SourceVideo    SourceModality = "video"
)

func (m SourceModality) Valid() bool {
	switch m {
	case SourceText, SourceImage, SourceDocument, SourceAudio, SourceVideo:
		return true
	default:
		return false
	}
}

// SourceRef identifies source material already authorized and supplied by the
// host. Core intentionally carries no URI, bytes, storage key, or ACL state.
type SourceRef struct {
	ID        string         `json:"id"`
	Modality  SourceModality `json:"modality"`
	MediaType string         `json:"media_type,omitempty"`
}

// ExtractionRequest describes one provider-neutral extraction operation.
type ExtractionRequest struct {
	Mode          ExtractionMode  `json:"mode"`
	Sources       []SourceRef     `json:"sources"`
	TargetSchema  json.RawMessage `json:"target_schema,omitempty"`
	MaxCandidates int             `json:"max_candidates,omitempty"`
}

// Confidence is a portable fixed-scale confidence assessment. 0 means no
// confidence and 10000 means maximum confidence; absence means unknown.
type Confidence int

func (c Confidence) Valid() bool { return c >= 0 && c <= ConfidenceScale }

// NormalizedRect is a source-relative rectangle. Coordinates are millionths of
// source width/height in [0, 1_000_000], avoiding provider-specific pixels and
// binary floating-point serialization differences.
type NormalizedRect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// TextSpan is a zero-based, half-open UTF-8 byte range in a stable text surface
// exposed by the host/adapter for the referenced source.
type TextSpan struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// TimeRange is a zero-based, half-open millisecond range for audio/video evidence.
type TimeRange struct {
	StartMS int64 `json:"start_ms"`
	EndMS   int64 `json:"end_ms"`
}

// EvidenceRef points at the location supporting an observation. SourceID is
// mandatory; page/region/span/time locators are optional and may be combined.
type EvidenceRef struct {
	SourceID  string          `json:"source_id"`
	Page      int             `json:"page,omitempty"`
	Region    *NormalizedRect `json:"region,omitempty"`
	TextSpan  *TextSpan       `json:"text_span,omitempty"`
	TimeRange *TimeRange      `json:"time_range,omitempty"`
}

// Observation is provider-produced evidence before semantic normalization. It
// can preserve exact observed text, an arbitrary portable JSON value, or both.
type Observation struct {
	ID         string          `json:"id"`
	Evidence   []EvidenceRef   `json:"evidence"`
	RawText    string          `json:"raw_text,omitempty"`
	Value      json.RawMessage `json:"value,omitempty"`
	Confidence *Confidence     `json:"confidence,omitempty"`
}

// CandidateBasis states how a field candidate relates to its evidence.
type CandidateBasis string

const (
	BasisObserved CandidateBasis = "observed"
	BasisDerived  CandidateBasis = "derived"
	BasisInferred CandidateBasis = "inferred"
)

func (b CandidateBasis) Valid() bool {
	return b == BasisObserved || b == BasisDerived || b == BasisInferred
}

// ValueState distinguishes an absent value from the JSON value null.
type ValueState string

const (
	ValuePresent ValueState = "present"
	ValueMissing ValueState = "missing"
)

func (s ValueState) Valid() bool { return s == ValuePresent || s == ValueMissing }

// ValidationStatus records whether explicit semantic/schema validation occurred.
type ValidationStatus string

const (
	ValidationNotValidated ValidationStatus = "not_validated"
	ValidationValid        ValidationStatus = "valid"
	ValidationInvalid      ValidationStatus = "invalid"
)

func (s ValidationStatus) Valid() bool {
	return s == ValidationNotValidated || s == ValidationValid || s == ValidationInvalid
}

// ReviewStatus records human/application review state without implementing UI.
type ReviewStatus string

const (
	ReviewUnreviewed  ReviewStatus = "unreviewed"
	ReviewNeedsReview ReviewStatus = "needs_review"
	ReviewAccepted    ReviewStatus = "accepted"
	ReviewRejected    ReviewStatus = "rejected"
)

func (s ReviewStatus) Valid() bool {
	return s == ReviewUnreviewed || s == ReviewNeedsReview || s == ReviewAccepted || s == ReviewRejected
}

// FieldCandidate is a semantic candidate backed by one or more observations.
// Path is a non-root RFC 6901 JSON Pointer. NormalizedValue is allowed only together
// with an explicit Core/consumer canonicalizer identifier.
type FieldCandidate struct {
	Path                   string           `json:"path"`
	State                  ValueState       `json:"state"`
	Value                  json.RawMessage  `json:"value,omitempty"`
	ObservationIDs         []string         `json:"observation_ids"`
	Basis                  CandidateBasis   `json:"basis"`
	Confidence             *Confidence      `json:"confidence,omitempty"`
	SemanticTypeCandidates []string         `json:"semantic_type_candidates,omitempty"`
	JurisdictionCandidates []string         `json:"jurisdiction_candidates,omitempty"`
	NormalizedValue        json.RawMessage  `json:"normalized_value,omitempty"`
	Normalizer             string           `json:"normalizer,omitempty"`
	Ambiguous              bool             `json:"ambiguous,omitempty"`
	Validation             ValidationStatus `json:"validation"`
	Review                 ReviewStatus     `json:"review"`
}

// WarningCode is a stable extraction warning that does not embed source data.
type WarningCode string

const (
	WarningPartialExtraction   WarningCode = "partial_extraction"
	WarningAmbiguous           WarningCode = "ambiguous"
	WarningConflictingEvidence WarningCode = "conflicting_evidence"
	WarningUnverifiedSemantics WarningCode = "unverified_semantics"
)

func (c WarningCode) Valid() bool {
	switch c {
	case WarningPartialExtraction, WarningAmbiguous, WarningConflictingEvidence, WarningUnverifiedSemantics:
		return true
	default:
		return false
	}
}

type Warning struct {
	Code WarningCode `json:"code"`
	Path string      `json:"path,omitempty"`
}

// ExtractionResult contains evidence-backed candidate data. StructuredData is
// used only by schema-guided extraction; discovery remains candidate-oriented
// until a target schema is deliberately selected.
type ExtractionResult struct {
	Mode           ExtractionMode   `json:"mode"`
	Observations   []Observation    `json:"observations"`
	Candidates     []FieldCandidate `json:"candidates"`
	StructuredData json.RawMessage  `json:"structured_data,omitempty"`
	Validation     ValidationStatus `json:"validation"`
	Review         ReviewStatus     `json:"review"`
	Warnings       []Warning        `json:"warnings,omitempty"`
}

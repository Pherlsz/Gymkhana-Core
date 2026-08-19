package ocr

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Pherlsz/Gymkhana-Core/portablejson"
)

func ValidateExtractionRequest(request ExtractionRequest) error {
	if !request.Mode.Valid() {
		return validationError(CodeInvalidMode, "mode")
	}
	if len(request.Sources) == 0 || len(request.Sources) > MaxSources {
		return validationError(CodeInvalidSource, "sources")
	}
	seen := make(map[string]struct{}, len(request.Sources))
	for _, source := range request.Sources {
		if err := validateSource(source); err != nil {
			return err
		}
		if _, duplicate := seen[source.ID]; duplicate {
			return validationError(CodeDuplicateSource, "sources")
		}
		seen[source.ID] = struct{}{}
	}
	if request.MaxCandidates < 0 || request.MaxCandidates > MaxCandidates {
		return validationError(CodeInvalidLimit, "max_candidates")
	}
	switch request.Mode {
	case ModeSchemaGuided:
		if len(request.TargetSchema) == 0 || portablejson.ValidateObjectSchema(request.TargetSchema) != nil {
			return validationError(CodeInvalidSchema, "target_schema")
		}
	case ModeDiscovery:
		if len(request.TargetSchema) != 0 {
			return validationError(CodeInvalidSchema, "target_schema")
		}
	}
	return nil
}

func validateSource(source SourceRef) error {
	if !validIdentifier(source.ID) || !source.Modality.Valid() {
		return validationError(CodeInvalidSource, "source")
	}
	if source.MediaType != "" && (!utf8.ValidString(source.MediaType) || utf8.RuneCountInString(source.MediaType) > MaxMediaTypeRunes) {
		return validationError(CodeInvalidSource, "source.media_type")
	}
	return nil
}

func ValidateExtractionResult(result ExtractionResult) error {
	if !result.Mode.Valid() {
		return validationError(CodeInvalidMode, "mode")
	}
	if len(result.Observations) > MaxObservations {
		return validationError(CodeInvalidObservation, "observations")
	}
	if len(result.Candidates) > MaxCandidates {
		return validationError(CodeInvalidCandidate, "candidates")
	}
	if !result.Validation.Valid() {
		return validationError(CodeInvalidValidation, "validation")
	}
	if !result.Review.Valid() {
		return validationError(CodeInvalidReview, "review")
	}
	if len(result.Warnings) > MaxWarnings {
		return validationError(CodeInvalidWarning, "warnings")
	}

	dataBytes := int64(len(result.StructuredData))
	if dataBytes > MaxResultDataBytes {
		return validationError(CodeInvalidLimit, "result_data")
	}
	observations := make(map[string]struct{}, len(result.Observations))
	for _, observation := range result.Observations {
		if err := validateObservation(observation); err != nil {
			return err
		}
		if _, duplicate := observations[observation.ID]; duplicate {
			return validationError(CodeDuplicateObservation, "observations")
		}
		observations[observation.ID] = struct{}{}
		dataBytes += int64(len(observation.RawText) + len(observation.Value))
		if dataBytes > MaxResultDataBytes {
			return validationError(CodeInvalidLimit, "result_data")
		}
	}
	for _, candidate := range result.Candidates {
		if err := validateCandidate(candidate); err != nil {
			return err
		}
		for _, id := range candidate.ObservationIDs {
			if _, ok := observations[id]; !ok {
				return validationError(CodeUnknownObservation, "candidate.observation_ids")
			}
		}
		dataBytes += int64(len(candidate.Value) + len(candidate.NormalizedValue))
		if dataBytes > MaxResultDataBytes {
			return validationError(CodeInvalidLimit, "result_data")
		}
	}
	if err := validateCandidateAmbiguity(result.Candidates); err != nil {
		return err
	}
	for _, warning := range result.Warnings {
		if !warning.Code.Valid() || (warning.Path != "" && !validJSONPointer(warning.Path)) {
			return validationError(CodeInvalidWarning, "warnings")
		}
	}

	switch result.Mode {
	case ModeSchemaGuided:
		if len(result.StructuredData) == 0 || portablejson.ValidateObject(result.StructuredData) != nil {
			return validationError(CodeInvalidStructuredData, "structured_data")
		}
	case ModeDiscovery:
		if len(result.StructuredData) != 0 || result.Validation != ValidationNotValidated {
			return validationError(CodeInvalidResult, "discovery")
		}
	}
	return nil
}

func validateObservation(observation Observation) error {
	if !validIdentifier(observation.ID) || len(observation.Evidence) == 0 || len(observation.Evidence) > MaxEvidencePerObservation {
		return validationError(CodeInvalidObservation, "observation")
	}
	if observation.RawText == "" && len(observation.Value) == 0 {
		return validationError(CodeInvalidObservation, "observation.value")
	}
	if observation.RawText != "" && (!utf8.ValidString(observation.RawText) || len(observation.RawText) > MaxObservationTextBytes) {
		return validationError(CodeInvalidObservation, "observation.raw_text")
	}
	if len(observation.Value) != 0 && portablejson.ValidateValue(observation.Value) != nil {
		return validationError(CodeInvalidObservation, "observation.value")
	}
	if observation.Confidence != nil && !observation.Confidence.Valid() {
		return validationError(CodeInvalidConfidence, "observation.confidence")
	}
	for _, evidence := range observation.Evidence {
		if err := validateEvidence(evidence); err != nil {
			return err
		}
	}
	return nil
}

func validateEvidence(evidence EvidenceRef) error {
	if !validIdentifier(evidence.SourceID) || evidence.Page < 0 || evidence.Page > MaxPage {
		return validationError(CodeInvalidEvidence, "evidence")
	}
	if evidence.Region != nil {
		r := evidence.Region
		if r.X < 0 || r.Y < 0 || r.Width <= 0 || r.Height <= 0 || r.X > NormalizedCoordinateScale || r.Y > NormalizedCoordinateScale || r.Width > NormalizedCoordinateScale || r.Height > NormalizedCoordinateScale || r.X+r.Width > NormalizedCoordinateScale || r.Y+r.Height > NormalizedCoordinateScale {
			return validationError(CodeInvalidEvidence, "evidence.region")
		}
	}
	if evidence.TextSpan != nil {
		if evidence.TextSpan.Start < 0 || evidence.TextSpan.End <= evidence.TextSpan.Start || evidence.TextSpan.End > portablejson.MaxSafeInteger {
			return validationError(CodeInvalidEvidence, "evidence.text_span")
		}
	}
	if evidence.TimeRange != nil {
		if evidence.TimeRange.StartMS < 0 || evidence.TimeRange.EndMS <= evidence.TimeRange.StartMS || evidence.TimeRange.EndMS > portablejson.MaxSafeInteger {
			return validationError(CodeInvalidEvidence, "evidence.time_range")
		}
	}
	return nil
}

func validateCandidate(candidate FieldCandidate) error {
	if !validJSONPointer(candidate.Path) || !candidate.State.Valid() || !candidate.Basis.Valid() || !candidate.Validation.Valid() || !candidate.Review.Valid() {
		return validationError(CodeInvalidCandidate, "candidate")
	}
	if len(candidate.ObservationIDs) == 0 || len(candidate.ObservationIDs) > MaxObservationsPerCandidate {
		return validationError(CodeInvalidCandidate, "candidate.observation_ids")
	}
	seen := make(map[string]struct{}, len(candidate.ObservationIDs))
	for _, id := range candidate.ObservationIDs {
		if !validIdentifier(id) {
			return validationError(CodeInvalidCandidate, "candidate.observation_ids")
		}
		if _, duplicate := seen[id]; duplicate {
			return validationError(CodeInvalidCandidate, "candidate.observation_ids")
		}
		seen[id] = struct{}{}
	}
	if candidate.Confidence != nil && !candidate.Confidence.Valid() {
		return validationError(CodeInvalidConfidence, "candidate.confidence")
	}
	if err := validateCandidateStrings(candidate.SemanticTypeCandidates, MaxSemanticCandidates, "candidate.semantic_type_candidates"); err != nil {
		return err
	}
	if err := validateCandidateStrings(candidate.JurisdictionCandidates, MaxJurisdictionCandidates, "candidate.jurisdiction_candidates"); err != nil {
		return err
	}

	switch candidate.State {
	case ValuePresent:
		if len(candidate.Value) == 0 || portablejson.ValidateValue(candidate.Value) != nil {
			return validationError(CodeInvalidCandidate, "candidate.value")
		}
	case ValueMissing:
		if len(candidate.Value) != 0 || len(candidate.NormalizedValue) != 0 || candidate.Normalizer != "" {
			return validationError(CodeInvalidCandidate, "candidate.value")
		}
	}
	if (len(candidate.NormalizedValue) == 0) != (candidate.Normalizer == "") {
		return validationError(CodeInvalidCandidate, "candidate.normalized_value")
	}
	if len(candidate.NormalizedValue) != 0 {
		if portablejson.ValidateValue(candidate.NormalizedValue) != nil || !validNamespace(candidate.Normalizer) {
			return validationError(CodeInvalidCandidate, "candidate.normalized_value")
		}
	}
	return nil
}

func validateCandidateStrings(values []string, max int, field string) error {
	if len(values) > max {
		return validationError(CodeInvalidCandidate, field)
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validNamespace(value) {
			return validationError(CodeInvalidCandidate, field)
		}
		if _, duplicate := seen[value]; duplicate {
			return validationError(CodeInvalidCandidate, field)
		}
		seen[value] = struct{}{}
	}
	return nil
}

// ValidateExtractionExchange validates cross-object invariants that require both
// the original request and provider-produced result.
func ValidateExtractionExchange(request ExtractionRequest, result ExtractionResult) error {
	if err := ValidateExtractionRequest(request); err != nil {
		return err
	}
	if err := ValidateExtractionResult(result); err != nil {
		return err
	}
	if request.Mode != result.Mode {
		return validationError(CodeInvalidResult, "mode")
	}
	if request.MaxCandidates > 0 && len(result.Candidates) > request.MaxCandidates {
		return validationError(CodeInvalidLimit, "candidates")
	}

	sources := make(map[string]SourceModality, len(request.Sources))
	for _, source := range request.Sources {
		sources[source.ID] = source.Modality
	}
	for _, observation := range result.Observations {
		for _, evidence := range observation.Evidence {
			modality, ok := sources[evidence.SourceID]
			if !ok {
				return validationError(CodeUnknownSource, "evidence.source_id")
			}
			if evidence.Page > 0 && modality != SourceDocument {
				return validationError(CodeInvalidEvidence, "evidence.page")
			}
			if evidence.TimeRange != nil && modality != SourceAudio && modality != SourceVideo {
				return validationError(CodeInvalidEvidence, "evidence.time_range")
			}
		}
	}

	if request.Mode == ModeSchemaGuided {
		if err := validateCandidatesAgainstStructuredData(result.Candidates, result.StructuredData); err != nil {
			return err
		}
		if err := portablejson.ValidateInstance(request.TargetSchema, result.StructuredData); err == nil {
			if result.Validation != ValidationValid {
				return validationError(CodeInvalidValidation, "validation")
			}
		} else if result.Validation != ValidationInvalid {
			return validationError(CodeInvalidValidation, "validation")
		}
	}
	return nil
}

func validateCandidatesAgainstStructuredData(candidates []FieldCandidate, raw json.RawMessage) error {
	root, err := decodePortableValue(raw)
	if err != nil {
		return validationError(CodeInvalidStructuredData, "structured_data")
	}
	for _, candidate := range candidates {
		value, exists := resolveJSONPointer(root, candidate.Path)
		switch candidate.State {
		case ValuePresent:
			if !exists {
				return validationError(CodeInvalidCandidate, "candidate.path")
			}
			candidateValue, err := decodePortableValue(candidate.Value)
			if err != nil || !portableValueEqual(candidateValue, value) {
				return validationError(CodeInvalidCandidate, "candidate.value")
			}
		case ValueMissing:
			if exists {
				return validationError(CodeInvalidCandidate, "candidate.path")
			}
		}
	}
	return nil
}

func decodePortableValue(raw json.RawMessage) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func resolveJSONPointer(root any, pointer string) (any, bool) {
	current := root
	for _, token := range strings.Split(pointer[1:], "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		switch value := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = value[token]
			if !ok {
				return nil, false
			}
		case []any:
			if token == "" || (len(token) > 1 && token[0] == '0') {
				return nil, false
			}
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(value) {
				return nil, false
			}
			current = value[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func portableValueEqual(left, right any) bool {
	switch l := left.(type) {
	case nil:
		return right == nil
	case bool:
		r, ok := right.(bool)
		return ok && l == r
	case string:
		r, ok := right.(string)
		return ok && l == r
	case json.Number:
		r, ok := right.(json.Number)
		return ok && portablejson.EqualNumbers(l, r)
	case []any:
		r, ok := right.([]any)
		if !ok || len(l) != len(r) {
			return false
		}
		for i := range l {
			if !portableValueEqual(l[i], r[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		r, ok := right.(map[string]any)
		if !ok || len(l) != len(r) {
			return false
		}
		for key, value := range l {
			other, exists := r[key]
			if !exists || !portableValueEqual(value, other) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func validIdentifier(value string) bool {
	if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > MaxIDRunes {
		return false
	}
	for _, r := range value {
		if r <= 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

func validNamespace(value string) bool {
	if !validIdentifier(value) {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' || r == '/' || r == ':' {
			continue
		}
		return false
	}
	return true
}

func validJSONPointer(value string) bool {
	if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > MaxPointerRunes || !strings.HasPrefix(value, "/") {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] != '~' {
			continue
		}
		if i+1 >= len(value) || (value[i+1] != '0' && value[i+1] != '1') {
			return false
		}
		i++
	}
	return true
}

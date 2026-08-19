package ocr

// SkillID identifies an OCR/extraction semantic skill independently of provider prompts.
type SkillID string

const SkillDataIdentificationV1 SkillID = "data_identification/v1"

// DataIdentificationPolicy captures the normative safety/correctness behavior
// of discovery extraction. It is semantic policy, not provider prompt text.
type DataIdentificationPolicy struct {
	PreserveRawObservations                bool `json:"preserve_raw_observations"`
	RequireEvidence                        bool `json:"require_evidence"`
	PreserveAmbiguity                      bool `json:"preserve_ambiguity"`
	DoNotInventMissingValues               bool `json:"do_not_invent_missing_values"`
	DoNotAssumeJurisdiction                bool `json:"do_not_assume_jurisdiction"`
	NormalizeOnlyWithExplicitCanonicalizer bool `json:"normalize_only_with_explicit_canonicalizer"`
}

type DataIdentificationSkill struct {
	ID     SkillID                  `json:"id"`
	Policy DataIdentificationPolicy `json:"policy"`
}

// BuiltinDataIdentificationSkill returns the portable discovery semantics.
func BuiltinDataIdentificationSkill() DataIdentificationSkill {
	return DataIdentificationSkill{
		ID: SkillDataIdentificationV1,
		Policy: DataIdentificationPolicy{
			PreserveRawObservations:                true,
			RequireEvidence:                        true,
			PreserveAmbiguity:                      true,
			DoNotInventMissingValues:               true,
			DoNotAssumeJurisdiction:                true,
			NormalizeOnlyWithExplicitCanonicalizer: true,
		},
	}
}

// RequiredSkills returns the built-in skills required by the request. Schema-
// guided extraction deliberately requires no discovery skill.
func RequiredSkills(request ExtractionRequest) ([]SkillID, error) {
	if err := ValidateExtractionRequest(request); err != nil {
		return nil, err
	}
	if request.Mode == ModeDiscovery {
		return []SkillID{SkillDataIdentificationV1}, nil
	}
	return []SkillID{}, nil
}

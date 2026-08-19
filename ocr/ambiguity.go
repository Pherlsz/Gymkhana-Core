package ocr

func validateCandidateAmbiguity(candidates []FieldCandidate) error {
	groups := make(map[string][]FieldCandidate)
	for _, candidate := range candidates {
		if len(candidate.SemanticTypeCandidates) > 1 || len(candidate.JurisdictionCandidates) > 1 {
			switch candidate.Review {
			case ReviewAccepted:
				return validationError(CodeInvalidReview, "candidate.review")
			case ReviewRejected:
			default:
				if !candidate.Ambiguous {
					return validationError(CodeInvalidCandidate, "candidate.ambiguous")
				}
			}
		}
		groups[candidate.Path] = append(groups[candidate.Path], candidate)
	}

	for _, group := range groups {
		if len(group) <= 1 {
			continue
		}
		accepted := 0
		active := make([]FieldCandidate, 0, len(group))
		for _, candidate := range group {
			switch candidate.Review {
			case ReviewRejected:
				continue
			case ReviewAccepted:
				accepted++
			}
			active = append(active, candidate)
		}
		if accepted > 1 || (accepted == 1 && len(active) > 1) {
			return validationError(CodeInvalidReview, "candidate.review")
		}
		if accepted == 0 && len(active) > 1 {
			for _, candidate := range active {
				if !candidate.Ambiguous {
					return validationError(CodeInvalidCandidate, "candidate.ambiguous")
				}
			}
		}
	}
	return nil
}

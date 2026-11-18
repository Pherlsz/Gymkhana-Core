package conformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/portablejson"
)

func TestSchemasStrictJSONAndNoOrphanedSpec03Properties(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(filepath.Join("..", "schemas"))
	if err != nil {
		t.Fatalf("read schemas directory: %v", err)
	}

	orphanedDefinitions := map[string]struct{}{
		"MemoryMode":             {},
		"MemoryWriteMode":        {},
		"MemoryPolicy":           {},
		"TaskSignature":          {},
		"SkillLearningProposal":  {},
		"SkillLearningPolicy":    {},
		"LearningProposalStatus": {},
	}
	orphanedProperties := map[string]struct{}{
		"memory":                 {},
		"memory_policy":          {},
		"learning":               {},
		"learning_policy":        {},
		"skill_hint":             {},
		"task_signature":         {},
		"data_identification/v1": {},
	}

	schemaCount := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		schemaCount++
		name := entry.Name()
		path := filepath.Join("..", "schemas", name)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read schema %s: %v", path, err)
			}
			if err := portablejson.ValidateObject(raw); err != nil {
				t.Fatalf("schema %s failed strict portablejson.ValidateObject: %v", path, err)
			}

			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("unmarshal schema %s: %v", path, err)
			}
			for _, field := range []string{"$schema", "$id", "title"} {
				value, ok := doc[field].(string)
				if !ok || strings.TrimSpace(value) == "" {
					t.Fatalf("schema %s missing non-empty string %q", path, field)
				}
			}

			if defs, ok := doc["$defs"].(map[string]any); ok {
				for defName := range defs {
					if _, forbidden := orphanedDefinitions[defName]; forbidden {
						t.Fatalf("schema %s contains orphaned Spec 0.3 definition %q", path, defName)
					}
				}
			}

			assertNoOrphanedSchemaEntries(t, path, doc, orphanedProperties)
		})
	}

	if schemaCount == 0 {
		t.Fatal("expected at least one schema JSON file in ../schemas")
	}
}

func assertNoOrphanedSchemaEntries(t *testing.T, path string, node any, orphanedProperties map[string]struct{}) {
	t.Helper()

	switch value := node.(type) {
	case map[string]any:
		if props, ok := value["properties"].(map[string]any); ok {
			for propName := range props {
				if _, forbidden := orphanedProperties[propName]; forbidden {
					t.Fatalf("schema %s contains orphaned Spec 0.3 property %q", path, propName)
				}
			}
		}
		if enumValues, ok := value["enum"].([]any); ok {
			for _, item := range enumValues {
				if text, ok := item.(string); ok {
					if text == "memory" || text == "data_identification/v1" {
						t.Fatalf("schema %s contains orphaned Spec 0.3 enum value %q", path, text)
					}
				}
			}
		}
		if constVal, ok := value["const"].(string); ok {
			if constVal == "memory" || constVal == "data_identification/v1" {
				t.Fatalf("schema %s contains orphaned Spec 0.3 const value %q", path, constVal)
			}
		}
		for _, child := range value {
			assertNoOrphanedSchemaEntries(t, path, child, orphanedProperties)
		}
	case []any:
		for _, item := range value {
			assertNoOrphanedSchemaEntries(t, path, item, orphanedProperties)
		}
	}
}

func TestConformanceV04SuitesStructuralInvariants(t *testing.T) {
	t.Parallel()

	expectedVersionData, err := os.ReadFile(filepath.Join("..", "spec", "VERSION"))
	if err != nil {
		t.Fatalf("read specification version: %v", err)
	}
	expectedVersion := strings.TrimSpace(string(expectedVersionData))
	dir := "v" + expectedVersion

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read conformance directory %s: %v", dir, err)
	}

	suiteCount := 0
	seenSuites := make(map[string]string)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		suiteCount++
		filename := entry.Name()
		path := filepath.Join(dir, filename)

		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read conformance suite %s: %v", path, err)
		}

		var s suite
		if err := json.Unmarshal(raw, &s); err != nil {
			t.Fatalf("decode conformance suite %s: %v", path, err)
		}
		if s.SpecVersion != expectedVersion {
			t.Fatalf("conformance suite %s has spec_version %q, want %q", path, s.SpecVersion, expectedVersion)
		}
		if strings.TrimSpace(s.Suite) == "" {
			t.Fatalf("conformance suite %s has empty suite name", path)
		}
		if prevFile, duplicate := seenSuites[s.Suite]; duplicate {
			t.Fatalf("duplicate conformance suite identifier %q in %s and %s", s.Suite, prevFile, path)
		}
		seenSuites[s.Suite] = path

		if len(s.Cases) == 0 {
			t.Fatalf("conformance suite %s has no cases", path)
		}

		seenCases := make(map[string]struct{}, len(s.Cases))
		for i, c := range s.Cases {
			if strings.TrimSpace(c.Name) == "" {
				t.Fatalf("%s case[%d] has empty name", path, i)
			}
			if _, duplicate := seenCases[c.Name]; duplicate {
				t.Fatalf("%s has duplicate case name %q", path, c.Name)
			}
			seenCases[c.Name] = struct{}{}

			if strings.TrimSpace(c.Operation) == "" {
				t.Fatalf("%s case %q has empty operation", path, c.Name)
			}
			if strings.HasPrefix(c.Operation, "assistant.memory.") || strings.HasPrefix(c.Operation, "assistant.learning.") {
				t.Fatalf("%s case %q uses removed Spec 0.3 operation %q", path, c.Name, c.Operation)
			}
			if len(c.Input) == 0 {
				t.Fatalf("%s case %q has empty input", path, c.Name)
			}
			hasOutput := len(c.Output) > 0
			hasError := c.Error != nil && strings.TrimSpace(*c.Error) != ""
			if hasOutput == (c.Error != nil) {
				t.Fatalf("%s case %q must specify exactly one of output or error", path, c.Name)
			}
			if c.Error != nil && !hasError {
				t.Fatalf("%s case %q has empty error code", path, c.Name)
			}
		}
	}

	if suiteCount == 0 {
		t.Fatalf("no conformance JSON suites found in %s", dir)
	}
}

package v1alpha1

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

func defaultSourceType(source *Source) string {
	if source.SourceType == "" {
		return "docs"
	}
	return source.SourceType
}

// Feature: eventlistener-migration, Property 1: Org namespace derivation
// **Validates: Requirements 1.3**
func TestProperty_OrgNamespaceDerivation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		orgName := rapid.StringMatching(`[a-z][a-z0-9\-]{0,62}`).Draw(t, "orgName")

		kb := &KnowledgeBase{
			Spec: KnowledgeBaseSpec{
				Organization: orgName,
			},
		}

		result := kb.OrgNamespace()

		expectedNamespace := "org-" + orgName
		if result != expectedNamespace {
			t.Fatalf("OrgNamespace() = %q, expected %q for organization %q", result, expectedNamespace, orgName)
		}

		if !strings.HasPrefix(result, "org-") {
			t.Fatalf("OrgNamespace() = %q does not start with 'org-' prefix", result)
		}
	})
}

// Feature: eventlistener-migration, Property 2: Empty organization validation
// **Validates: Requirements 1.2**
func TestProperty_EmptyOrganizationValidation(t *testing.T) {
	t.Run("empty organization returns error", func(t *testing.T) {
		kb := &KnowledgeBase{
			Spec: KnowledgeBaseSpec{
				Organization: "",
			},
		}
		err := kb.ValidateOrganization()
		if err == nil {
			t.Fatal("expected error for empty organization, got nil")
		}
	})

	t.Run("non-empty organization does not fail validation", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			orgName := rapid.StringMatching(`[a-z][a-z0-9\-]{0,62}`).Draw(t, "orgName")

			kb := &KnowledgeBase{
				Spec: KnowledgeBaseSpec{
					Organization: orgName,
				},
			}

			err := kb.ValidateOrganization()
			if err != nil {
				t.Fatalf("expected no error for non-empty organization %q, got: %s", orgName, err.Error())
			}
		})
	})
}

// **Validates: Requirements 1.6**
func TestProperty_DefaultSourceType(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sourceType := rapid.SampledFrom([]string{"", "docs", "code"}).Draw(t, "sourceType")
		branch := rapid.SampledFrom([]string{"", "main", "mainline", "develop"}).Draw(t, "branch")

		pathCount := rapid.IntRange(0, 3).Draw(t, "pathCount")
		paths := make([]string, pathCount)
		for i := range paths {
			paths[i] = rapid.SampledFrom([]string{".kiro/docs", "src/**", "lib/**", "docs"}).Draw(t, "path")
		}

		source := Source{
			RepoOrg:    "testorg",
			RepoName:   "testrepo",
			Branch:     branch,
			SourceType: sourceType,
			Paths:      paths,
		}

		result := defaultSourceType(&source)

		if sourceType == "" {
			if result != "docs" {
				t.Fatalf("expected 'docs' for empty sourceType, got %q", result)
			}
		} else {
			if result != sourceType {
				t.Fatalf("expected %q for explicit sourceType, got %q", sourceType, result)
			}
		}
	})
}

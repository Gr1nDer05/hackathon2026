package docs

import (
	"testing"

	"github.com/goccy/go-yaml"
)

func TestEmbeddedOpenAPIYAMLParses(t *testing.T) {
	t.Helper()

	content, err := embeddedFiles.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("failed to read embedded openapi.yaml: %v", err)
	}

	var document map[string]any
	if err := yaml.Unmarshal(content, &document); err != nil {
		t.Fatalf("failed to parse openapi.yaml: %v", err)
	}

	if document["openapi"] == nil {
		t.Fatalf("expected openapi version field")
	}

	if document["paths"] == nil {
		t.Fatalf("expected paths section")
	}
}

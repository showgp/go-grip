package internal

import "testing"

func TestRenderCollectsTOCEntries(t *testing.T) {
	t.Parallel()

	rendered, err := NewParser().Render([]byte("# Intro\n\n## Setup\n\n### Details\n"))
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}

	want := []TOCEntry{
		{Level: 1, Text: "Intro", ID: "intro"},
		{Level: 2, Text: "Setup", ID: "setup"},
		{Level: 3, Text: "Details", ID: "details"},
	}
	if len(rendered.TOC) != len(want) {
		t.Fatalf("expected %d TOC entries, got %d: %#v", len(want), len(rendered.TOC), rendered.TOC)
	}
	for i := range want {
		if rendered.TOC[i] != want[i] {
			t.Fatalf("entry %d: expected %#v, got %#v", i, want[i], rendered.TOC[i])
		}
	}
}

func TestRenderTOCUsesGeneratedDuplicateHeadingIDs(t *testing.T) {
	t.Parallel()

	rendered, err := NewParser().Render([]byte("# Intro\n\n# Intro\n"))
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}

	wantIDs := []string{"intro", "intro-1"}
	if len(rendered.TOC) != len(wantIDs) {
		t.Fatalf("expected %d TOC entries, got %d: %#v", len(wantIDs), len(rendered.TOC), rendered.TOC)
	}
	for i, want := range wantIDs {
		if rendered.TOC[i].ID != want {
			t.Fatalf("entry %d: expected ID %q, got %q", i, want, rendered.TOC[i].ID)
		}
	}
}

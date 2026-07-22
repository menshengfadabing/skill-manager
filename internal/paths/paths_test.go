package paths_test

import (
	"path/filepath"
	"testing"

	"github.com/no-story/skill-manager/internal/paths"
)

func TestResolveAlwaysGlobal(t *testing.T) {
	pFalse, err := paths.Resolve(t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	pTrue, err := paths.Resolve(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	if pFalse.Scope != "global" || pTrue.Scope != "global" {
		t.Fatalf("expected global-only, got %q / %q", pFalse.Scope, pTrue.Scope)
	}
}

func TestGlobalWorkDirsCoverSupportedTools(t *testing.T) {
	p, err := paths.Resolve(t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"agents": true, "claude": true, "cursor": true, "codex": true, "qwen": true, "pi": true}
	for _, w := range p.WorkDirs {
		delete(want, w.ID)
		if w.Path == "" {
			t.Fatalf("empty path for %s", w.ID)
		}
	}
	for id := range want {
		t.Fatalf("missing work dir id %s", id)
	}
	// must not manage skills-cursor
	for _, w := range p.WorkDirs {
		if filepath.Base(w.Path) == "skills-cursor" {
			t.Fatal("must not manage skills-cursor")
		}
	}
	actives := p.ActiveTargets()
	if len(actives) < 6 {
		t.Fatalf("expected >=6 actives, got %d", len(actives))
	}
}

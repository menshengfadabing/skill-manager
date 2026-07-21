package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProjectUsesSharedWarehouse(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SKILL_MANAGER_HOME", filepath.Join(home, ".agents"))

	root := filepath.Join(home, "repo")
	_ = os.MkdirAll(filepath.Join(root, ".git"), 0o755)

	p, err := Resolve(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if p.Scope != "project" {
		t.Fatalf("scope=%s", p.Scope)
	}
	wantWH := filepath.Join(home, ".agents", "skills-all")
	if p.Warehouse != wantWH {
		t.Fatalf("warehouse=%s want %s", p.Warehouse, wantWH)
	}
	if p.AgentsActive != filepath.Join(root, ".agents", "skills") {
		t.Fatalf("agents=%s", p.AgentsActive)
	}
	if p.ProfilesDir != filepath.Join(root, ".agents", "profiles") {
		t.Fatalf("profiles=%s", p.ProfilesDir)
	}
}

func TestResolveGlobal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SKILL_MANAGER_HOME", filepath.Join(home, ".agents"))
	t.Setenv("SKILL_MANAGER_CLAUDE", filepath.Join(home, ".claude", "skills"))

	p, err := Resolve(home, true)
	if err != nil {
		t.Fatal(err)
	}
	if p.Scope != "global" {
		t.Fatalf("scope=%s", p.Scope)
	}
	if p.Warehouse != filepath.Join(home, ".agents", "skills-all") {
		t.Fatalf("warehouse=%s", p.Warehouse)
	}
	if len(p.ActiveTargets()) != 2 {
		t.Fatalf("targets=%d", len(p.ActiveTargets()))
	}
}

func TestResolveNoGit(t *testing.T) {
	dir := t.TempDir()
	_, err := Resolve(dir, false)
	if err == nil {
		t.Fatal("expected error without git")
	}
}

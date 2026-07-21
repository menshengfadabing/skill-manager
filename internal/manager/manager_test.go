package manager_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/no-story/skill-manager/internal/manager"
	"github.com/no-story/skill-manager/internal/paths"
)

func TestEnableDisableDualLinks(t *testing.T) {
	root := t.TempDir()
	p := paths.Paths{
		Home:         root,
		AgentsHome:   filepath.Join(root, ".agents"),
		Warehouse:    filepath.Join(root, ".agents", "skills-all"),
		AgentsActive: filepath.Join(root, ".agents", "skills"),
		ClaudeActive: filepath.Join(root, ".claude", "skills"),
		ProfilesDir:  filepath.Join(root, ".agents", "profiles"),
		CurrentFile:  filepath.Join(root, ".agents", "profiles", ".current"),
	}
	mustMkSkill(t, filepath.Join(p.Warehouse, "foo"))
	m := manager.New(p)
	if err := m.Enable("foo"); err != nil {
		t.Fatal(err)
	}
	for _, active := range []string{p.AgentsActive, p.ClaudeActive} {
		link := filepath.Join(active, "foo")
		if !m.IsDirectWarehouseLink(link, "foo") {
			t.Fatalf("expected direct warehouse link at %s", link)
		}
	}
	if err := m.Disable("foo"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(p.AgentsActive, "foo")); !os.IsNotExist(err) {
		t.Fatal("expected removed")
	}
}

func TestSyncMigratesRealDirAndFixesTwoLevel(t *testing.T) {
	root := t.TempDir()
	p := paths.Paths{
		Home:         root,
		AgentsHome:   filepath.Join(root, ".agents"),
		Warehouse:    filepath.Join(root, ".agents", "skills-all"),
		AgentsActive: filepath.Join(root, ".agents", "skills"),
		ClaudeActive: filepath.Join(root, ".claude", "skills"),
		ProfilesDir:  filepath.Join(root, ".agents", "profiles"),
		CurrentFile:  filepath.Join(root, ".agents", "profiles", ".current"),
	}
	_ = os.MkdirAll(p.Warehouse, 0o755)
	_ = os.MkdirAll(p.AgentsActive, 0o755)
	_ = os.MkdirAll(p.ClaudeActive, 0o755)

	mustMkSkill(t, filepath.Join(p.AgentsActive, "bar"))
	// two-level: claude → agents/skills/bar
	if err := os.Symlink(filepath.Join(p.AgentsActive, "bar"), filepath.Join(p.ClaudeActive, "bar")); err != nil {
		t.Fatal(err)
	}

	m := manager.New(p)
	res, err := m.Sync()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(p.Warehouse, "bar", "SKILL.md")); err != nil {
		t.Fatalf("not in warehouse: %v", err)
	}
	if !m.IsDirectWarehouseLink(filepath.Join(p.AgentsActive, "bar"), "bar") {
		t.Fatal("agents not direct")
	}
	if !m.IsDirectWarehouseLink(filepath.Join(p.ClaudeActive, "bar"), "bar") {
		t.Fatal("claude not direct")
	}
	if m.IsTwoLevelLink(filepath.Join(p.ClaudeActive, "bar")) {
		t.Fatal("still two-level")
	}
	_ = res
}

func TestDoctorDetectsTwoLevel(t *testing.T) {
	root := t.TempDir()
	p := paths.Paths{
		AgentsHome:   filepath.Join(root, ".agents"),
		Warehouse:    filepath.Join(root, ".agents", "skills-all"),
		AgentsActive: filepath.Join(root, ".agents", "skills"),
		ClaudeActive: filepath.Join(root, ".claude", "skills"),
		ProfilesDir:  filepath.Join(root, ".agents", "profiles"),
		CurrentFile:  filepath.Join(root, ".agents", "profiles", ".current"),
	}
	mustMkSkill(t, filepath.Join(p.AgentsActive, "x"))
	_ = os.MkdirAll(p.ClaudeActive, 0o755)
	_ = os.MkdirAll(p.Warehouse, 0o755)
	_ = os.Symlink(filepath.Join(p.AgentsActive, "x"), filepath.Join(p.ClaudeActive, "x"))

	m := manager.New(p)
	issues, err := m.Doctor()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, iss := range issues {
		if iss.Level == "error" && (contains(iss.Message, "two-level") || contains(iss.Message, "real skill")) {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected two-level or real-dir issue, got %#v", issues)
	}
}

func mustMkSkill(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: test\ndescription: t\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}

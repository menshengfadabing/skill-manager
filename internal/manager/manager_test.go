package manager_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/no-story/skill-manager/internal/manager"
	"github.com/no-story/skill-manager/internal/paths"
)

func testPaths(t *testing.T, root string) paths.Paths {
	t.Helper()
	agents := filepath.Join(root, "project-agents")
	profiles := filepath.Join(agents, "profiles")
	wh := filepath.Join(root, "global-agents", "skills-all")
	return paths.Paths{
		Scope:        "project",
		Root:         root,
		Home:         root,
		AgentsHome:   agents,
		Warehouse:    wh, // shared warehouse (simulates ~/.agents/skills-all)
		AgentsActive: filepath.Join(agents, "skills"),
		ClaudeActive: filepath.Join(root, ".claude", "skills"),
		ProfilesDir:  profiles,
		CurrentFile:  filepath.Join(profiles, ".current"),
		BackupsDir:   filepath.Join(agents, "backups"),
	}
}

func TestApplySetDualLinks(t *testing.T) {
	root := t.TempDir()
	p := testPaths(t, root)
	mustMkSkill(t, filepath.Join(p.Warehouse, "foo"))
	m := manager.New(p)
	if err := m.ApplySet([]string{"foo"}); err != nil {
		t.Fatal(err)
	}
	for _, active := range p.ActiveTargets() {
		link := filepath.Join(active, "foo")
		if !m.IsDirectWarehouseLink(link, "foo") {
			t.Fatalf("expected direct warehouse link at %s", link)
		}
	}
}

func TestSyncMigratesIntoSharedWarehouse(t *testing.T) {
	root := t.TempDir()
	p := testPaths(t, root)
	_ = os.MkdirAll(p.Warehouse, 0o755)
	_ = os.MkdirAll(p.AgentsActive, 0o755)
	_ = os.MkdirAll(p.ClaudeActive, 0o755)

	mustMkSkill(t, filepath.Join(p.AgentsActive, "bar"))
	if err := os.Symlink(filepath.Join(p.AgentsActive, "bar"), filepath.Join(p.ClaudeActive, "bar")); err != nil {
		t.Fatal(err)
	}

	m := manager.New(p)
	res, err := m.Sync(manager.SyncOptions{Yes: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.BackupDir == "" {
		t.Fatal("expected backup")
	}
	if _, err := os.Stat(filepath.Join(p.Warehouse, "bar", "SKILL.md")); err != nil {
		t.Fatalf("not in shared warehouse: %v", err)
	}
	if !m.IsDirectWarehouseLink(filepath.Join(p.AgentsActive, "bar"), "bar") {
		t.Fatal("agents not direct")
	}
	if m.IsTwoLevelLink(filepath.Join(p.ClaudeActive, "bar")) {
		t.Fatal("still two-level")
	}
}

func TestSyncDryRunNoWrite(t *testing.T) {
	root := t.TempDir()
	p := testPaths(t, root)
	_ = os.MkdirAll(p.AgentsActive, 0o755)
	mustMkSkill(t, filepath.Join(p.AgentsActive, "x"))
	m := manager.New(p)
	res, err := m.Sync(manager.SyncOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.WouldMigrate) == 0 {
		t.Fatal("expected would migrate")
	}
	if _, err := os.Stat(filepath.Join(p.Warehouse, "x")); !os.IsNotExist(err) {
		t.Fatal("dry-run should not migrate")
	}
}

func TestYAMLProfileCreateDelete(t *testing.T) {
	root := t.TempDir()
	p := testPaths(t, root)
	m := manager.New(p)
	if err := m.CreateProfile("lean", "test"); err != nil {
		t.Fatal(err)
	}
	if err := m.SetCurrentProfile("lean"); err != nil {
		t.Fatal(err)
	}
	if err := m.DeleteProfile("lean", false); err == nil {
		t.Fatal("should refuse delete current")
	}
	if err := m.DeleteProfile("lean", true); err != nil {
		t.Fatal(err)
	}
}

func TestProjectSyncDoesNotTouchGlobalActivity(t *testing.T) {
	root := t.TempDir()
	p := testPaths(t, root)
	globalActive := filepath.Join(root, "global-agents", "skills")
	_ = os.MkdirAll(globalActive, 0o755)
	_ = os.WriteFile(filepath.Join(globalActive, "KEEP"), []byte("1"), 0o644)

	mustMkSkill(t, filepath.Join(p.AgentsActive, "only-from-project"))
	m := manager.New(p)
	if _, err := m.Sync(manager.SyncOptions{Yes: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(globalActive, "KEEP")); err != nil {
		t.Fatal("global activity should be untouched")
	}
	if _, err := os.Stat(filepath.Join(p.Warehouse, "only-from-project")); err != nil {
		t.Fatal("should land in shared warehouse")
	}
}

func TestDoctorDetectsTwoLevel(t *testing.T) {
	root := t.TempDir()
	p := testPaths(t, root)
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
		if iss.Level == "error" && (strings.Contains(iss.Message, "two-level") || strings.Contains(iss.Message, "real skill")) {
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

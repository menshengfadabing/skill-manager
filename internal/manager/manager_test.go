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
	agents := filepath.Join(root, ".agents")
	profiles := filepath.Join(agents, "profiles")
	agentsActive := filepath.Join(agents, "skills")
	claude := filepath.Join(root, ".claude", "skills")
	cursor := filepath.Join(root, ".cursor", "skills")
	codex := filepath.Join(root, ".codex", "skills")
	qwen := filepath.Join(root, ".qwen", "skills")
	pi := filepath.Join(root, ".pi", "agent", "skills")
	return paths.Paths{
		Scope:        "global",
		Root:         root,
		Home:         root,
		AgentsHome:   agents,
		Warehouse:    filepath.Join(agents, "skills-all"),
		AgentsActive: agentsActive,
		ClaudeActive: claude,
		WorkDirs: []paths.WorkDir{
			{ID: "agents", Path: agentsActive},
			{ID: "claude", Path: claude},
			{ID: "cursor", Path: cursor},
			{ID: "codex", Path: codex},
			{ID: "qwen", Path: qwen},
			{ID: "pi", Path: pi},
		},
		ProfilesDir: profiles,
		CurrentFile: filepath.Join(profiles, ".current"),
		BackupsDir:  filepath.Join(agents, "backups"),
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

func TestExtraIngestMigratesButDoesNotEnable(t *testing.T) {
	root := t.TempDir()
	p := testPaths(t, root)
	projectAgents := filepath.Join(root, "repo", ".agents", "skills")
	p.ExtraIngest = []string{projectAgents}

	_ = os.MkdirAll(p.AgentsActive, 0o755)
	_ = os.MkdirAll(p.Warehouse, 0o755)
	mustMkSkill(t, filepath.Join(projectAgents, "only-from-project"))

	m := manager.New(p)
	if _, err := m.Sync(manager.SyncOptions{Yes: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(p.Warehouse, "only-from-project")); err != nil {
		t.Fatal("should land in shared warehouse")
	}
	if _, err := os.Lstat(filepath.Join(p.AgentsActive, "only-from-project")); !os.IsNotExist(err) {
		t.Fatal("extra ingest must not enable on global work dirs")
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

func TestInitialSnapshotAndRestore(t *testing.T) {
	root := t.TempDir()
	p := testPaths(t, root)
	_ = os.MkdirAll(p.Warehouse, 0o755)
	_ = os.MkdirAll(p.AgentsActive, 0o755)
	_ = os.MkdirAll(p.ClaudeActive, 0o755)
	mustMkSkill(t, filepath.Join(p.AgentsActive, "legacy"))
	m := manager.New(p)

	initPath, err := m.EnsureInitialSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(initPath) != manager.InitialSnapshotID {
		t.Fatalf("got %s", initPath)
	}
	if _, err := m.EnsureInitialSnapshot(); err != nil {
		t.Fatal(err)
	}

	// mutate working dir
	_ = os.RemoveAll(filepath.Join(p.AgentsActive, "legacy"))
	mustMkSkill(t, filepath.Join(p.AgentsActive, "other"))

	if err := m.RestoreSnapshot("initial"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(p.AgentsActive, "legacy")); err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(p.AgentsActive, "other")); !os.IsNotExist(err) {
		t.Fatal("expected other removed after restore")
	}
	snaps, err := m.ListSnapshots()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range snaps {
		if s.IsInitial {
			found = true
		}
	}
	if !found {
		t.Fatal("expected INITIAL in log")
	}
}

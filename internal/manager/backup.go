package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupSnapshot copies mutable scope state into AgentsHome/backups/<ts>/.
func (m *Manager) BackupSnapshot(label string) (string, error) {
	if err := os.MkdirAll(m.P.BackupsDir, 0o755); err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102-150405")
	if label != "" {
		ts = ts + "-" + label
	}
	dst := filepath.Join(m.P.BackupsDir, ts)
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return "", err
	}

	// Manifest of paths + copy profiles + list enabled names (not full warehouse bodies)
	manifest := fmt.Sprintf("scope=%s\nroot=%s\nwarehouse=%s\n", m.P.Scope, m.P.Root, m.P.Warehouse)
	_ = os.WriteFile(filepath.Join(dst, "MANIFEST.txt"), []byte(manifest+m.P.Summary()+"\n"), 0o644)

	if err := copyIfExists(m.P.ProfilesDir, filepath.Join(dst, "profiles")); err != nil {
		return dst, err
	}
	for _, name := range []string{"agents", "claude"} {
		var src string
		switch name {
		case "agents":
			src = m.P.AgentsActive
		case "claude":
			src = m.P.ClaudeActive
		}
		list, _ := m.EnabledOn(src)
		var b string
		for _, n := range list {
			b += n + "\n"
		}
		_ = os.WriteFile(filepath.Join(dst, name+"-enabled.txt"), []byte(b), 0o644)
		// copy symlink forest as-is for restore reference
		_ = copyIfExists(src, filepath.Join(dst, name+"-skills"))
	}

	// warehouse name list only (bodies stay; conflict bak handles skill bodies)
	names, _ := m.SkillNames()
	var wb string
	for _, n := range names {
		wb += n + "\n"
	}
	_ = os.WriteFile(filepath.Join(dst, "warehouse-names.txt"), []byte(wb), 0o644)
	return dst, nil
}

func copyIfExists(src, dst string) error {
	st, err := os.Lstat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if st.Mode()&os.ModeSymlink != 0 || !st.IsDir() {
		return CopyDir(src, dst) // best effort
	}
	return CopyDir(src, dst)
}

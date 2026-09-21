package ideconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMarkerDirDetectsFreshInstall is the regression for auto-install
// skipping Codex CLI and Cursor on a fresh machine: both create their state
// directory on first run, but config.toml / mcp.json only after the user
// adds a setting or MCP server, and Detect used to require that file.
func TestMarkerDirDetectsFreshInstall(t *testing.T) {
	cases := []struct {
		name   string
		writer Writer
		marker string
		target string
	}{
		{"codex", &codexWriter{}, ".codex", filepath.Join(".codex", "config.toml")},
		{"cursor", &cursorWriter{}, ".cursor", filepath.Join(".cursor", "mcp.json")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			w := tc.writer
			target := filepath.Join(home, tc.target)

			if WriterFound(w) {
				t.Fatalf("detected without any state; Detect() = %q", w.Detect())
			}
			if err := os.MkdirAll(filepath.Join(home, tc.marker), 0o755); err != nil {
				t.Fatal(err)
			}
			if !WriterFound(w) {
				t.Fatalf("not detected with ~/%s present; Detect() = %q", tc.marker, w.Detect())
			}

			res := w.Write(DefaultSpec("/opt/a2abridge"), false)
			if res.Error != nil {
				t.Fatalf("Write: %v", res.Error)
			}
			if !res.Updated || res.Path != target {
				t.Fatalf("Write result = %+v, want Updated at %s", res, target)
			}
			b, err := os.ReadFile(target)
			if err != nil {
				t.Fatalf("config not written: %v", err)
			}
			if !strings.Contains(string(b), "/opt/a2abridge") {
				t.Fatalf("config lacks the a2a entry:\n%s", b)
			}
			if w.Detect() != target {
				t.Errorf("Detect() after write = %q, want %s", w.Detect(), target)
			}

			// Uninstall with the marker path Detect returned before the write
			// must still clean the real config file.
			if err := RemoveMCPEntry(w, filepath.Join(home, tc.marker)); err != nil {
				t.Fatalf("RemoveMCPEntry: %v", err)
			}
			b, err = os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(b), "/opt/a2abridge") {
				t.Errorf("a2a entry still present after uninstall:\n%s", b)
			}
		})
	}
}

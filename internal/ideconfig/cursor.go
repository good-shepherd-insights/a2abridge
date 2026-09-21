package ideconfig

import (
	"path/filepath"
)

// cursorWriter handles Cursor's MCP config at ~/.cursor/mcp.json.
//
// Schema is identical to Claude Code's user-level mcpServers block.
type cursorWriter struct{}

func (cursorWriter) Name() string { return "Cursor" }

// Detect answers "is Cursor installed?". Cursor creates ~/.cursor (argv.json,
// extensions) on first launch but mcp.json only when the first MCP server is
// added, so the directory also counts as a marker.
func (w cursorWriter) Detect() string {
	if target := w.writeTarget(); fileExists(target) {
		return target
	}
	h, err := homeDir()
	if err != nil {
		return ""
	}
	if marker := filepath.Join(h, ".cursor"); dirExists(marker) {
		return marker
	}
	return ""
}

// writeTarget is the global MCP config Cursor reads.
func (cursorWriter) writeTarget() string {
	h, err := homeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(h, ".cursor", "mcp.json")
}

func (w cursorWriter) Write(spec Spec, dryRun bool) Result {
	return writeJSONConfig(w.Name(), w.writeTarget(), dryRun, func(root map[string]any) bool {
		return setMCPServerEntry(root, spec)
	})
}

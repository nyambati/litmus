package workspace

import (
	"fmt"
)

// LoadAssembledWorkspace loads and assembles a workspace from config.
func Load(dir string) (*Workspace, error) {
	ws := New(dir)
	if _, err := ws.Assemble(); err != nil {
		return nil, fmt.Errorf("assembling workspace: %w", err)
	}
	return ws, nil
}
